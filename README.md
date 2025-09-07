# AWS Go Analytics

Go port of AWS analytics scripts with Kafka and OpenSearch integration.

## Project Structure

```
├── cmd/                        # Command-line applications
│   ├── kafka-load-test/       # Kafka load testing tool
│   └── opensearch-indexing/   # OpenSearch indexing tool
├── internal/                  # Private application code
│   ├── config/               # Configuration management
│   ├── services/             # Service implementations
│   └── utils/                # Utility functions
├── pkg/                      # Public library code
└── bin/                      # Built binaries
```

## Setup

1. **Install Go dependencies:**
   ```bash
   make deps
   ```

2. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your AWS service endpoints
   ```

3. **Build binaries:**
   ```bash
   make build
   ```

## Usage

### Kafka Load Testing
```bash
make run-kafka
# or
./bin/kafka-load-test
```

### OpenSearch Indexing
```bash
make run-opensearch
# or
./bin/opensearch-indexing
```

## Features

- **Kafka Integration**: High-performance message production with configurable load testing
- **OpenSearch Integration**: Document indexing and search capabilities
- **Structured Logging**: JSON-formatted logs with logrus
- **Configuration Management**: Environment-based configuration with validation
- **Graceful Shutdown**: Signal handling for clean application termination

## Configuration

All configuration is done via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENSEARCH_ENDPOINT` | OpenSearch domain endpoint | Required |
| `OPENSEARCH_USER` | OpenSearch username | Required |
| `OPENSEARCH_PWD` | OpenSearch password | Required |
| `OPENSEARCH_INDEX` | Target index name | `test_aws_service` |
| `MSK_BOOTSTRAP_SERVERS` | Kafka bootstrap servers | Required |
| `MSK_TOPIC` | Kafka topic name | `test-topic` |
| `AWS_REGION` | AWS region | `us-east-1` |

## Development

```bash
# Run tests
make test

# Clean build artifacts
make clean

# Format code
go fmt ./...

# Lint code
golangci-lint run
```