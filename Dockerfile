# Stage 1: build backend
FROM golang:1.23-alpine AS builder
ARG VERSION=dev
WORKDIR /app

# git in case go mod needs it
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags "-X main.version=$VERSION" -o coo-llm ./cmd/coo-llm

# Stage 2: build UI
FROM node:22-alpine AS ui-builder
ARG BUILD_UI=true
WORKDIR /app

# copy only webui context
COPY webui/ ./

RUN if [ "$BUILD_UI" = "true" ]; then \
  npm ci && \
  npm run build && \
  mkdir -p /ui-build && \
  cp -r build/* /ui-build/; \
  else \
  echo "Skipping UI build"; \
  fi

# Stage 3: runtime image
FROM alpine:3.19 AS runtime
ARG BUILD_UI=true
RUN apk add --no-cache ca-certificates tzdata

# create user
RUN addgroup -g 1001 -S appgroup && adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# copy backend
COPY --from=builder /app/coo-llm .

# copy UI build from ui-builder (only if built)
COPY --from=ui-builder /ui-build ./webui/build

# ensure executable bit
RUN chmod +x /app/coo-llm || true

USER appuser

# exec form so signals forwarded correctly
ENTRYPOINT ["./coo-llm"]
CMD ["-config", ""] 
