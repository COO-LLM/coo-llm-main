FROM golang:1.24-rc-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o truckllm ./cmd/truckllm

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/truckllm .
COPY configs/config.yaml ./configs/config.yaml
CMD ["./truckllm", "-config", "configs/config.yaml"]
