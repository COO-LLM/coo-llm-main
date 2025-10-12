
# 🧠 TỔNG QUAN DỰ ÁN

## 🎯 Mục tiêu

`LLM Provider Balancing (LPB)` là **một service trung gian** tương thích với **OpenAI API**, cho phép:

* Gửi request đến nhiều LLM provider (OpenAI, Gemini, Claude, v.v.)
* Tự động **cân bằng tải theo token, tần suất, hiệu suất, chi phí, lỗi 403**
* Quản lý cấu hình linh hoạt qua file **YAML hoặc REST API**
* Cung cấp **log, metrics, storage backend** có thể mở rộng

---

# 🧩 KIẾN TRÚC HỆ THỐNG

```
┌──────────────────────┐
│    Client (SDK, app) │
│  - openai-python      │
│  - openai-go, etc.    │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────────────┐
│      LLM Provider Balancer   │
│   (API tương thích OpenAI)   │
│                              │
│   ┌────────────────────────┐ │
│   │ API Layer (OpenAI API) │ │
│   ├────────────────────────┤ │
│   │ Balancer Engine        │ │
│   │  - Token selector      │ │
│   │  - Error recovery      │ │
│   │  - Weight policy       │ │
│   ├────────────────────────┤ │
│   │ Provider Adapters      │ │
│   │  - openai.go           │ │
│   │  - gemini.go           │ │
│   │  - claude.go           │ │
│   ├────────────────────────┤ │
│   │ Storage & Log Plugins  │ │
│   │  - redis / file / yaml │ │
│   └────────────────────────┘ │
└──────────┬──────────────────┘
           │
           ▼
┌────────────────────────────────────────┐
│ External LLM Providers                 │
│  - OpenAI, Gemini, Anthropic, v.v.     │
└────────────────────────────────────────┘
```

---

# 🧱 CẤU TRÚC DỰ ÁN (Go)

```
llm-balancer/
├── cmd/
│   └── truckllm/
│       └── main.go
├── internal/
│   ├── api/              # OpenAI-compatible endpoints
│   │   ├── chat_completions.go
│   │   ├── completions.go
│   │   ├── embeddings.go
│   │   └── models.go
│   ├── balancer/
│   │   ├── selector.go   # Strategy logic
│   │   ├── metrics.go    # Prometheus counters
│   │   └── policy.go
│   ├── config/
│   │   └── config.go     # Load YAML → struct
│   ├── provider/
│   │   ├── interface.go
│   │   ├── openai.go
│   │   ├── gemini.go
│   │   ├── claude.go
│   │   └── registry.go
│   ├── log/
│   │   └── logger.go
│   ├── store/
│   │   ├── redis.go
│   │   ├── file.go
│   │   └── interface.go
│   └── utils/
│       └── retry.go
├── pkg/
│   └── middleware/
│       ├── auth.go
│       ├── request_id.go
│       └── logging.go
├── configs/
│   └── config.yaml
├── webui/                # optional management UI (Vue/React)
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

# ⚙️ CẤU HÌNH YAML (config.yaml)

```yaml
version: "1.0"

server:
  port: 8080
  log_level: info
  enable_metrics: true

balancer:
  strategy: round_robin   # or "least_error", "weighted"
  retry_on_403: true
  max_retry: 2

providers:
  - name: openai
    type: openai
    api_keys:
      - sk-xxx1
      - sk-xxx2
  - name: gemini
    type: gemini
    api_keys:
      - gk-abc1
      - gk-abc2
    base_url: https://generativelanguage.googleapis.com/v1
  - name: claude
    type: anthropic
    api_keys:
      - ak-xxx

models:
  gpt-4o: openai:gpt-4o
  gemini-1.5-pro: gemini:gemini-1.5-pro
  claude-3-opus: claude:claude-3-opus

storage:
  type: redis
  config:
    host: redis:6379
    password: ""
    db: 0

logging:
  type: file
  config:
    path: ./logs/requests.log
```

---

# 🐳 DOCKER SETUP

### **Dockerfile**

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o truckllm ./cmd/truckllm

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/truckllm .
COPY configs/config.yaml ./configs/config.yaml
CMD ["./truckllm", "-config", "configs/config.yaml"]
```

---

### **docker-compose.yml**

```yaml
version: '3.8'
services:
  truckllm:
    build: .
    ports:
      - "8080:8080"
    environment:
      - CONFIG_PATH=/app/configs/config.yaml
    depends_on:
      - redis
    volumes:
      - ./logs:/app/logs
  redis:
    image: redis:7
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  redis_data:
```

---

# 📊 OBSERVABILITY

### Prometheus metrics (mặc định bật ở `/metrics`):

| Metric                     | Mô tả                                     |
| -------------------------- | ----------------------------------------- |
| `llm_requests_total`       | Tổng số request từng provider             |
| `llm_failures_total`       | Số request lỗi                            |
| `llm_avg_latency_seconds`  | Độ trễ trung bình                         |
| `llm_balance_distribution` | Tỷ lệ request phân phối giữa các provider |

### Log plugin

* Hỗ trợ các backend:

  * `file` (mặc định)
  * `prometheus pushgateway`
  * `stdout`
  * `custom webhook` (người dùng định nghĩa)

---

# 🧠 BALANCING STRATEGY

| Strategy    | Mô tả                                   | Use case                         |
| ----------- | --------------------------------------- | -------------------------------- |
| Round Robin | Luân phiên token                        | Bình thường                      |
| Least Error | Ưu tiên token ít lỗi 403 nhất           | Chống rate-limit                 |
| Weighted    | Phân bổ theo trọng số (hiệu suất / giá) | Khi có nhiều tài khoản khác nhau |
| Smart       | Học từ log, dự đoán provider hiệu quả   | Giai đoạn mở rộng sau            |

---

# 🧰 CI/CD GỢI Ý

* **GitHub Actions**

  * Build + lint Go
  * Run tests (`go test ./...`)
  * Build Docker image → push Docker Hub
* **Makefile**

  ```makefile
  build:
      go build -o bin/truckllm ./cmd/truckllm
  run:
      ./bin/truckllm -config configs/config.yaml
  docker:
      docker build -t truckllm:latest .
  test:
      go test ./...
  ```

---

# 🌐 MỞ RỘNG TƯƠNG LAI

* 🔌 UI quản lý (webui/)

  * CRUD config YAML (front-end dùng Vue 3 + PrimeVue)
  * Realtime log & metrics chart
* 🧩 Plugin SDK (Go interface)

  * Cho phép cộng đồng viết thêm log/storage provider
* 🧮 Token Quota tracking (per-key, per-user)
* 💰 Cost optimization layer (giảm chi phí bằng cách chọn model rẻ khi tương đương)

---

# ✅ KẾT LUẬN

| Thành phần        | Công nghệ                            |
| ----------------- | ------------------------------------ |
| Language          | **Go**                               |
| Config            | **YAML**                             |
| API compatibility | **OpenAI format**                    |
| Storage           | Redis / File                         |
| Logging           | File / Prometheus / Plugin           |
| Observability     | Prometheus                           |
| Packaging         | Docker, Docker Compose               |
| Extension         | UI (Vue), Plugin system              |
| Strategy          | Round Robin / Weighted / Error-Aware |
