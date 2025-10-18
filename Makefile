build:
	go build -o bin/coo-llm ./cmd/coo-llm

run:
	./bin/coo-llm -config configs/config.yaml

docker:
	docker build -t coo-llm:latest .

test:
	go test ./...

lint:
	golangci-lint run .

clean:
	rm -rf bin/

.PHONY: build run docker test lint clean