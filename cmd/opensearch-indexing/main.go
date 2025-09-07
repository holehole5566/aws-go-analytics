package main

import (
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

	// Create OpenSearch service
	osService, err := services.NewOpenSearchService(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to create OpenSearch service: %v", err)
	}

	indexName := cfg.OpenSearchIndex
	logger.Infof("Starting OpenSearch indexing to index: %s", indexName)

	// Generate sample documents
	documents := services.GenerateAWSDocuments(500)
	logger.Infof("Generated %d documents", len(documents))

	// Create index if it doesn't exist
	if err := osService.CreateIndex(indexName, nil); err != nil {
		logger.Warnf("Failed to create index (may already exist): %v", err)
	}

	// Bulk index documents
	if err := osService.BulkIndex(indexName, documents); err != nil {
		logger.Fatalf("Failed to bulk index documents: %v", err)
	}

	logger.Infof("Successfully indexed %d documents to %s", len(documents), indexName)
}