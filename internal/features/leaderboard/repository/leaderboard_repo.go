package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type LeaderboardRepository interface {
	IncrementScore(ctx context.Context, partnerID string, amount float64) error
	GetTopRank(ctx context.Context, limit int64) ([]redis.Z, error)
	GetUserRank(ctx context.Context, partnerID string) (int64, float64, error)
	SavePointHistory(ctx context.Context, partnerID string, amount int, reason string) error
}

type leaderboardRepo struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewLeaderboardRepository(db *pgxpool.Pool, redisClient *redis.Client) LeaderboardRepository {
	return &leaderboardRepo{db: db, redis: redisClient}
}

func (r *leaderboardRepo) IncrementScore(ctx context.Context, partnerID string, amount float64) error {
	return r.redis.ZIncrBy(ctx, "leaderboard:partner", amount, partnerID).Err()
}

func (r *leaderboardRepo) GetTopRank(ctx context.Context, limit int64) ([]redis.Z, error) {
	return r.redis.ZRevRangeWithScores(ctx, "leaderboard:partner", 0, limit-1).Result()
}

func (r *leaderboardRepo) GetUserRank(ctx context.Context, partnerID string) (int64, float64, error) {
	rank, err := r.redis.ZRevRank(ctx, "leaderboard:partner", partnerID).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, 0, nil // Not found in leaderboard
		}
		return 0, 0, err
	}

	score, err := r.redis.ZScore(ctx, "leaderboard:partner", partnerID).Result()
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
