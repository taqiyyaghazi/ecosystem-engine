package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/dto"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/entity"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/repository"
)

type PartnerUsecase interface {
	JoinAsPartner(ctx context.Context, userID string, serviceID string) (*dto.PartnerResponse, error)
}

type partnerUsecase struct {
	serviceRepo repository.ServiceRepository
	partnerRepo repository.PartnerRepository
}

func NewPartnerUsecase(serviceRepo repository.ServiceRepository, partnerRepo repository.PartnerRepository) PartnerUsecase {
	return &partnerUsecase{
		serviceRepo: serviceRepo,
		partnerRepo: partnerRepo,
	}
}

func (u *partnerUsecase) JoinAsPartner(ctx context.Context, userID string, serviceID string) (*dto.PartnerResponse, error) {
	// Validate UUID format
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.ErrInvalidUUID
	}
	sid, err := uuid.Parse(serviceID)
	if err != nil {
		return nil, apperror.ErrInvalidUUID
	}

	// Validate service existence
	_, _, err = u.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Check for existing partner record
	existingPartner, err := u.partnerRepo.GetPartnerByUserAndService(ctx, userID, serviceID)
	if err == nil && existingPartner != nil {
		return nil, apperror.NewInvalidInputErrorWithMessage("user already registered as partner for this service")
	}

	// Create partner record
	partner := &entity.Partner{
		UserID:    uid,
		ServiceID: sid,
	}

	if err := u.partnerRepo.CreatePartner(ctx, partner); err != nil {
		return nil, err
	}

	return u.toResponse(partner), nil
}

func (u *partnerUsecase) toResponse(p *entity.Partner) *dto.PartnerResponse {
	return &dto.PartnerResponse{
		ID:        p.ID.String(),
		UserID:    p.UserID.String(),
		ServiceID: p.ServiceID.String(),
		IsActive:  p.IsActive,
		Rating:    p.Rating,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
	}
}
