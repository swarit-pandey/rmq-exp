package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/accuknox/dev2/api/grpc/v2/summary"
	"github.com/swarit-pandey/rmq-exp/config"
	"github.com/swarit-pandey/rmq-exp/processor"
	"github.com/swarit-pandey/rmq-exp/rmq"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	rmqConfig := config.RabbitMQConf{
		Conection: config.Connection{
			URL:      "localhost",
			Port:     "5672",
			Username: "admin",
			Password: "password",
		},
		Exchange: config.Exchange{
			Name:       "summary-v2",
			Kind:       "direct",
			Durable:    true,
			AutoDelete: true,
		},
		Queue: config.Queue{
			Name:        "summary-v2",
			Durable:     true,
			AutoDelete:  false,
			ConsumerTag: "summary-processor", // Added consumer tag
		},
		QoS: config.QoS{
			Count:  100,
			Size:   0,
			Global: false,
		},
	}

	consumer := rmq.NewConsumer(rmqConfig, "stress-test")
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Wait for initial connection
	for i := 0; i < 30; i++ {
		if consumer.IsConnected() {
			break
		}
		if i == 29 {
			slog.Error("Failed to connect to RabbitMQ after retries")
			return
		}
		slog.Info("Waiting for RabbitMQ connection...", "attempt", i+1)
		time.Sleep(time.Second)
	}

	processor := processor.NewConsumer(consumer, 150)

	errChan := make(chan error, 1)
	go func() {
		errChan <- processor.ProcessMessage(ctx, &summary.SummaryEvent{})
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		slog.Info("Shutdown signal received", "signal", sig)
		cancel()
	case err := <-errChan:
		if err != nil {
			slog.Error("Processor failed", "error", err)
		}
	}

	// Allow time for cleanup
	time.Sleep(2 * time.Second)
}
