package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
)

type LeaderboardRepository interface {
	IncrementScore(ctx context.Context, serviceID, partnerID string, amount float64) error
	GetTopRank(ctx context.Context, serviceID string, limit int64) ([]redis.Z, error)
	GetUserRank(ctx context.Context, serviceID, partnerID string) (int64, float64, error)
	SavePointHistory(ctx context.Context, partnerID string, amount int, reason string) error
	GetPartnerIDByUserAndService(ctx context.Context, userID, serviceID string) (string, error)
	GetServiceIDByPartnerID(ctx context.Context, partnerID string) (string, error)
}

type leaderboardRepo struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewLeaderboardRepository(db *pgxpool.Pool, redisClient *redis.Client) LeaderboardRepository {
	return &leaderboardRepo{db: db, redis: redisClient}
}

func redisKey(serviceID string) string {
	return fmt.Sprintf("leaderboard:service:%s", serviceID)
}

func (r *leaderboardRepo) IncrementScore(ctx context.Context, serviceID, partnerID string, amount float64) error {
	return r.redis.ZIncrBy(ctx, redisKey(serviceID), amount, partnerID).Err()
}

func (r *leaderboardRepo) GetTopRank(ctx context.Context, serviceID string, limit int64) ([]redis.Z, error) {
	return r.redis.ZRevRangeWithScores(ctx, redisKey(serviceID), 0, limit-1).Result()
}

func (r *leaderboardRepo) GetUserRank(ctx context.Context, serviceID, partnerID string) (int64, float64, error) {
	key := redisKey(serviceID)

	rank, err := r.redis.ZRevRank(ctx, key, partnerID).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, 0, nil // not yet on the leaderboard
		}
		return 0, 0, err
	}

	score, err := r.redis.ZScore(ctx, key, partnerID).Result()
	if err != nil && err != redis.Nil {
		return 0, 0, err
	}

	return rank + 1, score, nil
}

func (r *leaderboardRepo) SavePointHistory(ctx context.Context, partnerID string, amount int, reason string) error {
	query := `INSERT INTO point_history (partner_id, amount, reason) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, partnerID, amount, reason)
	return err
}

func (r *leaderboardRepo) GetPartnerIDByUserAndService(ctx context.Context, userID, serviceID string) (string, error) {
	var partnerID string
	query := `SELECT id FROM partners WHERE user_id = $1 AND service_id = $2`
	err := r.db.QueryRow(ctx, query, userID, serviceID).Scan(&partnerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", apperror.ErrNotFound
		}
		return "", err
	}
	return partnerID, nil
}

func (r *leaderboardRepo) GetServiceIDByPartnerID(ctx context.Context, partnerID string) (string, error) {
	var serviceID string
	query := `SELECT service_id FROM partners WHERE id = $1`
	err := r.db.QueryRow(ctx, query, partnerID).Scan(&serviceID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", apperror.ErrNotFound
		}
		return "", err
	}
	return serviceID, nil
}
