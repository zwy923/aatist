package mq

import (
	"context"
	"fmt"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// PublishEmailVerification publishes an email verification request to the queue
func (r *RabbitMQ) PublishEmailVerification(message interface{}) error {
	if err := r.publishJSON(context.Background(), "", "email.verification", message); err != nil {
		return err
	}
	r.logger.Info("Published email verification message", zap.String("queue", "email.verification"))
	return nil
}

// PublishPasswordReset publishes a password reset request to the queue
func (r *RabbitMQ) PublishPasswordReset(message interface{}) error {
	if err := r.publishJSON(context.Background(), "", "password.reset", message); err != nil {
		return err
	}
	r.logger.Info("Published password reset message", zap.String("queue", "password.reset"))
	return nil
}

// ConsumeEmailVerification consumes email verification messages with DLQ support
func (r *RabbitMQ) ConsumeEmailVerification(handler func([]byte) error) error {
	return r.consumeWithDLQ("email.verification", func(msg amqp.Delivery) error {
		return handler(msg.Body)
	}, "email verification")
}

// ConsumePasswordReset consumes password reset messages with DLQ support
func (r *RabbitMQ) ConsumePasswordReset(handler func([]byte) error) error {
	return r.consumeWithDLQ("password.reset", func(msg amqp.Delivery) error {
		return handler(msg.Body)
	}, "password reset")
}

// consumeWithDLQ is a generic consumer with DLQ support
func (r *RabbitMQ) consumeWithDLQ(queuePrefix string, handler func(amqp.Delivery) error, logName string) error {
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
			if err := handler(msg); err != nil {
				r.logger.Error(fmt.Sprintf("Failed to process %s message", logName),
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

	r.logger.Info(fmt.Sprintf("Started consuming %s messages with DLQ", logName),
		zap.String("queue", queues.Main),
		zap.String("dlq", queues.DLQ),
	)
	return nil
}

