package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/accuknox/dev2/api/grpc/v2/summary"
	"github.com/swarit-pandey/rmq-exp/config"
	"github.com/swarit-pandey/rmq-exp/stress"
)

func main() {
	// Setup logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// RabbitMQ Configuration mein ye sab change kar lo apne hisaab se
	rmqConfig := config.RabbitMQConf{
		Conection: config.Connection{
			URL:      "127.0.0.1",
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
			Name:       "summary-v2",
			Durable:    true,
			AutoDelete: false,
		},
		QoS: config.QoS{
			Count:  100,
			Size:   0,
			Global: false,
		},
	}

	// Pipeline Configuration
	pipelineConfig := stress.PipelineConfig{
		RMQConfig:      rmqConfig,
		ProtoPath:      "/home/swarit/rmq-exp/summaryv2/summary.proto",
		BatchSize:      2000,
		BatchQueueSize: 1000,
		RoutingKey:     "stress_test",
		NumWorkers:     20,
	}

	pipeline, err := stress.NewPipeline(pipelineConfig)
	if err != nil {
		slog.Error("Failed to create pipeline", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if err := pipeline.Start(ctx); err != nil {
		slog.Error("Pipeline failed to start", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting stress test",
		"batchSize", pipelineConfig.BatchSize,
		"queuedBatches", pipelineConfig.BatchQueueSize,
		"qosCount", rmqConfig.QoS.Count,
	)

	<-sigChan
	slog.Info("Received shutdown signal")
	cancel()
	pipeline.Stop()
}
