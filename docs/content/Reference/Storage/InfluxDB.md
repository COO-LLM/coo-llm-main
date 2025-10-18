---
sidebar_position: 4
tags: [reference, storage, influxdb, time-series]
---

# InfluxDB Storage

Time-series database optimized for high-performance metrics storage and analytics.

## Configuration

```yaml
storage:
  runtime:
    type: "influxdb"
    addr: "http://localhost:8086"
    token: "${INFLUX_TOKEN}"
    org: "${INFLUX_ORG}"      # Optional, default "coo-llm"
    bucket: "${INFLUX_BUCKET}" # Optional, default "coo-llm"
```

## Features

- **Time-Series Optimized**: High ingestion and query performance for timestamped data
- **Flux Query Language**: Powerful data processing and analytics
- **Continuous Queries**: Automated data aggregation and downsampling
- **Retention Policies**: Automatic data expiration and lifecycle management
- **High Availability**: Clustering support for production deployments
- **Downsampling**: Automatic data aggregation for long-term storage
- **Task Engine**: Scheduled data processing and alerting

## Data Structure

**Usage Metrics (measurement: usage):**
```
time                provider  key_id    metric  value
----                --------  ------    ------  -----
1640995200000000000 openai    key1      req     1
1640995200000000000 openai    key1      tokens  150
```

**Cache Data (measurement: cache):**
```
time                key       value
----                ---       -----
1640995200000000000 cache_key "cached_response"
```

## Implementation Details

- **Line Protocol**: Efficient data ingestion format
- **Flux Queries**: Advanced analytics and aggregations
- **Retention Policies**: Automated data lifecycle management
- **Authentication**: Token-based secure access
- **Batch Writing**: High-throughput data ingestion
- **Query Optimization**: Time-range and tag-based filtering
