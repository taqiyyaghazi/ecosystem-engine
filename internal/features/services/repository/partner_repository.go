package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/entity"
)

type PartnerRepository interface {
	CreatePartner(ctx context.Context, partner *entity.Partner) error
	GetPartnerByUserAndService(ctx context.Context, userID string, serviceID string) (*entity.Partner, error)
}

type partnerRepository struct {
	db *pgxpool.Pool
}

func NewPartnerRepository(db *pgxpool.Pool) PartnerRepository {
	return &partnerRepository{
		db: db,
	}
}

func (r *partnerRepository) CreatePartner(ctx context.Context, p *entity.Partner) error {
	query := `INSERT INTO partners (user_id, service_id) VALUES ($1, $2) RETURNING id, is_active, rating, created_at`
	err := r.db.QueryRow(ctx, query, p.UserID, p.ServiceID).Scan(&p.ID, &p.IsActive, &p.Rating, &p.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *partnerRepository) GetPartnerByUserAndService(ctx context.Context, userID string, serviceID string) (*entity.Partner, error) {
	var p entity.Partner
	query := `SELECT id, user_id, service_id, is_active, rating, created_at FROM partners WHERE user_id = $1 AND service_id = $2`
	err := r.db.QueryRow(ctx, query, userID, serviceID).Scan(&p.ID, &p.UserID, &p.ServiceID, &p.IsActive, &p.Rating, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	return &p, nil
}
