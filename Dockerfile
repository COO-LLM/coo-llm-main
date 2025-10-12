FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o truckllm ./cmd/truckllm

FROM alpine:3.19
RUN apk update && apk upgrade && rm -rf /var/cache/apk/*
RUN addgroup -g 1001 -S appgroup && adduser -u 1001 -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/truckllm .
COPY configs/config.yaml ./configs/config.yaml
USER appuser
CMD ["./truckllm", "-config", "configs/config.yaml"]
