package queue

import (
	"context"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"example.com/gotube/internal/processing"
)

const (
	exchangeName = "gotube"
	queueName    = "video.uploaded.v1"
)

// Client publishes with confirms and consumes one uploaded event at a time.
type Client struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	confirms <-chan amqp.Confirmation
}

// Dial opens a confirming channel. The API does not call this.
func Dial(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("fila: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := ch.ExchangeDeclare(exchangeName, "direct", true, false, false, false, nil); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if _, err := ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := ch.QueueBind(queueName, queueName, exchangeName, false, nil); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := ch.Confirm(false); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := ch.Qos(1, 0, false); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &Client{conn: conn, ch: ch, confirms: ch.NotifyPublish(make(chan amqp.Confirmation, 8))}, nil
}

// Close releases the broker connection.
func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Publish waits for a broker confirm. A refusal is an error.
func (c *Client) Publish(ctx context.Context, eventType string, payload []byte) error {
	err := c.ch.PublishWithContext(ctx, exchangeName, eventType, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         payload,
	})
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case confirmation, ok := <-c.confirms:
		if !ok || !confirmation.Ack {
			return errors.New("a fila não confirmou a publicação")
		}
		return nil
	}
}

// Consume prefetches one message and acknowledges only the handler's durable outcome.
func (c *Client) Consume(ctx context.Context, handle func(context.Context, []byte) (processing.Outcome, error)) error {
	deliveries, err := c.ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return errors.New("a fila encerrou o consumo")
			}
			outcome, handleErr := handle(ctx, delivery.Body)
			if handleErr != nil || outcome == processing.OutcomeRetry {
				_ = delivery.Nack(false, true)
				continue
			}
			_ = delivery.Ack(false)
		}
	}
}
