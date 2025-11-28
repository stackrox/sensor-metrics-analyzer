.PHONY: build test lint clean validate-rules

build:
	go build -o bin/metrics-analyzer ./cmd/metrics-analyzer

test:
	go test -v ./...

lint:
	golangci-lint run || true

validate-rules:
	./bin/metrics-analyzer validate ./automated-rules

clean:
	rm -rf bin/

