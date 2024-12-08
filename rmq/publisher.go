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

const (
	defaultMaxRetries int = 15
)

type Publisher struct {
	config        config.RabbitMQConf
	connection    *amqp.Connection
	channel       *amqp.Channel
	done          chan os.Signal
	notifyClose   chan *amqp.Error
	notifyConfirm chan amqp.Confirmation
	isConnected   bool
	alive         bool
	routingKey    string
	mu            sync.Mutex
	maxRetries    int
}

func NewPublisher(conf config.RabbitMQConf, routingKey string) *Publisher {
	p := &Publisher{
		config:      conf,
		done:        make(chan os.Signal, 1),
		isConnected: false,
		alive:       true,
		routingKey:  routingKey,
		maxRetries:  defaultMaxRetries,
	}

	go p.handleReconnection()
	return p
}

func (p *Publisher) handleReconnection() error {
	for p.alive && p.maxRetries > 0 {
		p.mu.Lock()
		p.isConnected = false
		p.mu.Unlock()

		err := p.connect()
		if err != nil {
			slog.Error("Failed to connect", "error", err)
			time.Sleep(5 * time.Second)
			// p.maxRetries--
			continue
		}

		err = p.declareExchange()
		if err != nil {
			slog.Error("Failed to declare exchange", "error", err)
			p.cleanup()
			time.Sleep(5 * time.Second)
			// p.maxRetries--
			continue
		}

		err = p.declareQueue()
		if err != nil {
			slog.Error("Failed to declare queue", "error", err)
			p.cleanup()
			time.Sleep(5 * time.Second)
			// p.maxRetries--
			continue
		}

		p.mu.Lock()
		p.isConnected = true
		p.mu.Unlock()

		select {
		case <-p.done:
			p.alive = false
			return nil
		case err := <-p.notifyClose:
			slog.Error("Connection closed", "error", err)
			continue
		}
	}
	return nil
}

func (p *Publisher) cleanup() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.connection != nil {
		p.connection.Close()
	}
}

func (p *Publisher) connect() error {
	conf := p.config.Conection
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

	if p.config.QoS.Count > 0 {
		err = ch.Qos(
			p.config.QoS.Count,
			p.config.QoS.Size,
			p.config.QoS.Global,
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

	p.connection = conn
	p.channel = ch
	p.notifyClose = make(chan *amqp.Error)
	p.notifyConfirm = make(chan amqp.Confirmation, 1)
	p.channel.NotifyClose(p.notifyClose)
	p.channel.NotifyPublish(p.notifyConfirm)

	return nil
}

func (p *Publisher) declareExchange() error {
	return p.channel.ExchangeDeclare(
		p.config.Exchange.Name,
		p.config.Exchange.Kind,
		p.config.Exchange.Durable,
		p.config.Exchange.AutoDelete,
		false,
		false,
		nil,
	)
}

func (p *Publisher) declareQueue() error {
	queue, err := p.channel.QueueDeclarePassive(
		p.config.Queue.Name,
		p.config.Queue.Durable,
		p.config.Queue.AutoDelete,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = p.channel.QueueBind(
		queue.Name,
		p.routingKey,
		p.config.Exchange.Name,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind the queue %s to exchange %s: %w",
			queue.Name, p.config.Exchange.Name, err)
	}

	return nil
}

func (p *Publisher) Publish(ctx context.Context, data []byte) error {
	p.mu.Lock()
	connected := p.isConnected
	p.mu.Unlock()

	if !connected {
		return fmt.Errorf("not connected to rabbitmq")
	}

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Body:         data,
	}

	err := p.channel.PublishWithContext(
		ctx,
		p.config.Exchange.Name,
		p.routingKey,
		false,
		false,
		msg,
	)
	if err != nil {
		return fmt.Errorf("failed to publish the message: %w", err)
	}

	select {
	case confirm := <-p.notifyConfirm:
		if !confirm.Ack {
			return fmt.Errorf("failed to deliver message")
		}
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return fmt.Errorf("publish confirmation timeout")
	}

	return nil
}

func (p *Publisher) Close() error {
	p.alive = false
	p.done <- os.Interrupt
	p.cleanup()
	return nil
}

func (p *Publisher) IsConnected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isConnected
}
