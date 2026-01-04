package scaffold

import "fmt"

func amqpConsumerTemplate(name, namePascal, moduleImportPath string) string {
	return fmt.Sprintf(`package %s

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	service *Service
}

func NewConsumer(conn *amqp.Connection, service *Service) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return &Consumer{
		conn:    conn,
		channel: ch,
		service: service,
	}, nil
}

func (c *Consumer) Close() error {
	return c.channel.Close()
}

func (c *Consumer) Consume(ctx context.Context, queueName string) error {
	msgs, err := c.channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-msgs:
			if err := c.handleMessage(ctx, msg); err != nil {
				log.Printf("error handling message: %%v", err)
				msg.Nack(false, true)
				continue
			}
			msg.Ack(false)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg amqp.Delivery) error {
	switch msg.Type {
	case "%s.create":
		return c.handleCreate(ctx, msg.Body)
	case "%s.update":
		return c.handleUpdate(ctx, msg.Body)
	case "%s.delete":
		return c.handleDelete(ctx, msg.Body)
	default:
		log.Printf("unknown message type: %%s", msg.Type)
		return nil
	}
}

func (c *Consumer) handleCreate(ctx context.Context, body []byte) error {
	var entity %s
	if err := json.Unmarshal(body, &entity); err != nil {
		return err
	}
	return c.service.Create(ctx, &entity)
}

func (c *Consumer) handleUpdate(ctx context.Context, body []byte) error {
	var entity %s
	if err := json.Unmarshal(body, &entity); err != nil {
		return err
	}
	return c.service.Update(ctx, &entity)
}

type deleteMessage struct {
	ID string `+"`json:\"id\"`"+`
}

func (c *Consumer) handleDelete(ctx context.Context, body []byte) error {
	var msg deleteMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return err
	}
	id, err := parseUUID(msg.ID)
	if err != nil {
		return err
	}
	return c.service.Delete(ctx, id)
}

func parseUUID(s string) (uuid, error) {
	// Import github.com/google/uuid and use uuid.Parse
	return uuid{}, nil
}

type uuid struct{}
`, name, name, name, name, namePascal, namePascal)
}
