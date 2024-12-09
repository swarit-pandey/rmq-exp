package stress

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/swarit-pandey/rmq-exp/config"
	"github.com/swarit-pandey/rmq-exp/generator"
	"github.com/swarit-pandey/rmq-exp/rmq"
	"google.golang.org/protobuf/proto"
)

type Pipeline struct {
	pipelineConf PipelineConfig
	publisher    *rmq.Publisher
	generator    *generator.Generator
	messageTypes []string
	batchSize    int
	numWorkers   int
	publishCount uint64
	errorCount   uint64
	done         chan struct{}
	startTime    time.Time
	wg           sync.WaitGroup
}

type PipelineConfig struct {
	RMQConfig      config.RabbitMQConf
	ProtoPath      string
	BatchSize      int
	BatchQueueSize int
	RoutingKey     string
	NumWorkers     int
}

func NewPipeline(cfg PipelineConfig) (*Pipeline, error) {
	gen, err := generator.NewGenerator(cfg.ProtoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create generator: %w", err)
	}

	messageTypes := gen.ListMessageTypes()
	pub := rmq.NewPublisher(cfg.RMQConfig, cfg.RoutingKey)

	return &Pipeline{
		pipelineConf: cfg,
		publisher:    pub,
		generator:    gen,
		messageTypes: messageTypes,
		batchSize:    cfg.BatchSize,
		numWorkers:   cfg.NumWorkers,
		done:         make(chan struct{}),
	}, nil
}

func (p *Pipeline) generateBatch() []proto.Message {
	batch := make([]proto.Message, p.batchSize)
	for i := 0; i < p.batchSize; i++ {
		msgType := "summary.SummaryEvent"
		msg, err := p.generator.GenerateMessage(msgType)
		if err != nil {
			slog.Error("Failed to generate message", "type", msgType, "error", err)
			continue
		}
		batch[i] = msg
	}
	return batch
}

func (p *Pipeline) Start(ctx context.Context) error {
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		if p.publisher.IsConnected() {
			break
		}
		if i == maxRetries-1 {
			return fmt.Errorf("failed to connect to RabbitMQ after %d retries", maxRetries)
		}
		slog.Info("Waiting for RabbitMQ connection...", "attempt", i+1)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}

	p.startTime = time.Now()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	batchChan := make(chan []proto.Message, p.pipelineConf.BatchQueueSize)
	messageChan := make(chan proto.Message, p.batchSize*2)

	// Batch generator goroutine
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer close(batchChan)
		for {
			select {
			case <-ctx.Done():
				return
			case <-p.done:
				return
			default:
				batch := p.generateBatch()
				select {
				case batchChan <- batch:
				case <-ctx.Done():
					return
				case <-p.done:
					return
				default:
					slog.Debug("Skipped batch generation - publishers are falling behind")
				}
			}
		}
	}()

	// Message distributor goroutine
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer close(messageChan)
		for {
			select {
			case <-ctx.Done():
				return
			case <-p.done:
				return
			case batch, ok := <-batchChan:
				if !ok {
					return
				}
				for _, msg := range batch {
					select {
					case messageChan <- msg:
					case <-ctx.Done():
						return
					case <-p.done:
						return
					default:
						slog.Debug("Message channel full, skipping message")
					}
				}
			}
		}
	}()

	// Start workers
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i, messageChan)
	}

	// Stats reporting
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-p.done:
				return
			case <-ticker.C:
				elapsed := time.Since(p.startTime).Seconds()
				published := atomic.LoadUint64(&p.publishCount)
				errors := atomic.LoadUint64(&p.errorCount)
				rate := float64(published) / elapsed

				slog.Info("Pipeline stats",
					"published", published,
					"errors", errors,
					"rate", fmt.Sprintf("%.2f msgs/sec", rate),
					"uptime", fmt.Sprintf("%.0f sec", elapsed),
					"workers", p.numWorkers,
				)
			}
		}
	}()

	return nil
}

func (p *Pipeline) worker(ctx context.Context, id int, messageChan <-chan proto.Message) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.done:
			return
		case msg, ok := <-messageChan:
			if !ok {
				return
			}

			data, err := proto.Marshal(msg)
			if err != nil {
				atomic.AddUint64(&p.errorCount, 1)
				slog.Error("Failed to marshal message", "error", err, "worker", id)
				continue
			}

			err = p.publisher.Publish(ctx, data)
			if err != nil {
				atomic.AddUint64(&p.errorCount, 1)
				slog.Error("Failed to publish message", "error", err, "worker", id)
				continue
			}
			atomic.AddUint64(&p.publishCount, 1)
		}
	}
}

func (p *Pipeline) Stop() {
	close(p.done)

	stopped := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(stopped)
	}()

	select {
	case <-stopped:
		p.publisher.Close()
	case <-time.After(3 * time.Second):
		slog.Warn("Force closing pipeline")
		p.publisher.Close()
	}
}
