package main

import (
	"os"
	"os/signal"
	"syscall"

	"aws-go-ana/internal/config"
	"aws-go-ana/internal/services"
	"aws-go-ana/internal/utils"
)

func main() {
	logger := utils.NewLogger()

	// Load configuration
	cfg, err := config.LoadSettings()
	if err != nil {
		logger.Fatalf("Failed to load settings: %v", err)
	}

	// Create Kafka service
	kafkaService, err := services.NewKafkaService(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to create Kafka service: %v", err)
	}
	defer kafkaService.Close()

	logger.Info("Starting Kafka load test...")

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start load test in goroutine
	go func() {
		kafkaService.GenerateLoad(1, 10, 0) // 5 threads, 500 msg/sec, no duration limit
	}()

	// Wait for interrupt signal
	<-sigCh
	logger.Info("Received interrupt signal, shutting down...")
}
