package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Settings struct {
	// OpenSearch
	OpenSearchEndpoint string
	OpenSearchUser     string
	OpenSearchPwd      string
	OpenSearchIndex    string

	// Kafka/MSK
	KafkaBootstrapServers string
	KafkaTopic            string

	// AWS
	AWSRegion string
}

func LoadSettings() (*Settings, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	settings := &Settings{
		OpenSearchEndpoint:    os.Getenv("OPENSEARCH_ENDPOINT"),
		OpenSearchUser:        os.Getenv("OPENSEARCH_USER"),
		OpenSearchPwd:         os.Getenv("OPENSEARCH_PWD"),
		OpenSearchIndex:       getEnvOrDefault("OPENSEARCH_INDEX", "test_aws_service"),
		KafkaBootstrapServers: os.Getenv("MSK_BOOTSTRAP_SERVERS"),
		KafkaTopic:            getEnvOrDefault("MSK_TOPIC", "test-topic"),
		AWSRegion:             getEnvOrDefault("AWS_REGION", "us-east-1"),
	}

	return settings, settings.Validate()
}

func (s *Settings) Validate() error {
	if s.OpenSearchEndpoint == "" {
		return fmt.Errorf("OPENSEARCH_ENDPOINT is required")
	}
	if s.KafkaBootstrapServers == "" {
		return fmt.Errorf("MSK_BOOTSTRAP_SERVERS is required")
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}