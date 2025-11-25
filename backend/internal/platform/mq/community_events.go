package mq

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

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

