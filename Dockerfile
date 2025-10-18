FROM golang:1.23-alpine AS builder
ARG VERSION=dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags "-X main.version=$VERSION" -o coo-llm ./cmd/coo-llm

FROM node:22-alpine AS ui-builder
ARG VERSION=dev
ARG BUILD_UI=true
WORKDIR /app
COPY webui/ ./
RUN mkdir -p /ui-build
RUN if [ "$BUILD_UI" = "true" ]; then \
      npm ci && \
      npm run build && \
      cp -r build/* /ui-build/; \
    fi

FROM alpine:3.19
ARG BUILD_UI=true
RUN apk update && apk upgrade && rm -rf /var/cache/apk/*
RUN addgroup -g 1001 -S appgroup && adduser -u 1001 -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/coo-llm .
# Copy UI build if BUILD_UI is true
RUN if [ "$BUILD_UI" = "true" ]; then \
      mkdir -p ./webui/build && \
      cp -r /ui-build/* ./webui/build/ 2>/dev/null || true; \
    fi
USER appuser
# Default config path, can be overridden with -config flag or CONFIG_PATH env var
CMD ["sh", "-c", "./coo-llm ${CONFIG_PATH:+ -config $CONFIG_PATH}"]
