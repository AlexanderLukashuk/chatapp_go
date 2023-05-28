package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// RabbitMQ represents the RabbitMQ message broker
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

// NewRabbitMQ creates a new instance of RabbitMQ
func NewRabbitMQ() (*RabbitMQ, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	queue, err := channel.QueueDeclare(
		"chat_messages",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %v", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
		queue:   queue,
	}, nil
}

// PublishMessage publishes a message to the RabbitMQ queue
func (r *RabbitMQ) PublishMessage(message []byte) error {
	err := r.channel.Publish(
		"",
		r.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %v", err)
	}

	return nil
}

// ConsumeMessages consumes messages from the RabbitMQ queue
func (r *RabbitMQ) ConsumeMessages() (<-chan amqp.Delivery, error) {
	messages, err := r.channel.Consume(
		r.queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %v", err)
	}

	return messages, nil
}

// Close closes the RabbitMQ connection and channel
func (r *RabbitMQ) Close() {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			log.Printf("Error closing RabbitMQ channel: %v", err)
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		}
	}
}
