.PHONY: build clean test run-kafka run-opensearch

# Build all binaries
build:
	go build -o bin/kafka-load-test ./cmd/kafka-load-test
	go build -o bin/opensearch-indexing ./cmd/opensearch-indexing

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run Kafka load test
run-kafka: build
	./bin/kafka-load-test

# Run OpenSearch indexing
run-opensearch: build
	./bin/opensearch-indexing

# Initialize project
init:
	go mod download
	mkdir -p bin/