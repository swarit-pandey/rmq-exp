package rmq

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
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
	notifyClose   chan *amqp.Error
	notifyConfirm chan amqp.Confirmation
	isConnected   bool
	done          chan struct{}
	routingKey    string
	mu            sync.RWMutex
	shuttingDown  atomic.Bool
}

func NewPublisher(conf config.RabbitMQConf, routingKey string) *Publisher {
	p := &Publisher{
		config:     conf,
		routingKey: routingKey,
		done:       make(chan struct{}),
	}

	go p.handleReconnection()
	return p
}

func (p *Publisher) handleReconnection() {
	backoff := time.Second

	for {
		select {
		case <-p.done:
			return
		default:
			p.mu.Lock()
			p.isConnected = false
			p.mu.Unlock()

			if err := p.connect(); err != nil {
				slog.Error("Failed to connect", "error", err)
				time.Sleep(backoff)
				backoff = min(backoff*2, 30*time.Second)
				continue
			}

			if err := p.setup(); err != nil {
				slog.Error("Failed to setup", "error", err)
				p.cleanup()
				time.Sleep(backoff)
				backoff = min(backoff*2, 30*time.Second)
				continue
			}

			backoff = time.Second
			p.mu.Lock()
			p.isConnected = true
			p.mu.Unlock()

			select {
			case <-p.done:
				return
			case err := <-p.notifyClose:
				slog.Error("Connection closed", "error", err)
				p.cleanup()
			}
		}
	}
}

func (p *Publisher) setup() error {
	if err := p.declareExchange(); err != nil {
		return fmt.Errorf("exchange setup failed: %w", err)
	}

	if err := p.declareQueue(); err != nil {
		return fmt.Errorf("queue setup failed: %w", err)
	}

	return nil
}

func (p *Publisher) connect() error {
	conf := p.config.Conection
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		conf.Username, conf.Password, conf.URL, conf.Port)

	conn, err := amqp.Dial(connStr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	if p.config.QoS.Count > 0 {
		if err := ch.Qos(
			p.config.QoS.Count,
			p.config.QoS.Size,
			p.config.QoS.Global,
		); err != nil {
			ch.Close()
			conn.Close()
			return fmt.Errorf("failed to set QoS: %w", err)
		}
	}

	if err := ch.Confirm(false); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to enable confirms: %w", err)
	}

	p.connection = conn
	p.channel = ch
	p.notifyClose = make(chan *amqp.Error, 1)
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
	queue, err := p.channel.QueueDeclare(
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
	if p.shuttingDown.Load() {
		return fmt.Errorf("publisher is shutting down")
	}

	p.mu.RLock()
	connected := p.isConnected
	p.mu.RUnlock()

	if !connected {
		return fmt.Errorf("not connected to rabbitmq")
	}

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Body:         data,
	}

	// Use a separate context with timeout for publish operation
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := p.channel.PublishWithContext(
		publishCtx,
		p.config.Exchange.Name,
		p.routingKey,
		false,
		false,
		msg,
	)
	if err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	select {
	case confirm := <-p.notifyConfirm:
		if !confirm.Ack {
			return fmt.Errorf("publish not confirmed")
		}
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return fmt.Errorf("publish confirmation timeout")
	}

	return nil
}

func (p *Publisher) Close() error {
	p.shuttingDown.Store(true)
	close(p.done)

	// Give ongoing publishes time to complete
	select {
	case <-time.After(3 * time.Second):
		// Wait
	}

	p.cleanup()
	return nil
}

func (p *Publisher) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.channel != nil {
		p.channel.Close()
		p.channel = nil
	}
	if p.connection != nil {
		p.connection.Close()
		p.connection = nil
	}
}

func (p *Publisher) IsConnected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isConnected
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
