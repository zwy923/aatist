package mq

import (
	"encoding/json"
	"fmt"

	"github.com/aalto-talent-network/backend/internal/platform/log"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

const (
	// Queue names
	EmailVerificationQueue = "email_verification"
)

// RabbitMQ wraps RabbitMQ connection
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *log.Logger
}

// NewRabbitMQ creates a new RabbitMQ connection
func NewRabbitMQ(brokerURL string, logger *log.Logger) (*RabbitMQ, error) {
	conn, err := amqp.Dial(brokerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare email verification queue
	_, err = channel.QueueDeclare(
		EmailVerificationQueue, // name
		true,                   // durable
		false,                  // delete when unused
		false,                  // exclusive
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
		logger:  logger,
	}, nil
}

// PublishEmailVerification publishes an email verification request to the queue
func (r *RabbitMQ) PublishEmailVerification(message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.channel.Publish(
		"",                     // exchange
		EmailVerificationQueue, // routing key
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	r.logger.Info("Published email verification message", zap.String("queue", EmailVerificationQueue))
	return nil
}

// ConsumeEmailVerification consumes email verification messages from the queue
func (r *RabbitMQ) ConsumeEmailVerification(handler func([]byte) error) error {
	msgs, err := r.channel.Consume(
		EmailVerificationQueue, // queue
		"",                     // consumer
		false,                  // auto-ack (we'll ack manually)
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		nil,                    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				r.logger.Error("Failed to process email verification message",
					zap.Error(err),
				)
				msg.Nack(false, true) // Requeue on error
			} else {
				msg.Ack(false)
			}
		}
	}()

	r.logger.Info("Started consuming email verification messages", zap.String("queue", EmailVerificationQueue))
	return nil
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
