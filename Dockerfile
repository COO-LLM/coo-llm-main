FROM golang:1.23-alpine AS builder
ARG VERSION=dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags "-X main.version=$VERSION" -o coo-llm ./cmd/coo-llm

FROM alpine:3.19
RUN apk update && apk upgrade && rm -rf /var/cache/apk/*
RUN addgroup -g 1001 -S appgroup && adduser -u 1001 -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/coo-llm .
USER appuser
# Default config path, can be overridden with -config flag
CMD ["./coo-llm"]
