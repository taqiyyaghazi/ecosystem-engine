package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/entity"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/cache/cacheutil"
)

type ServiceRepository interface {
	Create(ctx context.Context, service *entity.Service) error
	GetAll(ctx context.Context) ([]entity.Service, string, error)
	GetByID(ctx context.Context, id string) (*entity.Service, string, error)
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

func (r *serviceRepository) GetAll(ctx context.Context) ([]entity.Service, string, error) {
	var services []entity.Service
	if r.getCache(ctx, cacheKeyAll, &services) {
		return services, cacheutil.StatusHIT, nil
	}

	slog.DebugContext(ctx, "Cache MISS", "key", cacheKeyAll)

	query := `SELECT id, name, description, price, created_at, updated_at FROM services ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	for rows.Next() {
		var s entity.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Price, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, "", err
		}
		services = append(services, s)
	}

	r.setCache(ctx, cacheKeyAll, services)

	return services, cacheutil.StatusMISS, nil
}

func (r *serviceRepository) GetByID(ctx context.Context, id string) (*entity.Service, string, error) {
	key := fmt.Sprintf(cacheKeyDetail, id)

	var s entity.Service
	if r.getCache(ctx, key, &s) {
		return &s, cacheutil.StatusHIT, nil
	}

	slog.DebugContext(ctx, "Cache MISS", "key", key)

	query := `SELECT id, name, description, price, created_at, updated_at FROM services WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&s.ID, &s.Name, &s.Description, &s.Price, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", apperror.ErrNotFound
		}
		return nil, "", err
	}

	r.setCache(ctx, key, s)

	return &s, cacheutil.StatusMISS, nil
}

func (r *serviceRepository) Update(ctx context.Context, s *entity.Service) error {
	query := `UPDATE services SET name = $1, description = $2, price = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4 RETURNING updated_at`
	err := r.db.QueryRow(ctx, query, s.Name, s.Description, s.Price, s.ID).Scan(&s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.ErrNotFound
		}
		return err
	}

	r.InvalidateCache(ctx, s.ID.String())
	return nil
}

func (r *serviceRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM services WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
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

func (r *serviceRepository) getCache(ctx context.Context, key string, dest interface{}) bool {
	val, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		return false
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false
	}

	slog.DebugContext(ctx, "Cache HIT", "key", key)
	return true
}

func (r *serviceRepository) setCache(ctx context.Context, key string, value interface{}) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	r.redis.Set(ctx, key, data, cacheTTL)
}
