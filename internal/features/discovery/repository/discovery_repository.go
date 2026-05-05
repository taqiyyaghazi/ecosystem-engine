package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/discovery/dto"
)

const (
	geoKey      = "partner:locations"
	lastSeenKey = "partner:last_seen"
)

// DiscoveryRepository defines Redis GEO operations for partner location tracking.
type DiscoveryRepository interface {
	UpdateLocation(ctx context.Context, partnerID string, lat, lon float64) error
	GetNearby(ctx context.Context, lat, lon, radius float64, unit string, limit int) ([]redis.GeoLocation, error)
	GetPartnerIDByUserID(ctx context.Context, userID string) (string, error)
	GetPartnersDetails(ctx context.Context, partnerIDs []string) (map[string]dto.PartnerDetail, error)
	SetOffline(ctx context.Context, partnerID string) error
	CleanupStaleLocations(ctx context.Context, beforeUnix int64) error
}

type discoveryRepository struct {
	redis *redis.Client
	db    *pgxpool.Pool
}

// NewDiscoveryRepository constructs a DiscoveryRepository backed by Redis and Postgres.
func NewDiscoveryRepository(redis *redis.Client, db *pgxpool.Pool) DiscoveryRepository {
	return &discoveryRepository{redis: redis, db: db}
}

// UpdateLocation executes GEOADD and ZADD in a pipeline with context deadline.
func (r *discoveryRepository) UpdateLocation(ctx context.Context, partnerID string, lat, lon float64) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	pipe := r.redis.Pipeline()
	pipe.GeoAdd(ctx, geoKey, &redis.GeoLocation{
		Name:      partnerID,
		Latitude:  lat,
		Longitude: lon,
	})
	pipe.ZAdd(ctx, lastSeenKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: partnerID,
	})
	_, err := pipe.Exec(ctx)
	return err
}

// GetNearby executes GEOSEARCH with count limit and context deadline.
func (r *discoveryRepository) GetNearby(ctx context.Context, lat, lon, radius float64, unit string, limit int) ([]redis.GeoLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return r.redis.GeoSearchLocation(ctx, geoKey, &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude:  lon,
			Latitude:   lat,
			Radius:     radius,
			RadiusUnit: unit,
			Sort:       "ASC",
			Count:      limit,
		},
		WithDist: true,
	}).Result()
}

// GetPartnerIDByUserID fetches the partner's ID from Postgres.
func (r *discoveryRepository) GetPartnerIDByUserID(ctx context.Context, userID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var partnerID string
	err := r.db.QueryRow(ctx, "SELECT id FROM partners WHERE user_id = $1", userID).Scan(&partnerID)
	return partnerID, err
}

// GetPartnersDetails performs a batch fetch of partner details.
func (r *discoveryRepository) GetPartnersDetails(ctx context.Context, partnerIDs []string) (map[string]dto.PartnerDetail, error) {
	if len(partnerIDs) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `
		SELECT p.id, u.username as name, p.rating, p.service_type
		FROM partners p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ANY($1)
	`
	rows, err := r.db.Query(ctx, query, partnerIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]dto.PartnerDetail)
	for rows.Next() {
		var p dto.PartnerDetail
		if err := rows.Scan(&p.ID, &p.Name, &p.Rating, &p.ServiceType); err != nil {
			return nil, err
		}
		result[p.ID] = p
	}
	return result, nil
}

// SetOffline explicitly removes a partner from the discovery pool.
func (r *discoveryRepository) SetOffline(ctx context.Context, partnerID string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	pipe := r.redis.Pipeline()
	pipe.ZRem(ctx, geoKey, partnerID)
	pipe.ZRem(ctx, lastSeenKey, partnerID)
	_, err := pipe.Exec(ctx)
	return err
}

// CleanupStaleLocations removes partners who haven't updated their location recently.
func (r *discoveryRepository) CleanupStaleLocations(ctx context.Context, beforeUnix int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	stalePartners, err := r.redis.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:     lastSeenKey,
		Min:     "-inf",
		Max:     strconv.FormatInt(beforeUnix, 10),
		ByScore: true,
	}).Result()

	if err != nil || len(stalePartners) == 0 {
		return err
	}

	pipe := r.redis.Pipeline()
	var interfaces []interface{}
	for _, p := range stalePartners {
		interfaces = append(interfaces, p)
	}
	pipe.ZRem(ctx, geoKey, interfaces...)
	pipe.ZRem(ctx, lastSeenKey, interfaces...)
	_, err = pipe.Exec(ctx)
	return err
}
