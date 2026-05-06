package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type ActivityRepository interface {
	RecordEvent(ctx context.Context, event map[string]interface{}) error
	ReadEvents(ctx context.Context, consumerName string) ([]redis.XMessage, error)
	AcknowledgeEvent(ctx context.Context, messageID string) error
	InsertActivityLog(ctx context.Context, actorID, action string, payload map[string]interface{}, ipAddress string) error
	InitConsumerGroup(ctx context.Context) error
}

type activityRepository struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewActivityRepository(db *pgxpool.Pool, rdb *redis.Client) ActivityRepository {
	return &activityRepository{db: db, rdb: rdb}
}

func (r *activityRepository) RecordEvent(ctx context.Context, event map[string]interface{}) error {
	return r.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "stream:activity",
		MaxLen: 10000,
		Approx: true,
		Values: event,
	}).Err()
}

func (r *activityRepository) ReadEvents(ctx context.Context, consumerName string) ([]redis.XMessage, error) {
	streams, err := r.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    "cg:activity_processor",
		Consumer: consumerName,
		Streams:  []string{"stream:activity", ">"},
		Count:    10,
		Block:    0,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return nil, nil // No messages
		}
		return nil, err
	}

	if len(streams) > 0 {
		return streams[0].Messages, nil
	}

	return nil, nil
}

func (r *activityRepository) AcknowledgeEvent(ctx context.Context, messageID string) error {
	return r.rdb.XAck(ctx, "stream:activity", "cg:activity_processor", messageID).Err()
}

func (r *activityRepository) InsertActivityLog(ctx context.Context, actorID, action string, payload map[string]interface{}, ipAddress string) error {
	var payloadBytes []byte
	if payload != nil {
		var err error
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	} else {
		payloadBytes = []byte("{}")
	}

	query := `
		INSERT INTO activity_logs (actor_id, action, payload, ip_address)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(ctx, query, actorID, action, payloadBytes, ipAddress)
	return err
}

func (r *activityRepository) InitConsumerGroup(ctx context.Context) error {
	err := r.rdb.XGroupCreateMkStream(ctx, "stream:activity", "cg:activity_processor", "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}
	return nil
}
