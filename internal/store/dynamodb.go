package store

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
)

type DynamoDBStore struct {
	client       *dynamodb.Client
	logger       zerolog.Logger
	tableUsage   string
	tableCache   string
	tableHistory string
}

func NewDynamoDBStore(region, tableUsage, tableCache, tableHistory string, logger zerolog.Logger) (*DynamoDBStore, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(cfg)

	// Test connection by describing tables
	tables := []string{tableUsage, tableCache, tableHistory}
	for _, table := range tables {
		_, err = client.DescribeTable(context.Background(), &dynamodb.DescribeTableInput{
			TableName: aws.String(table),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to DynamoDB table %s: %w", table, err)
		}
	}

	return &DynamoDBStore{
		client:       client,
		logger:       logger,
		tableUsage:   tableUsage,
		tableCache:   tableCache,
		tableHistory: tableHistory,
	}, nil
}

func (d *DynamoDBStore) getUsageKey(provider, keyID, metric string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USAGE#%s#%s", provider, keyID)},
		"sk": &types.AttributeValueMemberS{Value: metric},
	}
}

func (d *DynamoDBStore) getCacheKey(key string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("CACHE#%s", key)},
		"sk": &types.AttributeValueMemberS{Value: "DATA"},
	}
}

func (d *DynamoDBStore) GetUsage(provider, keyID, metric string) (float64, error) {
	ctx := context.Background()

	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableUsage),
		Key:       d.getUsageKey(provider, keyID, metric),
	})

	if err != nil {
		d.logger.Error().Err(err).Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Msg("store operation failed")
		return 0, err
	}

	if result.Item == nil {
		d.logger.Debug().Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", 0).Msg("store operation - item not found")
		return 0, nil
	}

	valueAttr, ok := result.Item["value"]
	if !ok {
		d.logger.Debug().Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", 0).Msg("store operation - value attribute not found")
		return 0, nil
	}

	valueStr, ok := valueAttr.(*types.AttributeValueMemberN)
	if !ok {
		d.logger.Error().Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Msg("store operation - invalid value type")
		return 0, fmt.Errorf("invalid value type")
	}

	value, err := strconv.ParseFloat(valueStr.Value, 64)
	if err != nil {
		d.logger.Error().Err(err).Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Str("valueStr", valueStr.Value).Msg("store operation - parse float failed")
		return 0, err
	}

	d.logger.Debug().Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", value).Msg("store operation")
	return value, nil
}

func (d *DynamoDBStore) SetUsage(provider, keyID, metric string, value float64) error {
	ctx := context.Background()

	item := map[string]types.AttributeValue{
		"pk":    &types.AttributeValueMemberS{Value: fmt.Sprintf("USAGE#%s#%s", provider, keyID)},
		"sk":    &types.AttributeValueMemberS{Value: metric},
		"value": &types.AttributeValueMemberN{Value: fmt.Sprintf("%.6f", value)},
	}

	_, err := d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.tableUsage),
		Item:      item,
	})

	if err != nil {
		d.logger.Error().Err(err).Str("operation", "SetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", value).Msg("store operation failed")
		return err
	}

	d.logger.Debug().Str("operation", "SetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", value).Msg("store operation")
	return nil
}

func (d *DynamoDBStore) IncrementUsage(provider, keyID, metric string, delta float64) error {
	ctx := context.Background()

	// Insert into history
	timestamp := time.Now().Unix()
	historyItem := map[string]types.AttributeValue{
		"pk":        &types.AttributeValueMemberS{Value: fmt.Sprintf("HISTORY#%s#%s#%s", provider, keyID, metric)},
		"sk":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", timestamp)},
		"delta":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%.6f", delta)},
		"timestamp": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", timestamp)},
	}

	_, err := d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.tableHistory),
		Item:      historyItem,
	})
	if err != nil {
		d.logger.Error().Err(err).Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Msg("insert history failed")
		return err
	}

	// Update total usage
	key := d.getUsageKey(provider, keyID, metric)

	updateExpr := "ADD #v :delta"
	exprAttrNames := map[string]string{
		"#v": "value",
	}
	exprAttrValues := map[string]types.AttributeValue{
		":delta": &types.AttributeValueMemberN{Value: fmt.Sprintf("%.6f", delta)},
	}

	_, err = d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(d.tableUsage),
		Key:                       key,
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  exprAttrNames,
		ExpressionAttributeValues: exprAttrValues,
	})

	if err != nil {
		d.logger.Error().Err(err).Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Msg("update total failed")
		return err
	}

	d.logger.Debug().Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Msg("store operation")
	return nil
}

func (d *DynamoDBStore) GetUsageInWindow(provider, keyID, metric string, windowSeconds int64) (float64, error) {
	ctx := context.Background()

	now := time.Now().Unix()
	start := now - windowSeconds

	// Query history table for items in time window
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(d.tableHistory),
		KeyConditionExpression: aws.String("pk = :pk AND sk BETWEEN :start AND :end"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":    &types.AttributeValueMemberS{Value: fmt.Sprintf("HISTORY#%s#%s#%s", provider, keyID, metric)},
			":start": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", start)},
			":end":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", now)},
		},
	}

	result, err := d.client.Query(ctx, queryInput)
	if err != nil {
		d.logger.Error().Err(err).Str("operation", "GetUsageInWindow").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Int64("windowSeconds", windowSeconds).Msg("store operation failed")
		return 0, err
	}

	total := 0.0
	for _, item := range result.Items {
		if deltaAttr, ok := item["delta"]; ok {
			if deltaStr, ok := deltaAttr.(*types.AttributeValueMemberN); ok {
				if delta, err := strconv.ParseFloat(deltaStr.Value, 64); err == nil {
					total += delta
				}
			}
		}
	}

	d.logger.Debug().Str("operation", "GetUsageInWindow").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Int64("windowSeconds", windowSeconds).Float64("total", total).Msg("store operation")
	return total, nil
}

func (d *DynamoDBStore) SetCache(key, value string, ttlSeconds int64) error {
	ctx := context.Background()

	expiry := time.Now().Add(time.Duration(ttlSeconds) * time.Second).Unix()

	item := map[string]types.AttributeValue{
		"pk":     &types.AttributeValueMemberS{Value: fmt.Sprintf("CACHE#%s", key)},
		"sk":     &types.AttributeValueMemberS{Value: "DATA"},
		"value":  &types.AttributeValueMemberS{Value: value},
		"expiry": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", expiry)},
	}

	_, err := d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.tableCache),
		Item:      item,
	})

	if err != nil {
		d.logger.Error().Err(err).Str("operation", "SetCache").Str("key", key).Int64("ttlSeconds", ttlSeconds).Msg("store operation failed")
		return err
	}

	d.logger.Debug().Str("operation", "SetCache").Str("key", key).Int64("ttlSeconds", ttlSeconds).Msg("store operation")
	return nil
}

func (d *DynamoDBStore) GetCache(key string) (string, error) {
	ctx := context.Background()

	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableCache),
		Key:       d.getCacheKey(key),
	})

	if err != nil {
		d.logger.Error().Err(err).Str("operation", "GetCache").Str("key", key).Msg("store operation failed")
		return "", err
	}

	if result.Item == nil {
		d.logger.Debug().Str("operation", "GetCache").Str("key", key).Msg("store operation - cache miss")
		return "", nil
	}

	expiryAttr, hasExpiry := result.Item["expiry"]
	if hasExpiry {
		expiryStr, ok := expiryAttr.(*types.AttributeValueMemberN)
		if ok {
			expiry, err := strconv.ParseInt(expiryStr.Value, 10, 64)
			if err == nil && time.Now().Unix() > expiry {
				d.logger.Debug().Str("operation", "GetCache").Str("key", key).Msg("store operation - cache expired")
				return "", nil
			}
		}
	}

	valueAttr, ok := result.Item["value"]
	if !ok {
		d.logger.Debug().Str("operation", "GetCache").Str("key", key).Msg("store operation - value attribute not found")
		return "", nil
	}

	valueStr, ok := valueAttr.(*types.AttributeValueMemberS)
	if !ok {
		d.logger.Error().Str("operation", "GetCache").Str("key", key).Msg("store operation - invalid value type")
		return "", fmt.Errorf("invalid value type")
	}

	d.logger.Debug().Str("operation", "GetCache").Str("key", key).Msg("store operation - cache hit")
	return valueStr.Value, nil
}
