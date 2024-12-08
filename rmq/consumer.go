package rmq

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/swarit-pandey/rmq-exp/config"
)

type Consumer struct {
	config      config.RabbitMQConf
	connection  *amqp.Connection
	channel     *amqp.Channel
	done        chan os.Signal
	notifyClose chan *amqp.Error
	isConnected bool
	alive       bool
	bindingKey  string
	mu          sync.Mutex
	maxRetries  int
	receiver    <-chan amqp.Delivery
}

func NewConsumer(conf config.RabbitMQConf, bindingKey string) *Consumer {
	c := &Consumer{
		config:      conf,
		done:        make(chan os.Signal),
		isConnected: false,
		alive:       true,
		bindingKey:  bindingKey,
		maxRetries:  defaultMaxRetries,
	}

	return c
}

func (c *Consumer) handleReconnection() {
	for err := range c.notifyClose {
		if !c.alive {
			return
		}

		slog.Error("Connection closed", "error", err)

		backoff := time.Second
		for retry := 0; retry < c.maxRetries; retry++ {
			if err := c.connect(); err != nil {
				slog.Error("Failed to reconnect", "attempt", retry+1, "error", err)
				time.Sleep(backoff)
				backoff *= 2
				continue
			}

			if err := c.declareExchange(); err != nil {
				slog.Error("Failed to declare exchange", "error", err)
				continue
			}

			if err := c.declareQueue(); err != nil {
				slog.Error("Failed to declare queue", "error", err)
				continue
			}

			slog.Info("Reconnected successfully")
			break
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

// declareExchange will declare an exchange passively since this is consumer
func (c *Consumer) declareExchange() error {
	return c.channel.ExchangeDeclarePassive(
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
	c.mu.Lock()
	connected := c.isConnected
	c.mu.Unlock()

	if !connected {
		return nil, fmt.Errorf("not connected to reabbitmq")
	}

	var err error
	c.receiver, err = c.channel.ConsumeWithContext(ctx,
		c.config.Queue.Name,
		c.config.Queue.ConsumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start consumption from queue %s, with binding key %s: %w", c.config.Queue.Name, c.bindingKey, err)
	}

	msgChan := make(chan Message)
	go func() {
		defer close(msgChan)
		for {
			select {
			case <-ctx.Done():
				slog.Info("Context cancelled, stopping consumption")
				return
			case msg, ok := <-c.receiver:
				if !ok {
					if c.channel.IsClosed() {
						slog.Error("Channel is closed", "queue", c.config.Queue.Name)
						return
					}
				}

				msgStruct := Message{}
				msgStruct.Body = msg.Body
				msgStruct.Original = &msg

				msgChan <- msgStruct
			}
		}
	}()

	return msgChan, nil
}
