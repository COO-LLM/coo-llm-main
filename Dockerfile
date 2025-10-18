# ============================
# Stage 1: Build backend (Go)
# ============================
FROM golang:1.23-alpine AS builder
ARG VERSION=dev
WORKDIR /app

# install git if go mod need
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags "-X main.version=$VERSION" -o coo-llm ./cmd/coo-llm


# ============================
# Stage 2: Build UI (Node)
# ============================
FROM node:22-alpine AS ui-builder
ARG BUILD_UI=true
WORKDIR /app

# Copy webui
COPY webui/ ./

# Build UI
RUN if [ "$BUILD_UI" = "true" ]; then \
  echo "Building UI..." && \
  rm -rf /ui-build && mkdir -p /ui-build && \
  npm ci && \
  npm run build && \
  cp -r build/* /ui-build/; \
  else \
  rm -rf /ui-build && mkdir -p /ui-build; \
  fi

# ============================
# Stage 3: Runtime image
# ============================
FROM alpine:3.19 AS runtime
ARG BUILD_UI=true
RUN apk add --no-cache ca-certificates tzdata

# Create user to run app
RUN addgroup -g 1001 -S appgroup && adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binary from backend builder
COPY --from=builder /app/coo-llm .

# Copy UI build (if it's exist)
COPY --from=ui-builder /ui-build ./webui/build

# Set permission
RUN chmod +x /app/coo-llm || true

USER appuser

# Entry point
ENTRYPOINT ["./coo-llm"]
CMD ["-config", ""]
