package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"

	"aws-go-ana/internal/config"
)

type KafkaService struct {
	config   *config.Settings
	logger   *logrus.Logger
	producer sarama.SyncProducer
}

type LoadTestMessage struct {
	Timestamp time.Time `json:"timestamp"`
	ThreadID  int       `json:"thread_id"`
	Count     int       `json:"count"`
}

func NewKafkaService(cfg *config.Settings, logger *logrus.Logger) (*KafkaService, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Compression = sarama.CompressionGZIP
	config.Producer.Flush.Frequency = 5 * time.Millisecond
	config.Producer.Flush.Messages = 32768

	brokers := strings.Split(cfg.KafkaBootstrapServers, ",")
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &KafkaService{
		config:   cfg,
		logger:   logger,
		producer: producer,
	}, nil
}

func (k *KafkaService) SendMessage(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: k.config.KafkaTopic,
		Value: sarama.StringEncoder(data),
	}

	partition, offset, err := k.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	k.logger.Infof("Message sent to partition %d at offset %d", partition, offset)
	return nil
}

func (k *KafkaService) GenerateLoad(numThreads int, messagesPerSec int, duration time.Duration) {
	k.logger.Infof("Starting load test: %d threads, %d msg/sec", numThreads, messagesPerSec)

	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	// Stop after duration if specified
	if duration > 0 {
		go func() {
			time.Sleep(duration)
			close(stopCh)
		}()
	}

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go k.worker(i, messagesPerSec, stopCh, &wg)
	}

	wg.Wait()
	k.logger.Info("Load test completed")
}

func (k *KafkaService) worker(threadID int, messagesPerSec int, stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(time.Second / time.Duration(messagesPerSec))
	defer ticker.Stop()

	count := 0
	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			message := LoadTestMessage{
				Timestamp: time.Now(),
				ThreadID:  threadID,
				Count:     count,
			}

			if err := k.SendMessage(message); err != nil {
				k.logger.Errorf("Thread %d failed to send message: %v", threadID, err)
			}

			count++
			if count%100 == 0 {
				k.logger.Infof("Thread %d: %d messages sent", threadID, count)
			}
		}
	}
}

func (k *KafkaService) Close() error {
	return k.producer.Close()
}
