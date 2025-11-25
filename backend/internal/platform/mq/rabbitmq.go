package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

const (
	// Queue names
	EmailVerificationQueue = "email_verification"

	// Exchanges
	CommunityEventsExchange = "community_events"

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

	defaultPublishConfirmTimeout = 5 * time.Second
	defaultPrefetchCount         = 50

	// Header keys
	HeaderRetryCount = "x-retry-count"
)

var steppedDelays = []time.Duration{retryDelay1, retryDelay2, retryDelay3, retryDelay4, retryDelay5}

// RabbitMQ wraps RabbitMQ connection
type RabbitMQ struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	confirmations  <-chan amqp.Confirmation
	returns        <-chan amqp.Return
	confirmTimeout time.Duration
	logger         *log.Logger
}

// QueueWithDLQ holds queue names for main, DLQ, and retry queues
type QueueWithDLQ struct {
	Main         string
	DLQ          string
	Retry        string
	RetryExhange string
}

// NewRabbitMQ creates a new RabbitMQ connection
func NewRabbitMQ(brokerURL string, confirmTimeout time.Duration, logger *log.Logger) (*RabbitMQ, error) {
	conn, err := amqp.Dial(brokerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	if err := channel.Confirm(false); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to enable publisher confirmations: %w", err)
	}
	confirmations := channel.NotifyPublish(make(chan amqp.Confirmation, 1024))
	returns := channel.NotifyReturn(make(chan amqp.Return, 1024))

	// Declare community events exchange (topic)
	err = channel.ExchangeDeclare(
		CommunityEventsExchange,
		"topic",
		true,
		false,
		false, // internal
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare community exchange: %w", err)
	}

	if confirmTimeout <= 0 {
		confirmTimeout = defaultPublishConfirmTimeout
	}

	rmq := &RabbitMQ{
		conn:           conn,
		channel:        channel,
		confirmations:  confirmations,
		returns:        returns,
		confirmTimeout: confirmTimeout,
		logger:         logger,
	}

	go rmq.logReturns()

	// Setup email verification queue with DLQ
	if err := rmq.setupQueueWithDLQ("email.verification"); err != nil {
		rmq.Close()
		return nil, fmt.Errorf("failed to setup email verification DLQ: %w", err)
	}

	return rmq, nil
}

// setupQueueWithDLQ creates main queue, DLQ, and retry queue with proper bindings
// queuePrefix: e.g., "email.verification" or "notification.community"
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

// SetupCommunityConsumerDLQ sets up DLQ infrastructure for a community event consumer
func (r *RabbitMQ) SetupCommunityConsumerDLQ(serviceName string) error {
	queuePrefix := fmt.Sprintf("%s.community", serviceName)
	return r.setupQueueWithDLQ(queuePrefix)
}

func buildQueueNames(prefix string) QueueWithDLQ {
	return QueueWithDLQ{
		Main:         prefix,
		DLQ:          prefix + DLQSuffix,
		Retry:        prefix + RetrySuffix,
		RetryExhange: prefix + RetryExchangeSuffix,
	}
}

// PublishEmailVerification publishes an email verification request to the queue
func (r *RabbitMQ) PublishEmailVerification(message interface{}) error {
	// Use the new queue name format
	if err := r.publishJSON(context.Background(), "", "email.verification", message); err != nil {
		return err
	}
	r.logger.Info("Published email verification message", zap.String("queue", "email.verification"))
	return nil
}

// PublishCommunityEvent publishes community domain events with publisher confirm.
func (r *RabbitMQ) PublishCommunityEvent(ctx context.Context, eventType string, payload interface{}) error {
	if r.channel == nil {
		return fmt.Errorf("channel is not initialized")
	}

	if err := r.publishJSON(ctx, CommunityEventsExchange, eventType, payload); err != nil {
		return err
	}
	r.logger.Info("Published community event", zap.String("event_type", eventType))
	return nil
}

// ConsumeEmailVerification consumes email verification messages with DLQ support
func (r *RabbitMQ) ConsumeEmailVerification(handler func([]byte) error) error {
	queuePrefix := "email.verification"
	queues := buildQueueNames(queuePrefix)

	if err := r.channel.Qos(defaultPrefetchCount, 0, false); err != nil {
		return fmt.Errorf("failed to set qos: %w", err)
	}

	// Consume from main queue
	msgs, err := r.channel.Consume(queues.Main, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Start DLQ retry consumer
	go r.consumeDLQWithRetry(queues, DefaultMaxRetries)

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				r.logger.Error("Failed to process email verification message",
					zap.Error(err),
					zap.Int("retry_count", getRetryCount(msg)),
				)
				// nack without requeue - goes to DLQ
				msg.Nack(false, false)
			} else {
				msg.Ack(false)
			}
		}
	}()

	r.logger.Info("Started consuming email verification messages with DLQ",
		zap.String("queue", queues.Main),
		zap.String("dlq", queues.DLQ),
	)
	return nil
}

// ConsumeCommunityEvents subscribes to community.* events with DLQ support
func (r *RabbitMQ) ConsumeCommunityEvents(serviceName string, bindingKeys []string, handler func(eventType string, payload []byte) error) error {
	if r.channel == nil {
		return errors.New("amqp channel is not initialized")
	}
	if serviceName == "" {
		serviceName = defaultServiceName()
	}
	if len(bindingKeys) == 0 {
		bindingKeys = []string{"community.post.*"}
	}

	queuePrefix := fmt.Sprintf("%s.community", serviceName)

	// Setup DLQ infrastructure first
	if err := r.setupQueueWithDLQ(queuePrefix); err != nil {
		return fmt.Errorf("failed to setup DLQ: %w", err)
	}

	queues := buildQueueNames(queuePrefix)

	// Bind main queue to community events exchange
	for _, key := range bindingKeys {
		if err := r.channel.QueueBind(queues.Main, key, CommunityEventsExchange, false, nil); err != nil {
			return fmt.Errorf("failed to bind community queue: %w", err)
		}
	}

	if err := r.channel.Qos(defaultPrefetchCount, 0, false); err != nil {
		return fmt.Errorf("failed to set qos: %w", err)
	}

	msgs, err := r.channel.Consume(queues.Main, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to consume community events: %w", err)
	}

	// Start DLQ retry consumer
	go r.consumeDLQWithRetry(queues, DefaultMaxRetries)

	go func() {
		for msg := range msgs {
			if err := handler(msg.RoutingKey, msg.Body); err != nil {
				r.logger.Error("Failed to process community event",
					zap.String("event_type", msg.RoutingKey),
					zap.Error(err),
					zap.Int("retry_count", getRetryCount(msg)),
				)
				// nack without requeue - goes to DLQ
				msg.Nack(false, false)
			} else {
				msg.Ack(false)
			}
		}
	}()

	r.logger.Info("Started consuming community events with DLQ",
		zap.String("queue", queues.Main),
		zap.String("dlq", queues.DLQ),
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

	// Preserve original routing key for community events
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

// Close closes the RabbitMQ connection
func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func (r *RabbitMQ) publishJSON(ctx context.Context, exchange, routingKey string, payload interface{}) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if r.channel == nil {
		return errors.New("amqp channel is not initialized")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := r.channel.Publish(exchange, routingKey, true, false, amqp.Publishing{
		ContentType:   "application/json",
		Body:          body,
		CorrelationId: uuid.NewString(),
		DeliveryMode:  amqp.Persistent,
	}); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return r.waitForConfirmation(ctx)
}

func (r *RabbitMQ) waitForConfirmation(ctx context.Context) error {
	if r.confirmations == nil {
		return nil
	}

	select {
	case confirm, ok := <-r.confirmations:
		if !ok {
			return errors.New("confirmation channel closed")
		}
		if !confirm.Ack {
			return errors.New("message not acknowledged by broker")
		}
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(r.confirmTimeout):
		return errors.New("publish confirmation timeout")
	}
	return nil
}

func (r *RabbitMQ) logReturns() {
	if r.returns == nil {
		return
	}
	for ret := range r.returns {
		r.logger.Warn("Message returned by broker",
			zap.String("exchange", ret.Exchange),
			zap.String("routing_key", ret.RoutingKey),
			zap.Int("reply_code", int(ret.ReplyCode)),
			zap.String("reply_text", ret.ReplyText),
		)
	}
}

func defaultServiceName() string {
	if svc := os.Getenv("SERVICE_NAME"); svc != "" {
		return svc
	}
	return "default"
}
