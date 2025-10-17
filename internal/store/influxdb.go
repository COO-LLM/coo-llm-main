package store

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rs/zerolog"
)

type InfluxDBStore struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	queryAPI api.QueryAPI
	org      string
	bucket   string
	logger   zerolog.Logger
}

func NewInfluxDBStore(url, token, org, bucket string, logger zerolog.Logger) *InfluxDBStore {
	client := influxdb2.NewClient(url, token)
	writeAPI := client.WriteAPIBlocking(org, bucket)
	queryAPI := client.QueryAPI(org)

	return &InfluxDBStore{
		client:   client,
		writeAPI: writeAPI,
		queryAPI: queryAPI,
		org:      org,
		bucket:   bucket,
		logger:   logger,
	}
}

func (i *InfluxDBStore) GetUsage(provider, keyID, metric string) (float64, error) {
	// For real-time usage, query latest value
	query := fmt.Sprintf(`from(bucket: "%s")
		|> range(start: -1h)
		|> filter(fn: (r) => r._measurement == "usage" and r.provider == "%s" and r.keyID == "%s" and r.metric == "%s")
		|> last()`, i.bucket, provider, keyID, metric)

	result, err := i.queryAPI.Query(context.Background(), query)
	if err != nil {
		return 0, err
	}

	if result.Next() {
		return result.Record().Value().(float64), nil
	}

	return 0, nil
}

func (i *InfluxDBStore) SetUsage(provider, keyID, metric string, value float64) error {
	point := influxdb2.NewPointWithMeasurement("usage").
		AddTag("provider", provider).
		AddTag("keyID", keyID).
		AddTag("metric", metric).
		AddField("value", value).
		SetTime(time.Now())

	return i.writeAPI.WritePoint(context.Background(), point)
}

func (i *InfluxDBStore) IncrementUsage(provider, keyID, metric string, delta float64) error {
	current, err := i.GetUsage(provider, keyID, metric)
	if err != nil {
		return err
	}
	return i.SetUsage(provider, keyID, metric, current+delta)
}

func (i *InfluxDBStore) GetUsageInWindow(provider, keyID, metric string, windowSeconds int64) (float64, error) {
	// Sum over the window
	query := fmt.Sprintf(`from(bucket: "%s")
		|> range(start: -%ds)
		|> filter(fn: (r) => r._measurement == "usage" and r.provider == "%s" and r.keyID == "%s" and r.metric == "%s")
		|> sum()`, i.bucket, windowSeconds, provider, keyID, metric)

	result, err := i.queryAPI.Query(context.Background(), query)
	if err != nil {
		return 0, err
	}

	if result.Next() {
		return result.Record().Value().(float64), nil
	}

	return 0, nil
}

func (i *InfluxDBStore) SetCache(key, value string, ttlSeconds int64) error {
	// Use InfluxDB for cache with TTL (InfluxDB handles retention)
	point := influxdb2.NewPointWithMeasurement("cache").
		AddTag("key", key).
		AddField("value", value).
		SetTime(time.Now())

	return i.writeAPI.WritePoint(context.Background(), point)
}

func (i *InfluxDBStore) GetCache(key string) (string, error) {
	query := fmt.Sprintf(`from(bucket: "%s")
		|> range(start: -1h)
		|> filter(fn: (r) => r._measurement == "cache" and r.key == "%s")
		|> last()`, i.bucket, key)

	result, err := i.queryAPI.Query(context.Background(), query)
	if err != nil {
		return "", err
	}

	if result.Next() {
		return result.Record().ValueByKey("value").(string), nil
	}

	return "", nil
}

func (i *InfluxDBStore) StoreMetric(name string, value float64, tags map[string]string, timestamp int64) error {
	point := influxdb2.NewPointWithMeasurement(name).
		SetTime(time.Unix(timestamp, 0))

	for k, v := range tags {
		point = point.AddTag(k, v)
	}

	point = point.AddField("value", value)

	return i.writeAPI.WritePoint(context.Background(), point)
}

func (i *InfluxDBStore) GetMetrics(name string, tags map[string]string, start, end int64) ([]MetricPoint, error) {
	query := fmt.Sprintf(`from(bucket: "%s")
		|> range(start: %d, stop: %d)
		|> filter(fn: (r) => r._measurement == "%s"`, i.bucket, start, end, name)

	for k, v := range tags {
		query += fmt.Sprintf(` and r.%s == "%s"`, k, v)
	}

	query += `)`

	result, err := i.queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var points []MetricPoint
	for result.Next() {
		record := result.Record()
		tags := make(map[string]string)
		values := record.Values()
		for k, v := range values {
			if k != "_value" && k != "_time" && k != "_measurement" {
				if str, ok := v.(string); ok {
					tags[k] = str
				}
			}
		}
		points = append(points, MetricPoint{
			Value:     record.Value().(float64),
			Timestamp: record.Time().Unix(),
			Tags:      tags,
		})
	}

	return points, nil
}

func (i *InfluxDBStore) Close() {
	i.client.Close()
}
