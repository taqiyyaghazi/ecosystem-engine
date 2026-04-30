package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/auth/entity"
)

type SessionRepository interface {
	CreateSession(ctx context.Context, sessionID string, user entity.User, userAgent, ipAddress string) error
	GetSession(ctx context.Context, sessionID string) (map[string]string, error)
	DeleteSession(ctx context.Context, sessionID string) error
	RefreshTTL(ctx context.Context, sessionID string) error
}

type sessionRepository struct {
	redis *redis.Client
}

func NewSessionRepository(redis *redis.Client) SessionRepository {
	return &sessionRepository{redis: redis}
}

func sessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

func (r *sessionRepository) CreateSession(ctx context.Context, sessionID string, user entity.User, userAgent, ipAddress string) error {
	key := sessionKey(sessionID)

	data := map[string]interface{}{
		"user_id":       user.ID.String(),
		"role":          user.Role,
		"user_agent":    userAgent,
		"ip_address":    ipAddress,
		"last_activity": time.Now().Format(time.RFC3339),
	}

	pipe := r.redis.TxPipeline()
	pipe.HSet(ctx, key, data)
	pipe.Expire(ctx, key, entity.SessionTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *sessionRepository) GetSession(ctx context.Context, sessionID string) (map[string]string, error) {
	return r.redis.HGetAll(ctx, sessionKey(sessionID)).Result()
}

func (r *sessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	return r.redis.Del(ctx, sessionKey(sessionID)).Err()
}

func (r *sessionRepository) RefreshTTL(ctx context.Context, sessionID string) error {
	return r.redis.Expire(ctx, sessionKey(sessionID), entity.SessionTTL).Err()
}
