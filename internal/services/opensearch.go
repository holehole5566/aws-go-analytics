package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/sirupsen/logrus"

	"aws-go-ana/internal/config"
)

type OpenSearchService struct {
	config *config.Settings
	logger *logrus.Logger
	client *opensearch.Client
}

type AWSDocument struct {
	ID           string    `json:"id"`
	Metadata     Metadata  `json:"metadata"`
	Message      string    `json:"message"`
	Time         int64     `json:"time"`
	Severity     string    `json:"severity"`
	ActivityName string    `json:"activity_name"`
	Timestamp    time.Time `json:"@timestamp"`
}

type Metadata struct {
	Product Product `json:"product"`
}

type Product struct {
	Name string `json:"name"`
}

func NewOpenSearchService(cfg *config.Settings, logger *logrus.Logger) (*OpenSearchService, error) {
	if cfg.OpenSearchEndpoint == "" || cfg.OpenSearchUser == "" || cfg.OpenSearchPwd == "" {
		return nil, fmt.Errorf("OpenSearch credentials not configured")
	}

	client, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Addresses: []string{fmt.Sprintf("https://%s", cfg.OpenSearchEndpoint)},
		Username:  cfg.OpenSearchUser,
		Password:  cfg.OpenSearchPwd,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	return &OpenSearchService{
		config: cfg,
		logger: logger,
		client: client,
	}, nil
}

func (o *OpenSearchService) IndexDocument(indexName string, document interface{}, docID string) error {
	data, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      indexName,
		DocumentID: docID,
		Body:       bytes.NewReader(data),
	}

	res, err := req.Do(context.Background(), o.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("indexing failed: %s", res.Status())
	}

	o.logger.Infof("Document indexed with ID: %s", docID)
	return nil
}

func (o *OpenSearchService) BulkIndex(indexName string, documents []AWSDocument) error {
	var buf bytes.Buffer

	for _, doc := range documents {
		// Action line
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_id":    doc.ID,
			},
		}
		actionBytes, _ := json.Marshal(action)
		buf.Write(actionBytes)
		buf.WriteByte('\n')

		// Document line
		docBytes, _ := json.Marshal(doc)
		buf.Write(docBytes)
		buf.WriteByte('\n')
	}

	req := opensearchapi.BulkRequest{
		Body: &buf,
	}

	res, err := req.Do(context.Background(), o.client)
	if err != nil {
		return fmt.Errorf("bulk indexing failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk indexing failed: %s", res.Status())
	}

	o.logger.Infof("Bulk indexed %d documents to %s", len(documents), indexName)
	return nil
}

func (o *OpenSearchService) Search(indexName string, query map[string]interface{}, size int) (map[string]interface{}, error) {
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{indexName},
		Body:  bytes.NewReader(queryBytes),
		Size:  &size,
	}

	res, err := req.Do(context.Background(), o.client)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search failed: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

func (o *OpenSearchService) CreateIndex(indexName string, mapping map[string]interface{}) error {
	var body map[string]interface{}
	if mapping != nil {
		body = map[string]interface{}{
			"mappings": mapping,
		}
	}

	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}

	req := opensearchapi.IndicesCreateRequest{
		Index: indexName,
		Body:  &buf,
	}

	res, err := req.Do(context.Background(), o.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && !strings.Contains(res.Status(), "400") { // Ignore if already exists
		return fmt.Errorf("failed to create index: %s", res.Status())
	}

	o.logger.Infof("Index created: %s", indexName)
	return nil
}

func GenerateAWSDocuments(numDocs int) []AWSDocument {
	services := []string{"EKS", "S3", "EC2", "RDS", "Lambda"}
	documents := make([]AWSDocument, numDocs)

	for i := 0; i < numDocs; i++ {
		service := services[i%len(services)]
		now := time.Now()

		documents[i] = AWSDocument{
			ID: uuid.New().String(),
			Metadata: Metadata{
				Product: Product{
					Name: service,
				},
			},
			Message:      "ResponseComplete",
			Time:         now.UnixMilli(),
			Severity:     "Informational",
			ActivityName: "Update",
			Timestamp:    now,
		}
	}

	return documents
}