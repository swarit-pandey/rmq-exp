package rmq

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/swarit-pandey/rmq-exp/config"
)

type Consumer struct {
	config      config.RabbitMQConf
	connection  *amqp.Connection
	channel     *amqp.Channel
	done        chan struct{}
	notifyClose chan *amqp.Error
	isConnected bool
	bindingKey  string
	mu          sync.RWMutex
	receiver    <-chan amqp.Delivery
}

func NewConsumer(conf config.RabbitMQConf, bindingKey string) *Consumer {
	c := &Consumer{
		config:     conf,
		done:       make(chan struct{}),
		bindingKey: bindingKey,
	}

	go c.handleReconnection()
	return c
}

func (c *Consumer) handleReconnection() {
	backoff := time.Second

	for {
		select {
		case <-c.done:
			return
		default:
			c.mu.Lock()
			c.isConnected = false
			c.mu.Unlock()

			if err := c.connect(); err != nil {
				slog.Error("Failed to connect", "error", err)
				time.Sleep(backoff)
				backoff = min(backoff*2, 30*time.Second)
				continue
			}

			if err := c.setup(); err != nil {
				slog.Error("Failed to setup", "error", err)
				c.cleanup()
				time.Sleep(backoff)
				backoff = min(backoff*2, 30*time.Second)
				continue
			}

			backoff = time.Second
			c.mu.Lock()
			c.isConnected = true
			c.mu.Unlock()

			select {
			case <-c.done:
				return
			case err := <-c.notifyClose:
				slog.Error("Connection closed", "error", err)
				c.cleanup()
			}
		}
	}
}

func (c *Consumer) connect() error {
	conf := c.config.Conection
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		conf.Username,
		conf.Password,
		conf.URL,
		conf.Port,
	)

	conn, err := amqp.Dial(connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	if c.config.QoS.Count > 0 {
		err = ch.Qos(
			c.config.QoS.Count,
			c.config.QoS.Size,
			c.config.QoS.Global,
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return fmt.Errorf("failed to set QoS: %w", err)
		}
	}

	err = ch.Confirm(false)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	c.connection = conn
	c.channel = ch
	c.notifyClose = make(chan *amqp.Error)
	c.channel.NotifyClose(c.notifyClose)

	return nil
}

func (c *Consumer) setup() error {
	if err := c.declareExchange(); err != nil {
		return fmt.Errorf("exchange setup failed: %w", err)
	}

	if err := c.declareQueue(); err != nil {
		return fmt.Errorf("queue setup failed: %w", err)
	}

	return nil
}

// declareExchange will declare an exchange passively since this is consumer
func (c *Consumer) declareExchange() error {
	return c.channel.ExchangeDeclare(
		c.config.Exchange.Name,
		c.config.Exchange.Kind,
		c.config.Exchange.Durable,
		c.config.Exchange.AutoDelete,
		false,
		false,
		nil,
	)
}

// declareQueue will declare a queue
func (c *Consumer) declareQueue() error {
	queue, err := c.channel.QueueDeclare(
		c.config.Queue.Name,
		c.config.Queue.Durable,
		c.config.Queue.AutoDelete,
		false,
		false,
		nil,
	)

	err = c.channel.QueueBind(
		queue.Name,
		c.bindingKey,
		c.config.Exchange.Name,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind the queue %s to exchange %s: %w",
			queue.Name, c.config.Exchange.Name, err)
	}

	return nil
}

func (c *Consumer) cleanup() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.connection != nil {
		c.connection.Close()
	}
}

type Message struct {
	Body     []byte
	Original *amqp.Delivery
}

func (c *Consumer) Consume(ctx context.Context) (chan Message, error) {
	c.mu.RLock()
	connected := c.isConnected
	c.mu.RUnlock()

	if !connected {
		return nil, fmt.Errorf("not connected to rabbitmq")
	}

	msgChan := make(chan Message)

	deliveries, err := c.channel.Consume(
		c.config.Queue.Name,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start consuming: %w", err)
	}

	go func() {
		defer close(msgChan)
		for {
			select {
			case <-ctx.Done():
				return
			case delivery, ok := <-deliveries:
				if !ok {
					slog.Error("Delivery channel closed")
					return
				}
				msgChan <- Message{
					Body:     delivery.Body,
					Original: &delivery,
				}
			}
		}
	}()

	return msgChan, nil
}

func (c *Consumer) Close() {
	close(c.done)
	c.cleanup()
}

func (c *Consumer) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}
