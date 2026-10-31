package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SMSJob struct {
	MessageID int64     `json:"message_id"`
	Timestamp time.Time `json:"timestamp"`
}

type RabbitMQClient struct {
	url       string
	conn      *amqp.Connection
	ch        *amqp.Channel
	queueName string
	mu        sync.RWMutex
	closed    bool
}

func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	client := &RabbitMQClient{
		url:       url,
		queueName: "sms_dispatch_queue",
	}
	if err := client.connect(); err != nil {
		return nil, err
	}
	return client, nil
}

func (r *RabbitMQClient) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := amqp.Dial(r.url)
	if err != nil {
		return fmt.Errorf("rabbitmq dial error: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("rabbitmq channel error: %w", err)
	}

	_, err = ch.QueueDeclare(
		r.queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("rabbitmq queue declare error: %w", err)
	}

	r.conn = conn
	r.ch = ch
	return nil
}

func (r *RabbitMQClient) PublishMessage(ctx context.Context, msgID int64) error {
	r.mu.RLock()
	ch := r.ch
	r.mu.RUnlock()

	if ch == nil {
		return fmt.Errorf("rabbitmq channel not initialized")
	}

	job := SMSJob{
		MessageID: msgID,
		Timestamp: time.Now(),
	}
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(ctx,
		"",
		r.queueName,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}

func (r *RabbitMQClient) StartConsumer(ctx context.Context, handler func(msgID int64) error) error {
	r.mu.RLock()
	ch := r.ch
	r.mu.RUnlock()

	if ch == nil {
		return fmt.Errorf("rabbitmq channel not initialized")
	}

	_ = ch.Qos(1, 0, false)

	msgs, err := ch.Consume(
		r.queueName,
		"",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("rabbitmq consume error: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				var job SMSJob
				if err := json.Unmarshal(d.Body, &job); err == nil {
					if err := handler(job.MessageID); err == nil {
						_ = d.Ack(false)
					} else {
						_ = d.Nack(false, true)
					}
				} else {
					_ = d.Ack(false)
				}
			}
		}
	}()

	return nil
}

func (r *RabbitMQClient) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	if r.ch != nil {
		_ = r.ch.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}
