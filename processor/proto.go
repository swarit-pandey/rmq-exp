package processor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/swarit-pandey/rmq-exp/rmq"
	"google.golang.org/protobuf/proto"
)

type ProtoMessage interface {
	proto.Message
}

type Stats struct {
	processed uint64
	failed    uint64
	mu        sync.Mutex
}

// ProtoProcessor is meant to only process message where the wire type is
// protobuf

type Proto struct {
	source        *rmq.Consumer
	controlFactor int
	stats         *Stats
}

func NewConsumer(cons *rmq.Consumer, cf int) *Proto {
	return &Proto{
		source:        cons,
		controlFactor: cf,
		stats:         &Stats{},
	}
}

func (p *Proto) ProcessMessage(ctx context.Context, messageType ProtoMessage) error {
	msgChan, err := p.source.Consume(ctx)
	if err != nil {
		return fmt.Errorf("failed to get the message channel: %w", err)
	}

	tick := time.NewTicker(3 * time.Minute)
	defer tick.Stop()

	// Need to rate limit the number of messages based on congtrolFactor
	// to simulate slow and fast processors
	sem := make(chan struct{}, p.controlFactor)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled stopping message consumption")
			p.logStats()
			return nil

		case <-tick.C:
			p.logStats()

		case msg, ok := <-msgChan:
			if !ok {
				slog.Info("Message channel closed")
				return nil
			}

			sem <- struct{}{}

			go func(message rmq.Message) {
				defer func() { <-sem }()

				protoMsg := proto.Clone(messageType)

				err := proto.Unmarshal(message.Body, protoMsg)
				if err != nil {
					slog.Error("failed to unmarshal message", "error", err)
					message.Original.Nack(false, false)
					return
				}

				message.Original.Ack(false)
			}(msg)
		}
	}
}

func (p *Proto) logStats() {
	p.stats.mu.Lock()
	defer p.stats.mu.Unlock()

	slog.Info("Processing stats",
		"processed", p.stats.processed,
		"failed", p.stats.failed,
		"success_rate", fmt.Sprintf("%.2f%%", float64(p.stats.processed)/float64(p.stats.processed+p.stats.failed)*100))
}
