package mq

import (
	"fmt"
	"strconv"
	"time"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

const (
	// DLQ/Retry settings
	DefaultMaxRetries   = 5
	RetryExchangeSuffix = ".retry.exchange"
	DLQSuffix           = ".dlq"
	RetrySuffix         = ".retry"

	// Stepped delays (exponential backoff): 1s, 5s, 30s, 2m, 10m
	retryDelay1 = 1 * time.Second
	retryDelay2 = 5 * time.Second
	retryDelay3 = 30 * time.Second
	retryDelay4 = 2 * time.Minute
	retryDelay5 = 10 * time.Minute

	// Header keys
	HeaderRetryCount = "x-retry-count"
)

var steppedDelays = []time.Duration{retryDelay1, retryDelay2, retryDelay3, retryDelay4, retryDelay5}

// QueueWithDLQ holds queue names for main, DLQ, and retry queues
type QueueWithDLQ struct {
	Main         string
	DLQ          string
	Retry        string
	RetryExhange string
}

// buildQueueNames generates queue names from a prefix
func buildQueueNames(prefix string) QueueWithDLQ {
	return QueueWithDLQ{
		Main:         prefix,
		DLQ:          prefix + DLQSuffix,
		Retry:        prefix + RetrySuffix,
		RetryExhange: prefix + RetryExchangeSuffix,
	}
}

// setupQueueWithDLQ creates main queue, DLQ, and retry queue with proper bindings
// queuePrefix: e.g., "email.verification" or "password.reset"
func (r *RabbitMQ) setupQueueWithDLQ(queuePrefix string) error {
	queues := buildQueueNames(queuePrefix)

	// 1. Declare retry exchange (direct) for routing back to main queue
	err := r.channel.ExchangeDeclare(
		queues.RetryExhange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare retry exchange: %w", err)
	}

	// 2. Declare main queue with DLQ routing
	_, err = r.channel.QueueDeclare(
		queues.Main,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "",         // default exchange
			"x-dead-letter-routing-key": queues.DLQ, // route to DLQ
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare main queue: %w", err)
	}

	// 3. Declare DLQ (no dead-letter, messages stay here if max retries exceeded)
	_, err = r.channel.QueueDeclare(
		queues.DLQ,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// 4. Declare retry queue with dead-letter back to main queue
	_, err = r.channel.QueueDeclare(
		queues.Retry,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "",          // default exchange
			"x-dead-letter-routing-key": queues.Main, // route back to main
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare retry queue: %w", err)
	}

	r.logger.Info("Setup queue with DLQ",
		zap.String("main", queues.Main),
		zap.String("dlq", queues.DLQ),
		zap.String("retry", queues.Retry),
	)

	return nil
}

// consumeDLQWithRetry consumes from DLQ and retries with stepped delays
func (r *RabbitMQ) consumeDLQWithRetry(queues QueueWithDLQ, maxRetries int) {
	msgs, err := r.channel.Consume(queues.DLQ, "", false, false, false, false, nil)
	if err != nil {
		r.logger.Error("Failed to consume from DLQ", zap.String("dlq", queues.DLQ), zap.Error(err))
		return
	}

	r.logger.Info("Started DLQ retry consumer", zap.String("dlq", queues.DLQ))

	for msg := range msgs {
		retryCount := getRetryCountFromXDeath(msg)

		if retryCount >= maxRetries {
			// Max retries exceeded, leave in DLQ (or move to permanent DLQ)
			r.logger.Warn("Message exceeded max retries, leaving in DLQ",
				zap.String("dlq", queues.DLQ),
				zap.Int("retry_count", retryCount),
				zap.Int("max_retries", maxRetries),
				zap.String("correlation_id", msg.CorrelationId),
			)
			// Ack to remove from DLQ - it's been logged for manual intervention
			msg.Ack(false)
			continue
		}

		// Calculate stepped delay
		delay := getSteppedDelay(retryCount)

		// Publish to retry queue with TTL
		err := r.publishToRetryQueue(queues.Retry, msg, retryCount+1, delay)
		if err != nil {
			r.logger.Error("Failed to publish to retry queue",
				zap.Error(err),
				zap.String("retry_queue", queues.Retry),
			)
			// Nack and requeue back to DLQ for another attempt
			msg.Nack(false, true)
			continue
		}

		r.logger.Info("Message scheduled for retry",
			zap.String("retry_queue", queues.Retry),
			zap.Int("retry_count", retryCount+1),
			zap.Duration("delay", delay),
			zap.String("correlation_id", msg.CorrelationId),
		)
		msg.Ack(false)
	}
}

// publishToRetryQueue publishes message to retry queue with TTL for stepped delay
func (r *RabbitMQ) publishToRetryQueue(retryQueue string, originalMsg amqp.Delivery, retryCount int, delay time.Duration) error {
	// Copy headers and add retry count
	headers := amqp.Table{}
	if originalMsg.Headers != nil {
		for k, v := range originalMsg.Headers {
			headers[k] = v
		}
	}
	headers[HeaderRetryCount] = retryCount

	// Preserve original routing key for traceability
	if originalMsg.RoutingKey != "" {
		headers["x-original-routing-key"] = originalMsg.RoutingKey
	}

	return r.channel.Publish(
		"",         // default exchange
		retryQueue, // routing key = queue name
		false,
		false,
		amqp.Publishing{
			ContentType:   originalMsg.ContentType,
			Body:          originalMsg.Body,
			Headers:       headers,
			CorrelationId: originalMsg.CorrelationId,
			Expiration:    strconv.FormatInt(delay.Milliseconds(), 10), // TTL in ms
			DeliveryMode:  amqp.Persistent,
		},
	)
}

// getRetryCountFromXDeath extracts retry count from x-death header (RabbitMQ native)
func getRetryCountFromXDeath(msg amqp.Delivery) int {
	// First check our custom header
	if msg.Headers != nil {
		if count, ok := msg.Headers[HeaderRetryCount].(int); ok {
			return count
		}
		if count, ok := msg.Headers[HeaderRetryCount].(int32); ok {
			return int(count)
		}
		if count, ok := msg.Headers[HeaderRetryCount].(int64); ok {
			return int(count)
		}
	}

	// Fallback to x-death header (native RabbitMQ)
	if msg.Headers == nil {
		return 0
	}

	xDeath, ok := msg.Headers["x-death"].([]interface{})
	if !ok || len(xDeath) == 0 {
		return 0
	}

	// Sum up all death counts
	totalCount := 0
	for _, death := range xDeath {
		if deathMap, ok := death.(amqp.Table); ok {
			if count, ok := deathMap["count"].(int64); ok {
				totalCount += int(count)
			}
		}
	}

	return totalCount
}

// getRetryCount extracts retry count from message (simple version for main queue)
func getRetryCount(msg amqp.Delivery) int {
	if msg.Headers == nil {
		return 0
	}
	if count, ok := msg.Headers[HeaderRetryCount].(int); ok {
		return count
	}
	if count, ok := msg.Headers[HeaderRetryCount].(int32); ok {
		return int(count)
	}
	if count, ok := msg.Headers[HeaderRetryCount].(int64); ok {
		return int(count)
	}
	return 0
}

// getSteppedDelay returns delay based on retry count (exponential backoff)
func getSteppedDelay(retryCount int) time.Duration {
	if retryCount < 0 {
		retryCount = 0
	}
	if retryCount >= len(steppedDelays) {
		return steppedDelays[len(steppedDelays)-1]
	}
	return steppedDelays[retryCount]
}
