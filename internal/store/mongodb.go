package store

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBStore struct {
	client   *mongo.Client
	database *mongo.Database
	logger   zerolog.Logger
}

type UsageDocument struct {
	Provider string  `bson:"provider"`
	KeyID    string  `bson:"key_id"`
	Metric   string  `bson:"metric"`
	Value    float64 `bson:"value"`
}

type UsageHistoryDocument struct {
	Provider  string    `bson:"provider"`
	KeyID     string    `bson:"key_id"`
	Metric    string    `bson:"metric"`
	Delta     float64   `bson:"delta"`
	Timestamp time.Time `bson:"timestamp"`
}

type CacheDocument struct {
	Key    string    `bson:"_id"`
	Value  string    `bson:"value"`
	Expiry time.Time `bson:"expiry,omitempty"`
}

func NewMongoDBStore(uri, database string, logger zerolog.Logger) (*MongoDBStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(database)

	// Create indexes
	if err := createMongoIndexes(db); err != nil {
		return nil, err
	}

	return &MongoDBStore{
		client:   client,
		database: db,
		logger:   logger,
	}, nil
}

func createMongoIndexes(db *mongo.Database) error {
	ctx := context.Background()

	// Usage metrics index
	_, err := db.Collection("usage_metrics").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "provider", Value: 1},
			{Key: "key_id", Value: 1},
			{Key: "metric", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	// Usage history indexes
	_, err = db.Collection("usage_history").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "timestamp", Value: 1}},
	})
	if err != nil {
		return err
	}

	_, err = db.Collection("usage_history").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "provider", Value: 1},
			{Key: "key_id", Value: 1},
			{Key: "metric", Value: 1},
		},
	})
	if err != nil {
		return err
	}

	// Cache index
	_, err = db.Collection("cache").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "expiry", Value: 1}},
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *MongoDBStore) GetUsage(provider, keyID, metric string) (float64, error) {
	ctx := context.Background()
	collection := m.database.Collection("usage_metrics")

	filter := bson.M{
		"provider": provider,
		"key_id":   keyID,
		"metric":   metric,
	}

	var doc UsageDocument
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		m.logger.Debug().Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", 0).Msg("store operation - no documents")
		return 0, nil
	}
	if err != nil {
		m.logger.Error().Err(err).Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Msg("store operation failed")
		return 0, err
	}

	m.logger.Debug().Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", doc.Value).Msg("store operation")
	return doc.Value, nil
}

func (m *MongoDBStore) SetUsage(provider, keyID, metric string, value float64) error {
	ctx := context.Background()
	collection := m.database.Collection("usage_metrics")

	filter := bson.M{
		"provider": provider,
		"key_id":   keyID,
		"metric":   metric,
	}

	update := bson.M{
		"$set": bson.M{
			"provider": provider,
			"key_id":   keyID,
			"metric":   metric,
			"value":    value,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		m.logger.Error().Err(err).Str("operation", "SetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", value).Msg("store operation failed")
		return err
	}

	m.logger.Debug().Str("operation", "SetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", value).Msg("store operation")
	return nil
}

func (m *MongoDBStore) IncrementUsage(provider, keyID, metric string, delta float64) error {
	ctx := context.Background()

	// Insert into history
	historyCollection := m.database.Collection("usage_history")
	historyDoc := UsageHistoryDocument{
		Provider:  provider,
		KeyID:     keyID,
		Metric:    metric,
		Delta:     delta,
		Timestamp: time.Now(),
	}

	_, err := historyCollection.InsertOne(ctx, historyDoc)
	if err != nil {
		m.logger.Error().Err(err).Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Msg("insert history failed")
		return err
	}

	// Update total
	metricsCollection := m.database.Collection("usage_metrics")
	filter := bson.M{
		"provider": provider,
		"key_id":   keyID,
		"metric":   metric,
	}

	update := bson.M{
		"$inc": bson.M{"value": delta},
		"$setOnInsert": bson.M{
			"provider": provider,
			"key_id":   keyID,
			"metric":   metric,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err = metricsCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		m.logger.Error().Err(err).Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Msg("update total failed")
		return err
	}

	m.logger.Debug().Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Msg("store operation")
	return nil
}

func (m *MongoDBStore) GetUsageInWindow(provider, keyID, metric string, windowSeconds int64) (float64, error) {
	ctx := context.Background()
	collection := m.database.Collection("usage_history")

	filter := bson.M{
		"provider":  provider,
		"key_id":    keyID,
		"metric":    metric,
		"timestamp": bson.M{"$gt": time.Now().Add(-time.Duration(windowSeconds) * time.Second)},
	}

	pipeline := mongo.Pipeline{
		{primitive.E{Key: "$match", Value: filter}},
		{primitive.E{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$delta"},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		m.logger.Error().Err(err).Str("operation", "GetUsageInWindow").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Int64("windowSeconds", windowSeconds).Msg("store operation failed")
		return 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		Total float64 `bson:"total"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			m.logger.Error().Err(err).Str("operation", "GetUsageInWindow").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Int64("windowSeconds", windowSeconds).Msg("decode failed")
			return 0, err
		}
	}

	m.logger.Debug().Str("operation", "GetUsageInWindow").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Int64("windowSeconds", windowSeconds).Float64("total", result.Total).Msg("store operation")
	return result.Total, nil
}

func (m *MongoDBStore) SetCache(key, value string, ttlSeconds int64) error {
	ctx := context.Background()
	collection := m.database.Collection("cache")

	expiry := time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	doc := CacheDocument{
		Key:    key,
		Value:  value,
		Expiry: expiry,
	}

	filter := bson.M{"_id": key}
	update := bson.M{"$set": doc}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		m.logger.Error().Err(err).Str("operation", "SetCache").Str("key", key).Int64("ttlSeconds", ttlSeconds).Msg("store operation failed")
		return err
	}

	m.logger.Debug().Str("operation", "SetCache").Str("key", key).Int64("ttlSeconds", ttlSeconds).Msg("store operation")
	return nil
}

func (m *MongoDBStore) GetCache(key string) (string, error) {
	ctx := context.Background()
	collection := m.database.Collection("cache")

	filter := bson.M{
		"_id": key,
		"$or": []bson.M{
			{"expiry": bson.M{"$exists": false}},
			{"expiry": bson.M{"$gt": time.Now()}},
		},
	}

	var doc CacheDocument
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		m.logger.Debug().Str("operation", "GetCache").Str("key", key).Msg("store operation - cache miss")
		return "", nil
	}
	if err != nil {
		m.logger.Error().Err(err).Str("operation", "GetCache").Str("key", key).Msg("store operation failed")
		return "", err
	}

	m.logger.Debug().Str("operation", "GetCache").Str("key", key).Msg("store operation - cache hit")
	return doc.Value, nil
}
