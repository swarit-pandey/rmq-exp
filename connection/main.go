package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// APIConfig holds RabbitMQ management API configuration
type APIConfig struct {
	BaseURL  string
	Username string
	Password string
}

// Config holds all configuration parameters
type Config struct {
	ConnectionURL   string
	APIConfig       APIConfig
	MinConnections  int
	MaxConnections  int
	StepSize        int
	PublishInterval time.Duration
	StabilizeTime   time.Duration
}

type Connection struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ConnectionURL, "url", "amqp://guest:guest@localhost:5672/", "RabbitMQ connection URL")
	flag.StringVar(&cfg.APIConfig.BaseURL, "api-url", "http://localhost:15672", "RabbitMQ management API base URL")
	flag.StringVar(&cfg.APIConfig.Username, "api-user", "guest", "RabbitMQ management API username")
	flag.StringVar(&cfg.APIConfig.Password, "api-pass", "guest", "RabbitMQ management API password")
	flag.IntVar(&cfg.MinConnections, "min", 10, "Minimum number of connections")
	flag.IntVar(&cfg.MaxConnections, "max", 500, "Maximum number of connections")
	flag.IntVar(&cfg.StepSize, "step", 50, "Connection count increment step")
	publishInterval := flag.Int("publish-interval", 100, "Message publish interval in milliseconds")
	stabilizeTime := flag.Int("stabilize-time", 10, "Time to wait for metrics to stabilize in seconds")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of RabbitMQ Connection Tester:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if cfg.MinConnections < 1 {
		log.Fatal("Minimum connections must be at least 1")
	}
	if cfg.MaxConnections < cfg.MinConnections {
		log.Fatal("Maximum connections must be greater than or equal to minimum connections")
	}
	if cfg.StepSize < 1 {
		log.Fatal("Step size must be at least 1")
	}

	if _, err := url.Parse(cfg.APIConfig.BaseURL); err != nil {
		log.Fatalf("Invalid API URL format: %v", err)
	}

	cfg.PublishInterval = time.Duration(*publishInterval) * time.Millisecond
	cfg.StabilizeTime = time.Duration(*stabilizeTime) * time.Second

	return cfg
}

func generateMessage(size int) []byte {
	message := make([]byte, size)
	rand.Read(message)
	return message
}

func startPublishing(ctx context.Context, conn *Connection, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			messageSize := rand.Intn(9*1024) + 1024
			message := generateMessage(messageSize)

			err := conn.channel.PublishWithContext(ctx,
				"",         // exchange
				conn.queue, // routing key
				false,      // mandatory
				false,      // immediate
				amqp.Publishing{
					ContentType: "application/octet-stream",
					Body:        message,
				})
			if err != nil {
				log.Printf("Failed to publish message: %v", err)
			}
		}
	}
}

func startConsuming(ctx context.Context, conn *Connection) {
	msgs, err := conn.channel.Consume(
		conn.queue, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		log.Printf("Failed to start consuming: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-msgs:
			_ = msg
		}
	}
}

func createConnection(url string, queueNamePrefix string, index int) (*Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	queueName := fmt.Sprintf("%s-%d", queueNamePrefix, index)
	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		true,      // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %v", err)
	}

	return &Connection{
		conn:    conn,
		channel: ch,
		queue:   q.Name,
	}, nil
}

func cleanup(connections []*Connection) {
	for _, conn := range connections {
		if conn.channel != nil {
			conn.channel.Close()
		}
		if conn.conn != nil {
			conn.conn.Close()
		}
	}
}

func makeAPIRequest(cfg APIConfig, path string) (*http.Response, error) {
	apiURL := fmt.Sprintf("%s/api/%s", strings.TrimRight(cfg.BaseURL, "/"), strings.TrimLeft(path, "/"))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(cfg.Username, cfg.Password)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return resp, nil
}

func getDetailedMemoryStats(cfg APIConfig, nodeName string) (map[string]interface{}, error) {
	path := fmt.Sprintf("nodes/%s?memory=true", url.PathEscape(nodeName))
	resp, err := makeAPIRequest(cfg, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %v", err)
	}
	defer resp.Body.Close()

	var nodeStats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&nodeStats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return nodeStats, nil
}

func getNodes(cfg APIConfig) ([]map[string]interface{}, error) {
	resp, err := makeAPIRequest(cfg, "nodes")
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %v", err)
	}
	defer resp.Body.Close()

	var nodes []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return nodes, nil
}

func printStats(cfg APIConfig, currentConnections int, startTime time.Time) error {
	nodes, err := getNodes(cfg)
	if err != nil {
		return fmt.Errorf("failed to get nodes: %v", err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes found")
	}

	nodeName, ok := nodes[0]["name"].(string)
	if !ok {
		return fmt.Errorf("could not get node name")
	}

	// Get detailed memory stats
	stats, err := getDetailedMemoryStats(cfg, nodeName)
	if err != nil {
		return fmt.Errorf("failed to get detailed stats: %v", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("\nStats after %s:\n", duration.Round(time.Second))
	fmt.Printf("Current Connections: %d\n", currentConnections)

	if memoryStats, ok := stats["memory"].(map[string]interface{}); ok {
		fmt.Println("\nConnection Memory Breakdown:")

		readers, _ := memoryStats["connection_readers"].(float64)
		writers, _ := memoryStats["connection_writers"].(float64)
		channels, _ := memoryStats["connection_channels"].(float64)
		connOther, _ := memoryStats["connection_other"].(float64)

		totalConnMemory := readers + writers + channels + connOther

		fmt.Printf("Connection Readers: %.2f MB\n", readers/1024/1024)
		fmt.Printf("Connection Writers: %.2f MB\n", writers/1024/1024)
		fmt.Printf("Connection Channels: %.2f MB\n", channels/1024/1024)
		fmt.Printf("Connection Other: %.2f MB\n", connOther/1024/1024)
		fmt.Printf("Total Connection Memory: %.2f MB\n", totalConnMemory/1024/1024)

		if currentConnections > 0 {
			fmt.Printf("Average Memory per Connection: %.2f KB\n", totalConnMemory/float64(currentConnections)/1024)
		}

		// Total memory stats
		if total, ok := memoryStats["total"].(map[string]interface{}); ok {
			fmt.Println("\nTotal Memory Stats:")
			if erlang, ok := total["erlang"].(float64); ok {
				fmt.Printf("Erlang Memory: %.2f MB\n", erlang/1024/1024)
			}
			if rss, ok := total["rss"].(float64); ok {
				fmt.Printf("RSS Memory: %.2f MB\n", rss/1024/1024)
			}
			if allocated, ok := total["allocated"].(float64); ok {
				fmt.Printf("Allocated Memory: %.2f MB\n", allocated/1024/1024)
			}
		}
	}

	if err := printMessageStats(cfg); err != nil {
		return fmt.Errorf("failed to print message stats: %v", err)
	}

	fmt.Println("----------------------------------------")
	return nil
}

func printMessageStats(cfg APIConfig) error {
	resp, err := makeAPIRequest(cfg, "overview")
	if err != nil {
		return fmt.Errorf("failed to get overview stats: %v", err)
	}
	defer resp.Body.Close()

	var overviewStats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&overviewStats); err != nil {
		return fmt.Errorf("failed to decode overview stats: %v", err)
	}

	if messageStats, ok := overviewStats["message_stats"].(map[string]interface{}); ok {
		fmt.Println("\nMessage Stats:")
		if publishDetails, ok := messageStats["publish_details"].(map[string]interface{}); ok {
			if rate, ok := publishDetails["rate"].(float64); ok {
				fmt.Printf("Message Publish Rate: %.2f/s\n", rate)
			}
		}
		if deliverDetails, ok := messageStats["deliver_details"].(map[string]interface{}); ok {
			if rate, ok := deliverDetails["rate"].(float64); ok {
				fmt.Printf("Message Delivery Rate: %.2f/s\n", rate)
			}
		}
	}

	return nil
}

func main() {
	config := parseFlags()

	var connections []*Connection
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		cancel()
		cleanup(connections)
		os.Exit(0)
	}()

	startTime := time.Now()

	for currentTarget := config.MinConnections; currentTarget <= config.MaxConnections; currentTarget += config.StepSize {
		fmt.Printf("\nScaling to %d connections...\n", currentTarget)

		currentCount := len(connections)
		var wg sync.WaitGroup

		for i := currentCount; i < currentTarget; i++ {
			conn, err := createConnection(config.ConnectionURL, "queue", i)
			if err != nil {
				log.Printf("Failed to create connection %d: %v", i, err)
				break
			}
			connections = append(connections, conn)

			wg.Add(2)
			go func(c *Connection) {
				defer wg.Done()
				startPublishing(ctx, c, config.PublishInterval)
			}(conn)

			go func(c *Connection) {
				defer wg.Done()
				startConsuming(ctx, c)
			}(conn)
		}

		fmt.Printf("Waiting %v for metrics to stabilize...\n", config.StabilizeTime)
		time.Sleep(config.StabilizeTime)

		printStats(config.APIConfig, len(connections), startTime)
	}

	fmt.Println("\nTest complete. Press Ctrl+C to exit...")
	<-ctx.Done()
}
