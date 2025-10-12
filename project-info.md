# 🧠 TruckLLM

**Một reverse proxy thông minh cho các hệ thống LLM**, tương thích hoàn toàn với OpenAI API, giúp cân bằng tải giữa nhiều tài khoản (API keys) và nhiều nhà cung cấp (OpenAI, Gemini, Claude, v.v.), đồng thời hỗ trợ logging, storage linh hoạt và cấu hình YAML như Docker Compose.

---

## ⚙️ 1. Mục tiêu hệ thống

| Mục tiêu                    | Mô tả                                                                                                    |
| --------------------------- | -------------------------------------------------------------------------------------------------------- |
| **API tương thích OpenAI**  | Người dùng chỉ cần thay `https://api.openai.com/v1` → `https://llm-balancer.local/v1` mà không đổi code. |
| **Cân bằng tải thông minh** | Giữa nhiều API key và nhiều provider để tránh 403 / 429 / overload.                                      |
| **Cấu hình linh hoạt**      | Tất cả config trong file YAML giống Docker Compose / K8s.                                                |
| **Quan sát & giám sát dễ**  | Tích hợp Prometheus, file log, hoặc user-defined log provider.                                           |
| **Mở rộng dễ dàng**         | Cho phép thêm provider mới (local hoặc public API) mà không đổi code.                                    |
| **Hiệu năng cao**           | Viết bằng Go, hỗ trợ concurrency & streaming tốt.                                                        |

---

## 🧩 2. Kiến trúc tổng thể

### 🔷 Sơ đồ tổng quan

```
                  ┌────────────────────────────┐
                  │   Client / SDK (OpenAI)   │
                  │  (requests to /v1/* APIs) │
                  └──────────────┬────────────┘
                                 │
                    ┌────────────▼─────────────┐
                    │   LLM Provider Balancer  │
                    │ (drop-in OpenAI gateway) │
                    ├───────────────────────────┤
                    │  API Layer (/v1 routes)   │
                    │  Balancer Logic           │
                    │  Provider Adapters        │
                    │  Storage (Redis/File)     │
                    │  Logging (File/Prom/Graf) │
                    │  Config Loader (YAML)     │
                    └────────────┬──────────────┘
                                 │
            ┌────────────────────┼────────────────────┐
            │                    │                    │
┌────────────▼────────────┐┌─────▼─────────────┐┌─────▼──────────────┐
│ OpenAI Provider Adapter ││ Gemini Adapter    ││ Claude Adapter     │
│  (api.openai.com)       ││  (googleapis.com) ││  (api.anthropic.com)│
└─────────────────────────┘└───────────────────┘└────────────────────┘
```

---

## 📦 3. Cấu trúc dự án (Go)

```
llm-balancer/
├── cmd/
│   └── main.go
├── internal/
│   ├── api/
│   │   ├── chat_completions.go
│   │   ├── completions.go
│   │   ├── embeddings.go
│   │   └── models.go
│   ├── balancer/
│   │   └── selector.go
│   ├── provider/
│   │   ├── interface.go
│   │   ├── openai.go
│   │   ├── gemini.go
│   │   └── claude.go
│   ├── config/
│   │   └── config.go
│   ├── store/
│   │   ├── runtime_store.go
│   │   └── config_store.go
│   └── log/
│       ├── logger.go
│       └── prometheus.go
├── config.example.yaml
├── go.mod
└── go.sum
```

---

## 🧱 4. Cấu hình YAML (ví dụ)

```yaml
version: "1.0"

server:
  listen: ":8080"
  admin_api_key: "admin-secret"

logging:
  file:
    enabled: true
    path: "./logs/llm.log"
    max_size_mb: 100
    max_backups: 5
  prometheus:
    enabled: true
    endpoint: "/metrics"
  providers:
    - name: "webhook"
      type: "http"
      endpoint: "https://logs.example.com/ingest"
      batch:
        enabled: true
        size: 50
        interval_seconds: 10

storage:
  config:
    type: "file"
    path: "./config.yaml"
  runtime:
    type: "redis"
    addr: "localhost:6379"
    password: ""

providers:
  - id: "openai"
    name: "OpenAI"
    base_url: "https://api.openai.com/v1"
    keys:
      - id: "oa-1"
        secret: "${OPENAI_KEY_1}"
        limit_req_per_min: 200
        limit_tokens_per_min: 100000
  - id: "gemini"
    name: "Gemini"
    base_url: "https://generativelanguage.googleapis.com/v1"
    keys:
      - id: "gm-1"
        secret: "${GEMINI_KEY_1}"
        limit_req_per_min: 150
        limit_tokens_per_min: 80000
  - id: "claude"
    name: "Claude"
    base_url: "https://api.anthropic.com/v1"
    keys:
      - id: "cl-1"
        secret: "${CLAUDE_KEY_1}"
        limit_req_per_min: 100
        limit_tokens_per_min: 60000

model_aliases:
  gpt-4o: openai:gpt-4o
  gemini-1.5-pro: gemini:gemini-1.5-pro
  claude-3-opus: claude:claude-3-opus

policy:
  strategy: "hybrid" # cost-first | least-used | hybrid
  hybrid_weights:
    token_ratio: 0.5
    req_ratio: 0.3
    error_score: 0.1
    latency: 0.1
```

---

## 🔧 5. API Interface (tương thích OpenAI)

| Method | Endpoint               | Mô tả                               |
| ------ | ---------------------- | ----------------------------------- |
| `POST` | `/v1/chat/completions` | Tạo phản hồi hội thoại (như OpenAI) |
| `POST` | `/v1/completions`      | Sinh văn bản thuần                  |
| `POST` | `/v1/embeddings`       | Tạo vector embedding                |
| `GET`  | `/v1/models`           | Danh sách models hiện có            |

> Tất cả dùng `Authorization: Bearer <api_key>` — mapping tới provider config.

---

## ⚖️ 6. Balancer Logic

### Mục tiêu:

* Phân phối request hợp lý theo:

  * Số request / phút (`req_usage`)
  * Số token / phút (`token_usage`)
  * Error rate
  * Latency trung bình

### Pseudocode:

```go
func SelectBest(model string) (*Provider, string) {
    provider := config.ResolveProvider(model)
    keys := runtime.GetActiveKeys(provider)
    
    best := keys[0]
    score := math.MaxFloat64
    for _, key := range keys {
        s := weightedScore(key)
        if s < score {
            best = key
            score = s
        }
    }
    return provider, best
}
```

---

## 🌉 7. Provider Interface

```go
type Provider interface {
	Name() string
	Generate(ctx context.Context, req Request) (*Response, error)
	ListModels(ctx context.Context) ([]string, error)
}

type Request struct {
	Model string                 `json:"model"`
	Input map[string]interface{} `json:"input"`
}

type Response struct {
	RawResponse []byte
	HTTPCode    int
	Err         error
}
```

---

## 🧩 8. API Layer (Chat Completion ví dụ)

```go
func ChatCompletionsHandler(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	json.NewDecoder(r.Body).Decode(&payload)
	model := payload["model"].(string)

	provider, key := balancer.SelectBest(model)
	resp, err := provider.Generate(r.Context(), balancer.ToProviderRequest(payload))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.HTTPCode)
	w.Write(resp.RawResponse)
}
```

---

## 🧰 9. Storage Layer

### Runtime Store (Redis)

* Giữ usage count / minute / key
* Dạng:

  ```
  key_usage:{provider}:{key_id}:req
  key_usage:{provider}:{key_id}:tokens
  ```
* TTL: 60s
* Dùng Lua script để update atomic.

### Config Store

* `FileConfigStore`: load YAML từ file (default)
* `HTTPConfigStore`: load từ remote API
* `S3ConfigStore`: optional (cho distributed config)

---

## 🧾 10. Logging

### Tùy chọn logging:

| Loại             | Mục đích                         |
| ---------------- | -------------------------------- |
| File             | Log hoạt động & lỗi local        |
| Prometheus       | Metric cho Grafana               |
| Webhook Provider | Log tới endpoint user định nghĩa |

Mỗi entry:

```json
{
  "timestamp": "2025-10-12T12:00:00Z",
  "provider": "gemini",
  "model": "gemini-1.5-pro",
  "req_id": "uuid",
  "latency_ms": 423,
  "status": 200,
  "tokens": 190,
  "error": ""
}
```

---

## 🧭 11. Admin API (Quản trị cấu hình)

| Method | Endpoint                    | Mô tả                         |
| ------ | --------------------------- | ----------------------------- |
| `GET`  | `/admin/v1/config`          | Lấy config hiện tại           |
| `POST` | `/admin/v1/config/validate` | Kiểm tra YAML hợp lệ          |
| `POST` | `/admin/v1/config`          | Cập nhật config mới           |
| `POST` | `/admin/v1/reload`          | Hot-reload cấu hình           |
| `GET`  | `/admin/v1/providers`       | Liệt kê provider & trạng thái |
| `GET`  | `/admin/v1/logs`            | Tail log gần nhất             |

Bảo mật: `Authorization: Bearer <admin_api_key>`

---

## 🖥️ 12. Web Dashboard (tùy chọn)

Xây bằng **Vue 3 + Tailwind** hoặc **SvelteKit**:

* **Dashboard tổng quan**: TPS, token/min, error rate
* **Providers**: danh sách key, health, quota
* **Config editor**: YAML live validate + apply
* **Log viewer**: realtime tail + filter theo provider

---

## 🔒 13. Bảo mật

* Mã hóa API keys trong config (AES hoặc KMS)
* Chặn request không có Authorization
* Không ghi secret vào log
* Giới hạn request/second / IP (rate limiter)
* TLS bắt buộc trong production

---

## 🧮 14. Triển khai

**Dockerfile**

```dockerfile
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o llm-balancer ./cmd

FROM alpine
COPY --from=build /app/llm-balancer /usr/local/bin/
CMD ["llm-balancer", "-config", "/etc/llm/config.yaml"]
```

**docker-compose.yaml**

```yaml
services:
  llm-balancer:
    image: llm-balancer:latest
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/etc/llm/config.yaml
    environment:
      - REDIS_ADDR=redis:6379
  redis:
    image: redis:7
```

---

## 📊 15. Quan sát (Observability)

* `/metrics` → Prometheus endpoint
* Các metric:

  ```
  llm_requests_total{provider="openai"}
  llm_latency_ms_avg{provider="gemini"}
  llm_error_rate{provider="claude"}
  llm_active_keys_total
  ```
* Grafana dashboard template kèm sẵn.

---

## 🔧 16. Mở rộng

| Mở rộng              | Ý nghĩa                             |
| -------------------- | ----------------------------------- |
| Caching Layer        | Cache response embeddings / prompts |
| Weighted Cost Policy | Lựa chọn model rẻ nhất trước        |
| Token Translator     | Mapping model alias tự động         |
| Auto Key Disable     | Tự disable key khi 403 quá nhiều    |
| Multi-instance       | Redis cluster + stateless backend   |

---

## 💡 17. Ngôn ngữ & Thư viện chính

| Thành phần   | Công nghệ                      |
| ------------ | ------------------------------ |
| **Ngôn ngữ** | Go 1.23                        |
| **Web**      | `net/http`, `chi` hoặc `gin`   |
| **Storage**  | `go-redis`, `viper`, `yaml.v3` |
| **Logging**  | `zerolog` hoặc `zap`           |
| **Metrics**  | `prometheus/client_golang`     |
| **Testing**  | `testify`, `httptest`          |

---

## 🧰 18. Lợi ích chính

✅ Tương thích hoàn toàn với SDK OpenAI (Python, JS, v.v.)
✅ Cấu hình YAML dễ hiểu, portable
✅ Hỗ trợ multi-provider, multi-key, balancing
✅ Logging + metrics chuẩn production
✅ Mở rộng dễ (thêm provider, thêm log sink)
✅ Viết bằng Go → tốc độ, concurrency, memory tốt

---

## 🚀 19. Hướng phát triển tiếp theo

1. **Tích hợp caching / quota policy**
2. **Viết plugin provider (local llama, ollama, etc.)**
3. **Thêm CLI quản lý (`truckllmctl`)**
4. **WebUI admin (SvelteKit)**
5. **Triển khai Helm chart cho K8s**

---

## 🏁 20. Kết luận

Hệ thống **LLM Provider Balancer** này sẽ là một **lớp trung gian thống nhất** giúp:

* **Gom và tối ưu hóa nhiều provider LLM khác nhau**,
* **Cân bằng thông minh** dựa trên usage, lỗi, latency, chi phí,
* **Tương thích hoàn toàn OpenAI API** để người dùng **không cần đổi SDK**,
* Và được **viết bằng Go** — nhẹ, nhanh, dễ deploy.
