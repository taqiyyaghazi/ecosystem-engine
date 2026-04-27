package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/dto"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/entity"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/repository"
)

type ServiceUsecase interface {
	Create(ctx context.Context, req dto.CreateServiceRequest) (*dto.ServiceResponse, error)
	GetAll(ctx context.Context) ([]dto.ServiceResponse, context.Context, error)
	GetByID(ctx context.Context, id string) (*dto.ServiceResponse, context.Context, error)
	Update(ctx context.Context, id string, req dto.UpdateServiceRequest) (*dto.ServiceResponse, error)
	Delete(ctx context.Context, id string) error
}

type serviceUsecase struct {
	repo repository.ServiceRepository
}

func NewServiceUsecase(repo repository.ServiceRepository) ServiceUsecase {
	return &serviceUsecase{
		repo: repo,
	}
}

func (u *serviceUsecase) Create(ctx context.Context, req dto.CreateServiceRequest) (*dto.ServiceResponse, error) {
	s := &entity.Service{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	}

	if err := u.repo.Create(ctx, s); err != nil {
		return nil, err
	}

	return u.toResponse(s), nil
}

func (u *serviceUsecase) GetAll(ctx context.Context) ([]dto.ServiceResponse, context.Context, error) {
	services, ctx, err := u.repo.GetAll(ctx)
	if err != nil {
		return nil, ctx, err
	}

	var res []dto.ServiceResponse
	for _, s := range services {
		res = append(res, *u.toResponse(&s))
	}

	return res, ctx, nil
}

func (u *serviceUsecase) GetByID(ctx context.Context, id string) (*dto.ServiceResponse, context.Context, error) {
	s, ctx, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ctx, apperror.ErrNotFound
		}
		return nil, ctx, err
	}

	return u.toResponse(s), ctx, nil
}

func (u *serviceUsecase) Update(ctx context.Context, id string, req dto.UpdateServiceRequest) (*dto.ServiceResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperror.ErrInvalidUUID
	}

	s, _, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		s.Name = *req.Name
	}
	if req.Description != nil {
		s.Description = *req.Description
	}
	if req.Price != nil {
		s.Price = *req.Price
	}
	s.ID = uid

	if err := u.repo.Update(ctx, s); err != nil {
		return nil, err
	}

	return u.toResponse(s), nil
}

func (u *serviceUsecase) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return apperror.ErrInvalidUUID
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

func (u *serviceUsecase) toResponse(s *entity.Service) *dto.ServiceResponse {
	return &dto.ServiceResponse{
		ID:          s.ID.String(),
		Name:        s.Name,
		Description: s.Description,
		Price:       s.Price,
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
	}
}
