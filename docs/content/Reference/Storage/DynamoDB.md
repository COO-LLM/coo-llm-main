---
sidebar_position: 3
tags: [reference, storage, dynamodb, aws]
---

# DynamoDB Storage

Serverless NoSQL database service for scalable COO-LLM data storage on AWS.

## Configuration

```yaml
storage:
  runtime:
    type: "dynamodb"
    addr: "us-east-1"  # AWS region
    table_usage: "coo-llm-usage"
    table_cache: "coo-llm-cache" 
    table_history: "coo-llm-history"
    access_key: "${AWS_ACCESS_KEY_ID}"    # Optional, uses IAM roles
    secret_key: "${AWS_SECRET_ACCESS_KEY}" # Optional, uses IAM roles
```

## Features

- **Serverless**: No server management or capacity planning
- **Auto Scaling**: Automatic throughput scaling based on demand
- **Global Tables**: Multi-region replication for low-latency access
- **Point-in-Time Recovery**: Continuous backups with 35-day retention
- **Streams**: Real-time change data capture
- **DynamoDB Accelerator (DAX)**: In-memory caching for microsecond response times
- **On-Demand Pricing**: Pay-per-request pricing model

## Data Structure

**Usage Table:**
```
Primary Key: provider (HASH), key_id + metric (RANGE)
Attributes: value (Number), updated_at (String)
```

**History Table:**
```
Primary Key: provider (HASH), timestamp (RANGE)
Attributes: key_id, metric, delta
Global Secondary Index: provider + key_id + metric (RANGE: timestamp)
```

**Cache Table:**
```
Primary Key: cache_key (HASH)
Attributes: value (String), expiry (Number)
TTL enabled on expiry attribute
```

## Implementation Details

- **AWS SDK Integration**: Native DynamoDB API usage
- **Batch Operations**: Efficient bulk read/write operations
- **Conditional Updates**: Atomic operations with conditions
- **Error Handling**: Automatic retry with exponential backoff
- **Cost Optimization**: On-demand capacity mode
- **IAM Integration**: Secure credential management
