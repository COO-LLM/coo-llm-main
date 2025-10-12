build:
	go build -o bin/truckllm ./cmd/truckllm

run:
	./bin/truckllm -config configs/config.yaml

docker:
	docker build -t truckllm:latest .

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/

.PHONY: build run docker test lint clean