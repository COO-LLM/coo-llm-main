# Deployment

---
sidebar_position: 3
tags: [user-guide, deployment]
---

This guide covers deploying TruckLLM in various environments, from development to production.

## Quick Start

### Local Development

1. **Clone and build:**
   ```bash
   git clone https://github.com/your-org/truckllm.git
   cd truckllm
   go build -o bin/truckllm ./cmd/truckllm
   ```

2. **Configure environment:**
   ```bash
   export OPENAI_API_KEY="sk-your-key"
   export GEMINI_API_KEY="your-gemini-key"
   ```

3. **Create config:**
   ```yaml
   # configs/config.yaml
   version: "1.0"
   server:
     listen: ":8080"
   providers:
     - id: openai
       base_url: "https://api.openai.com/v1"
       keys:
         - secret: "${OPENAI_API_KEY}"
   model_aliases:
     gpt-4: openai:gpt-4
   ```

4. **Run:**
   ```bash
   ./bin/truckllm -config configs/config.yaml
   ```

5. **Test:**
   ```bash
   curl -X POST http://localhost:8080/v1/chat/completions \
     -H "Authorization: Bearer test" \
     -d '{"model": "gpt-4", "messages": [{"role": "user", "content": "Hello"}]}'
   ```

### Docker Deployment

1. **Build image:**
   ```bash
   docker build -t truckllm:latest .
   ```

2. **Run container:**
   ```bash
   docker run -p 8080:8080 \
     -e OPENAI_API_KEY="sk-your-key" \
     -v $(pwd)/configs:/app/configs \
     truckllm:latest
   ```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  truckllm:
    image: truckllm:latest
    ports:
      - "8080:8080"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - GEMINI_API_KEY=${GEMINI_API_KEY}
    volumes:
      - ./configs:/app/configs
      - ./logs:/app/logs
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  redis_data:
```

**Run:**
```bash
docker-compose up -d
```

## Production Deployment

### Kubernetes

**Deployment:**
```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: truckllm
spec:
  replicas: 3
  selector:
    matchLabels:
      app: truckllm
  template:
    metadata:
      labels:
        app: truckllm
    spec:
      containers:
      - name: truckllm
        image: truckllm:latest
        ports:
        - containerPort: 8080
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: llm-secrets
              key: openai-key
        volumeMounts:
        - name: config
          mountPath: /app/configs
      volumes:
      - name: config
        configMap:
          name: truckllm-config
---
apiVersion: v1
kind: Service
metadata:
  name: truckllm
spec:
  selector:
    app: truckllm
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

**ConfigMap:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: truckllm-config
data:
  config.yaml: |
    version: "1.0"
    server:
      listen: ":8080"
    providers:
      - id: openai
        base_url: "https://api.openai.com/v1"
        keys:
          - secret: "${OPENAI_API_KEY}"
    storage:
      runtime:
        type: redis
        addr: redis-service:6379
```

**Apply:**
```bash
kubectl apply -f k8s/
```

### AWS ECS

**Task Definition:**
```json
{
  "family": "truckllm",
  "taskRoleArn": "arn:aws:iam::123456789012:role/ecsTaskRole",
  "executionRoleArn": "arn:aws:iam::123456789012:role/ecsTaskExecutionRole",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "containerDefinitions": [
    {
      "name": "truckllm",
      "image": "truckllm:latest",
      "essential": true,
      "portMappings": [
        {
          "containerPort": 8080,
          "hostPort": 8080
        }
      ],
      "environment": [
        {
          "name": "OPENAI_API_KEY",
          "valueFrom": "arn:aws:secretsmanager:region:123456789012:secret:openai-key"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/truckllm",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

### Docker Swarm

```yaml
# docker-compose.swarm.yml
version: '3.8'
services:
  truckllm:
    image: truckllm:latest
    ports:
      - "8080:8080"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    configs:
      - source: truckllm_config
        target: /app/configs/config.yaml
    deploy:
      mode: replicated
      replicas: 3
      restart_policy:
        condition: on-failure

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    deploy:
      mode: global

configs:
  truckllm_config:
    file: ./configs/config.yaml

volumes:
  redis_data:
```

## Configuration Management

### Environment Variables

**Development:**
```bash
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="..."
export CLAUDE_API_KEY="..."
export REDIS_URL="redis://localhost:6379"
```

**Production:**
Use secret management services:
- AWS Secrets Manager
- Google Secret Manager
- Azure Key Vault
- HashiCorp Vault

### Config Files

**Directory Structure:**
```
configs/
├── config.yaml          # Main config
├── config.prod.yaml     # Production overrides
├── config.dev.yaml      # Development overrides
└── secrets/             # Encrypted secrets
```

**Environment-specific configs:**
```bash
# Development
./truckllm -config configs/config.dev.yaml

# Production
./truckllm -config configs/config.prod.yaml
```

## Networking

### Load Balancing

**nginx:**
```nginx
upstream truckllm {
    server truckllm-1:8080;
    server truckllm-2:8080;
    server truckllm-3:8080;
}

server {
    listen 80;
    server_name api.example.com;

    location / {
        proxy_pass http://truckllm;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

**AWS ALB:**
- Target Group: EC2 instances or ECS tasks
- Health Check: `GET /health`
- SSL Termination: ACM certificate

### Security

**TLS Configuration:**
```yaml
server:
  listen: ":8443"
  tls:
    cert_file: "/etc/ssl/certs/truckllm.crt"
    key_file: "/etc/ssl/private/truckllm.key"
```

**nginx with TLS:**
```nginx
server {
    listen 443 ssl;
    server_name api.example.com;

    ssl_certificate /etc/ssl/certs/api.crt;
    ssl_certificate_key /etc/ssl/private/api.key;

    location / {
        proxy_pass http://truckllm;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Monitoring

### Health Checks

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "24h30m45s",
  "checks": {
    "redis": "ok",
    "providers": "ok"
  }
}
```

**Kubernetes Probe:**
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

### Metrics

**Prometheus:**
```yaml
scrape_configs:
  - job_name: 'truckllm'
    static_configs:
      - targets: ['truckllm:8080']
    metrics_path: '/metrics'
```

**Grafana Dashboard:**
Import dashboard from `monitoring/grafana-dashboard.json`

### Logging

**Centralized Logging:**
```yaml
logging:
  providers:
    - name: "elasticsearch"
      type: "http"
      endpoint: "https://es.example.com/_bulk"
```

**Log Aggregation:**
- ELK Stack (Elasticsearch, Logstash, Kibana)
- Splunk
- Datadog
- CloudWatch Logs

## Scaling

### Horizontal Scaling

**Kubernetes HPA:**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: truckllm-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: truckllm
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Vertical Scaling

**Resource Limits:**
```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### Database Scaling

**Redis Cluster:**
```yaml
storage:
  runtime:
    type: redis
    addr: "redis-cluster:6379"
    cluster: true
```

## Backup & Recovery

### Configuration Backup

```bash
# Backup configs
tar -czf configs-backup-$(date +%Y%m%d).tar.gz configs/

# Restore
tar -xzf configs-backup-20240101.tar.gz
```

### Data Backup

**Redis Backup:**
```bash
# Create snapshot
redis-cli save

# Copy RDB file
docker cp redis:/data/dump.rdb ./backup/
```

### Disaster Recovery

1. **Restore from backup:**
   ```bash
   docker run -d --name redis-restore -v ./backup:/data redis:7-alpine
   ```

2. **Update configuration:**
   ```yaml
   storage:
     runtime:
       type: redis
       addr: redis-restore:6379
   ```

3. **Gradual rollout:**
   ```bash
   kubectl rollout restart deployment/truckllm
   ```

## Troubleshooting

### Common Issues

**Container won't start:**
```bash
docker logs truckllm
# Check for config errors or missing env vars
```

**High latency:**
```bash
# Check provider status
curl http://localhost:8080/admin/v1/providers

# Check Redis connection
redis-cli ping
```

**Rate limiting:**
```bash
# Monitor usage
curl http://localhost:8080/admin/v1/providers | jq '.providers[].keys[]'
```

### Debug Mode

```bash
# Enable debug logging
export LOG_LEVEL=debug

# Run with verbose output
./truckllm -config config.yaml -verbose
```

### Performance Tuning

**Go flags:**
```bash
export GOGC=100
export GOMAXPROCS=4
```

**Redis optimization:**
```yaml
storage:
  runtime:
    pool_size: 20
    min_idle_conns: 5
```

## CI/CD

### GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy
on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - run: go test ./...

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: docker/build-push-action@v4
      with:
        push: true
        tags: truckllm:latest

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
    - uses: azure/k8s-deploy@v4
      with:
        manifests: k8s/
        images: truckllm:latest
```

### ArgoCD

```yaml
# argocd/application.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: truckllm
spec:
  project: default
  source:
    repoURL: https://github.com/your-org/truckllm
    path: k8s
    targetRevision: HEAD
  destination:
    server: https://kubernetes.default.svc
    namespace: default
```

## Security

### Secrets Management

**Kubernetes Secrets:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: llm-secrets
type: Opaque
data:
  openai-key: <base64-encoded-key>
```

**AWS Secrets:**
```bash
aws secretsmanager create-secret \
  --name truckllm/openai-key \
  --secret-string "sk-your-key"
```

### Network Security

**Security Groups:**
- Allow inbound on port 8080 from load balancer
- Restrict Redis access to internal network
- Enable VPC flow logs

**Pod Security:**
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  readOnlyRootFilesystem: true
```

### Compliance

**GDPR:**
- Data minimization in logs
- Right to erasure implementation
- Data processing agreements

**SOC 2:**
- Access logging
- Change management
- Incident response procedures