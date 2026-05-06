package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
	notificationDTO "github.com/taqiyyaghazi/ecosystem-engine/internal/features/notifications/dto"
	notificationUsecase "github.com/taqiyyaghazi/ecosystem-engine/internal/features/notifications/usecase"
	serviceDTO "github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/dto"
	serviceEntity "github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/entity"
	serviceRepository "github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/repository"
)

type PartnerUsecase interface {
	JoinAsPartner(ctx context.Context, userID string, serviceID string) (*serviceDTO.PartnerResponse, error)
}

type partnerUsecase struct {
	serviceRepository   serviceRepository.ServiceRepository
	partnerRepository   serviceRepository.PartnerRepository
	notificationUseCase notificationUsecase.NotificationUsecase
}

func NewPartnerUsecase(serviceRepository serviceRepository.ServiceRepository, partnerRepository serviceRepository.PartnerRepository, notificationUseCase notificationUsecase.NotificationUsecase) PartnerUsecase {
	return &partnerUsecase{
		serviceRepository:   serviceRepository,
		partnerRepository:   partnerRepository,
		notificationUseCase: notificationUseCase,
	}
}

func (u *partnerUsecase) JoinAsPartner(ctx context.Context, userID string, serviceID string) (*serviceDTO.PartnerResponse, error) {
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
	_, _, err = u.serviceRepository.GetByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Check for existing partner record
	existingPartner, err := u.partnerRepository.GetPartnerByUserAndService(ctx, userID, serviceID)
	if err == nil && existingPartner != nil {
		return nil, apperror.NewInvalidInputErrorWithMessage("user already registered as partner for this service")
	}

	// Create partner record
	partner := &serviceEntity.Partner{
		UserID:    uid,
		ServiceID: sid,
	}

	if err := u.partnerRepository.CreatePartner(ctx, partner); err != nil {
		return nil, err
	}

	res := u.toResponse(partner)

	// Publish welcome notification as fire-and-forget
	channel := "notifications:user:" + res.ID // Using Partner ID as recipient ID
	go func() {
		msg := notificationDTO.NotificationMessage{
			RecipientID: res.ID,
			Title:       "Welcome to the Program!",
			Message:     "You have successfully joined as a partner for this service.",
		}
		// Use a detached background context with timeout just in case it takes time
		if err := u.notificationUseCase.Send(context.Background(), channel, msg); err != nil {
			slog.Error("failed to publish welcome notification", "error", err, "partner_id", res.ID)
		}
	}()

	return res, nil
}

func (u *partnerUsecase) toResponse(p *serviceEntity.Partner) *serviceDTO.PartnerResponse {
	return &serviceDTO.PartnerResponse{
		ID:        p.ID.String(),
		UserID:    p.UserID.String(),
		ServiceID: p.ServiceID.String(),
		IsActive:  p.IsActive,
		Rating:    p.Rating,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
	}
}
