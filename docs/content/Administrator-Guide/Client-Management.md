---
sidebar_position: 3
tags: [administrator-guide, client-management]
---

# Client Management

Guide to managing API clients, authentication, and access control in COO-LLM.

## Overview

COO-LLM supports multi-tenant client management with:
- **Dynamic client registration**: Create and manage API clients programmatically
- **Provider restrictions**: Limit clients to specific LLM providers
- **Usage tracking**: Monitor per-client metrics and costs
- **Access control**: Fine-grained permissions and restrictions

## Client Architecture

### Client Data Model

```go
type ClientInfo struct {
    ID               string   `json:"id"`
    APIKey           string   `json:"api_key"`
    Description      string   `json:"description"`
    AllowedProviders []string `json:"allowed_providers"`
    CreatedAt        int64    `json:"created_at"`
    LastUsed         int64    `json:"last_used"`
}
```

### Storage Backends

Clients can be stored in:
- **Memory**: For development and testing
- **Redis**: For distributed deployments
- **Database**: PostgreSQL, MySQL, MongoDB
- **File**: JSON/YAML configuration files

## Managing Clients

### Creating Clients

```bash
# Create a new client via Admin API
curl -X POST http://localhost:2906/admin/v1/clients \
  -H "Authorization: Bearer your-admin-key" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "prod-app-001",
    "api_key": "sk-prod-app-key-123",
    "description": "Production mobile application",
    "allowed_providers": ["openai", "anthropic"]
  }'
```

### Listing Clients

```bash
# List all clients
curl -H "Authorization: Bearer your-admin-key" \
  http://localhost:2906/admin/v1/clients/list
```

### Updating Clients

```bash
# Update client configuration
curl -X PUT http://localhost:2906/admin/v1/clients/prod-app-001 \
  -H "Authorization: Bearer your-admin-key" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated production mobile app",
    "allowed_providers": ["openai", "anthropic", "gemini"]
  }'
```

### Deleting Clients

```bash
# Remove a client
curl -X DELETE http://localhost:2906/admin/v1/clients/prod-app-001 \
  -H "Authorization: Bearer your-admin-key"
```

## Client Authentication

### API Key Validation

Clients authenticate using API keys:

```bash
# Client makes request with API key
curl http://localhost:2906/v1/chat/completions \
  -H "Authorization: Bearer sk-prod-app-key-123" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

### Provider Restrictions

Clients are restricted to allowed providers:

```json
{
  "client_id": "restricted-client",
  "allowed_providers": ["openai", "anthropic"],
  "description": "Client limited to OpenAI and Anthropic"
}
```

**Behavior:**
- ✅ Requests to allowed providers succeed
- ❌ Requests to disallowed providers return 403 Forbidden
- ✅ `["*"]` allows all providers (wildcard)

## Monitoring Client Usage

### Client Metrics API

Get detailed usage statistics per client:

```bash
# Get client metrics for last 24 hours
curl -H "Authorization: Bearer your-admin-key" \
  "http://localhost:2906/admin/v1/metrics/clients/prod-app-001?start=$(date -d '24 hours ago' +%s)&end=$(date +%s)"
```

**Response:**
```json
{
  "client_id": "prod-app-001",
  "total_requests": 1250,
  "total_tokens": 45000,
  "total_cost": 22.50,
  "success_rate": 0.987,
  "avg_latency": 145.2,
  "last_request_time": 1700001000
}
```

### Cost Tracking

Monitor client spending:

```bash
# Get cost breakdown by provider for a client
curl -H "Authorization: Bearer your-admin-key" \
  "http://localhost:2906/admin/v1/stats?group_by=provider&client_key=prod-app-001"
```

### Usage Alerts

Set up alerts for client behavior:

```yaml
# Prometheus alert rules for clients
groups:
  - name: client-usage
    rules:
      - alert: HighClientCost
        expr: rate(coo_llm_client_cost_total{client_id="prod-app-001"}[1h]) > 10
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "Client prod-app-001 exceeding cost threshold"

      - alert: ClientErrorSpike
        expr: rate(coo_llm_client_errors_total{client_id="prod-app-001"}[5m]) / rate(coo_llm_client_requests_total{client_id="prod-app-001"}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate for client prod-app-001"
```

## Security Best Practices

### API Key Management

1. **Rotate keys regularly**: Change API keys every 90 days
2. **Use strong keys**: Generate cryptographically secure API keys
3. **Monitor usage**: Alert on unusual access patterns
4. **Revoke compromised keys**: Immediately disable suspicious keys

### Access Control

1. **Principle of least privilege**: Only allow necessary providers
2. **Regular audits**: Review client permissions quarterly
3. **Environment separation**: Different keys for dev/staging/prod
4. **Rate limiting**: Apply per-client rate limits

### Monitoring

1. **Log all access**: Track every API call with client context
2. **Alert on anomalies**: Unusual request patterns or volumes
3. **Cost monitoring**: Set budgets and alerts per client
4. **Performance tracking**: Monitor latency and success rates

## Configuration Examples

### Development Setup

```yaml
# config.yaml - Development
api_keys:
  - key: "dev-key-123"
    allowed_providers: ["*"]  # Allow all providers for development

# Or use dynamic client management
# Clients created via Admin API will be stored in configured backend
```

### Production Setup

```yaml
# config.yaml - Production
storage:
  runtime:
    type: "redis"
    addr: "redis-cluster:6379"

# Clients managed dynamically via Admin API
# No static API keys in config for better security
```

### Multi-Tenant Setup

```yaml
# config.yaml - Multi-tenant
storage:
  runtime:
    type: "postgres"
    dsn: "postgres://user:pass@db:5432/coo_llm"

# All clients managed through Admin API
# Each tenant gets isolated client credentials
```

## Troubleshooting

### Common Issues

#### Client Authentication Fails

**Symptoms:**
- 401 Unauthorized responses
- "Invalid API key" errors

**Solutions:**
1. Verify API key is correct and active
2. Check key hasn't expired
3. Ensure key is URL-encoded if needed
4. Confirm client exists in storage

#### Provider Access Denied

**Symptoms:**
- 403 Forbidden responses
- "Provider not allowed" errors

**Solutions:**
1. Check client's `allowed_providers` list
2. Update client permissions via Admin API
3. Verify provider name spelling
4. Confirm provider is configured in system

#### Metrics Not Showing

**Symptoms:**
- Empty metrics responses
- Missing client data

**Solutions:**
1. Verify storage backend is configured
2. Check storage connectivity
3. Confirm metrics collection is enabled
4. Validate time range parameters

### Debug Commands

```bash
# Check client exists
curl -H "Authorization: Bearer admin-key" \
  http://localhost:2906/admin/v1/clients/debug-client

# Test client authentication
curl -H "Authorization: Bearer client-key" \
  http://localhost:2906/v1/models

# Check client metrics
curl -H "Authorization: Bearer admin-key" \
  "http://localhost:2906/admin/v1/metrics/clients/debug-client?start=$(date +%s)"
```

## Integration Examples

### Programmatic Client Management

```python
import requests

ADMIN_KEY = "your-admin-key"
BASE_URL = "http://localhost:2906"

def create_client(client_id, api_key, description, providers):
    response = requests.post(
        f"{BASE_URL}/admin/v1/clients",
        headers={
            "Authorization": f"Bearer {ADMIN_KEY}",
            "Content-Type": "application/json"
        },
        json={
            "client_id": client_id,
            "api_key": api_key,
            "description": description,
            "allowed_providers": providers
        }
    )
    return response.json()

def get_client_metrics(client_id, hours=24):
    start = int(time.time()) - (hours * 3600)
    end = int(time.time())

    response = requests.get(
        f"{BASE_URL}/admin/v1/metrics/clients/{client_id}",
        headers={"Authorization": f"Bearer {ADMIN_KEY}"},
        params={"start": start, "end": end}
    )
    return response.json()
```

### Terraform Integration

```hcl
resource "http_request" "coo_llm_client" {
  url    = "http://coo-llm:2906/admin/v1/clients"
  method = "POST"

  headers = {
    Authorization = "Bearer ${var.admin_key}"
    Content-Type  = "application/json"
  }

  body = jsonencode({
    client_id         = "terraform-client"
    api_key           = random_password.api_key.result
    description       = "Managed by Terraform"
    allowed_providers = ["openai", "anthropic"]
  })
}

resource "random_password" "api_key" {
  length  = 32
  special = false
}
```