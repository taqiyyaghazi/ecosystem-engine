package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/dto"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/entity"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/services/repository"
)

type ServiceUsecase interface {
	Create(ctx context.Context, req dto.CreateServiceRequest) (*dto.ServiceResponse, error)
	GetAll(ctx context.Context) ([]dto.ServiceResponse, error)
	GetByID(ctx context.Context, id string) (*dto.ServiceResponse, error)
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

func (u *serviceUsecase) GetAll(ctx context.Context) ([]dto.ServiceResponse, error) {
	services, err := u.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var res []dto.ServiceResponse
	for _, s := range services {
		res = append(res, *u.toResponse(&s))
	}

	return res, nil
}

func (u *serviceUsecase) GetByID(ctx context.Context, id string) (*dto.ServiceResponse, error) {
	s, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return u.toResponse(s), nil
}

func (u *serviceUsecase) Update(ctx context.Context, id string, req dto.UpdateServiceRequest) (*dto.ServiceResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	s, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		s.Name = req.Name
	}
	if req.Description != "" {
		s.Description = req.Description
	}
	if req.Price != 0 {
		s.Price = req.Price
	}
	s.ID = uid

	if err := u.repo.Update(ctx, s); err != nil {
		return nil, err
	}

	return u.toResponse(s), nil
}

func (u *serviceUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func (u *serviceUsecase) toResponse(s *entity.Service) *dto.ServiceResponse {
	return &dto.ServiceResponse{
		ID:          s.ID.String(),
		Name:        s.Name,
		Description: s.Description,
		Price:       s.Price,
		CreatedAt:   s.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   s.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
