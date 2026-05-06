package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type NotificationRepository interface {
	Publish(ctx context.Context, channel string, message []byte) error
	Subscribe(ctx context.Context, channelPattern string) *redis.PubSub
	SaveHistory(ctx context.Context, recipientID string, title string, message string) error
}

type notificationRepository struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewNotificationRepository(db *pgxpool.Pool, redisClient *redis.Client) NotificationRepository {
	return &notificationRepository{
		db:    db,
		redis: redisClient,
	}
}

func (r *notificationRepository) Publish(ctx context.Context, channel string, message []byte) error {
	return r.redis.Publish(ctx, channel, message).Err()
}

func (r *notificationRepository) Subscribe(ctx context.Context, channelPattern string) *redis.PubSub {
	return r.redis.PSubscribe(ctx, channelPattern)
}

func (r *notificationRepository) SaveHistory(ctx context.Context, recipientID string, title string, message string) error {
	query := `INSERT INTO notification_history (recipient_id, title, message) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, recipientID, title, message)
	return err
}
