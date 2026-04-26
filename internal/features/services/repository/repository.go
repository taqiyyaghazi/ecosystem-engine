package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/entity"
)

type ServiceRepository interface {
	Create(ctx context.Context, service *entity.Service) error
	GetAll(ctx context.Context) ([]entity.Service, error)
	GetByID(ctx context.Context, id string) (*entity.Service, error)
	Update(ctx context.Context, service *entity.Service) error
	Delete(ctx context.Context, id string) error
	InvalidateCache(ctx context.Context, id string)
}

type serviceRepository struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewServiceRepository(db *pgxpool.Pool, redis *redis.Client) ServiceRepository {
	return &serviceRepository{
		db:    db,
		redis: redis,
	}
}

const (
	cacheKeyAll    = "services:all"
	cacheKeyDetail = "services:id:%s"
	cacheTTL       = 1 * time.Hour
)

func (r *serviceRepository) Create(ctx context.Context, s *entity.Service) error {
	query := `INSERT INTO services (name, description, price) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, s.Name, s.Description, s.Price).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return err
	}

	r.InvalidateCache(ctx, "")
	return nil
}

func (r *serviceRepository) GetAll(ctx context.Context) ([]entity.Service, error) {
	// 1. Try get from Redis
	val, err := r.redis.Get(ctx, cacheKeyAll).Result()
	if err == nil {
		var services []entity.Service
		if err := json.Unmarshal([]byte(val), &services); err == nil {
			slog.DebugContext(ctx, "Cache HIT", "key", cacheKeyAll)
			return services, nil
		}
	}

	slog.DebugContext(ctx, "Cache MISS", "key", cacheKeyAll)

	// 2. Fallback to DB
	query := `SELECT id, name, description, price, created_at, updated_at FROM services ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []entity.Service
	for rows.Next() {
		var s entity.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Price, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		services = append(services, s)
	}

	// 3. Set to Redis
	if data, err := json.Marshal(services); err == nil {
		r.redis.Set(ctx, cacheKeyAll, data, cacheTTL)
	}

	return services, nil
}

func (r *serviceRepository) GetByID(ctx context.Context, id string) (*entity.Service, error) {
	key := fmt.Sprintf(cacheKeyDetail, id)

	// 1. Try get from Redis
	val, err := r.redis.Get(ctx, key).Result()
	if err == nil {
		var s entity.Service
		if err := json.Unmarshal([]byte(val), &s); err == nil {
			slog.DebugContext(ctx, "Cache HIT", "key", key)
			return &s, nil
		}
	}

	slog.DebugContext(ctx, "Cache MISS", "key", key)

	// 2. Fallback to DB
	var s entity.Service
	query := `SELECT id, name, description, price, created_at, updated_at FROM services WHERE id = $1`
	err = r.db.QueryRow(ctx, query, id).Scan(&s.ID, &s.Name, &s.Description, &s.Price, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// 3. Set to Redis
	if data, err := json.Marshal(s); err == nil {
		r.redis.Set(ctx, key, data, cacheTTL)
	}

	return &s, nil
}

func (r *serviceRepository) Update(ctx context.Context, s *entity.Service) error {
	query := `UPDATE services SET name = $1, description = $2, price = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4 RETURNING updated_at`
	err := r.db.QueryRow(ctx, query, s.Name, s.Description, s.Price, s.ID).Scan(&s.UpdatedAt)
	if err != nil {
		return err
	}

	r.InvalidateCache(ctx, s.ID.String())
	return nil
}

func (r *serviceRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM services WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	r.InvalidateCache(ctx, id)
	return nil
}

func (r *serviceRepository) InvalidateCache(ctx context.Context, id string) {
	keys := []string{cacheKeyAll}
	if id != "" {
		keys = append(keys, fmt.Sprintf(cacheKeyDetail, id))
	}

	if err := r.redis.Del(ctx, keys...).Err(); err != nil {
		slog.ErrorContext(ctx, "Failed to invalidate cache", "keys", keys, "error", err)
		return
	}

	slog.InfoContext(ctx, "Cache invalidated", "keys", keys)
}
