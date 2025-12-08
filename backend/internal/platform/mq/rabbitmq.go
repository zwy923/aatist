package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aatist/backend/internal/platform/log"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

const (
	// Queue names
	EmailVerificationQueue = "email_verification"
	PasswordResetQueue     = "password_reset"

	// Exchanges
	CommunityEventsExchange = "community_events"

	defaultPublishConfirmTimeout = 5 * time.Second
	defaultPrefetchCount         = 50
)

// RabbitMQ wraps RabbitMQ connection
type RabbitMQ struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	confirmations  <-chan amqp.Confirmation
	returns        <-chan amqp.Return
	confirmTimeout time.Duration
	logger         *log.Logger
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

	// Setup password reset queue with DLQ
	if err := rmq.setupQueueWithDLQ("password.reset"); err != nil {
		rmq.Close()
		return nil, fmt.Errorf("failed to setup password reset DLQ: %w", err)
	}

	return rmq, nil
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

// publishJSON publishes a JSON message to the specified exchange/queue
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

// waitForConfirmation waits for broker acknowledgment
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

// logReturns logs returned messages from broker
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

// defaultServiceName returns the service name from env or default
func defaultServiceName() string {
	if svc := os.Getenv("SERVICE_NAME"); svc != "" {
		return svc
	}
	return "default"
}
