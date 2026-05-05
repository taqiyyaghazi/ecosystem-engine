package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/discovery/dto"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/discovery/repository"
)

// DiscoveryUsecase is the business-logic contract for Phase 4 discovery.
type DiscoveryUsecase interface {
	UpdateLocation(ctx context.Context, userID string, req dto.LocationRequest) error
	GetNearby(ctx context.Context, lat, lon, radius float64, unit string, limit int) ([]dto.NearbyPartnerResponse, error)
	SetOffline(ctx context.Context, userID string) error
	StartCleanupWorker(ctx context.Context)
}

type discoveryUsecase struct {
	repo repository.DiscoveryRepository
}

// NewDiscoveryUsecase constructs a DiscoveryUsecase.
func NewDiscoveryUsecase(repo repository.DiscoveryRepository) DiscoveryUsecase {
	return &discoveryUsecase{repo: repo}
}

// validateCoordinates enforces geographic bounds as specified in the DoD.
func validateCoordinates(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return apperror.NewInvalidInputErrorWithMessage(
			fmt.Sprintf("latitude must be between -90 and 90, got %.6f", lat),
		)
	}
	if lon < -180 || lon > 180 {
		return apperror.NewInvalidInputErrorWithMessage(
			fmt.Sprintf("longitude must be between -180 and 180, got %.6f", lon),
		)
	}
	return nil
}

// formatDistance converts a raw float distance (in km by default) to a
// human-readable string such as "1.2 km" or "850 m".
func formatDistance(dist float64, unit string) string {
	if unit == "m" || unit == "M" {
		if dist >= 1000 {
			return fmt.Sprintf("%.1f km", dist/1000)
		}
		return fmt.Sprintf("%.0f m", dist)
	}
	// km (default)
	return fmt.Sprintf("%.1f km", dist)
}

// UpdateLocation validates and stores the partner's location in Redis.
func (u *discoveryUsecase) UpdateLocation(ctx context.Context, userID string, req dto.LocationRequest) error {
	if err := validateCoordinates(req.Latitude, req.Longitude); err != nil {
		return err
	}
	partnerID, err := u.repo.GetPartnerIDByUserID(ctx, userID)
	if err != nil {
		return err // could map to not found
	}
	return u.repo.UpdateLocation(ctx, partnerID, req.Latitude, req.Longitude)
}

// GetNearby validates the search centre, queries Redis GEOSEARCH, and
// returns enriched results with human-readable distances.
func (u *discoveryUsecase) GetNearby(ctx context.Context, lat, lon, radius float64, unit string, limit int) ([]dto.NearbyPartnerResponse, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return nil, err
	}

	if unit == "" {
		unit = "km"
	}

	locations, err := u.repo.GetNearby(ctx, lat, lon, radius, unit, limit)
	if err != nil {
		return nil, err
	}

	if len(locations) == 0 {
		return []dto.NearbyPartnerResponse{}, nil
	}

	// Extract IDs for batch fetching
	partnerIDs := make([]string, 0, len(locations))
	for _, loc := range locations {
		partnerIDs = append(partnerIDs, loc.Name)
	}

	// Batch fetch partner details
	detailsMap, err := u.repo.GetPartnersDetails(ctx, partnerIDs)
	if err != nil {
		return nil, err
	}

	results := make([]dto.NearbyPartnerResponse, 0, len(locations))
	for _, loc := range locations {
		detail, ok := detailsMap[loc.Name]
		if !ok {
			continue // Skip if partner no longer exists in DB
		}
		results = append(results, dto.NearbyPartnerResponse{
			PartnerID:   loc.Name,
			Name:        detail.Name,
			Rating:      detail.Rating,
			ServiceType: detail.ServiceType,
			Distance:    formatDistance(loc.Dist, unit),
			RawDist:     loc.Dist,
		})
	}
	return results, nil
}

// SetOffline explicitly removes a partner's location from Redis.
func (u *discoveryUsecase) SetOffline(ctx context.Context, userID string) error {
	partnerID, err := u.repo.GetPartnerIDByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return u.repo.SetOffline(ctx, partnerID)
}

// StartCleanupWorker starts a background goroutine to clean up stale locations.
func (u *discoveryUsecase) StartCleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				cutoff := time.Now().Add(-5 * time.Minute).Unix()
				_ = u.repo.CleanupStaleLocations(context.Background(), cutoff)
			}
		}
	}()
}
