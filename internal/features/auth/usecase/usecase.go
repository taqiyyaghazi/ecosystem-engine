package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/auth/dto"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/auth/entity"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/auth/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase interface {
	Register(ctx context.Context, req dto.RegisterRequest) error
	Login(ctx context.Context, req dto.LoginRequest, userAgent, ipAddress string) (string, error)
	Logout(ctx context.Context, sessionID string) error
	Me(ctx context.Context, sessionID string) (*dto.SessionResponse, error)
	RefreshSession(ctx context.Context, sessionID string) error
}

type authUsecase struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewAuthUsecase(userRepo repository.UserRepository, sessionRepo repository.SessionRepository) AuthUsecase {
	return &authUsecase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (u *authUsecase) Register(ctx context.Context, req dto.RegisterRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &entity.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     "user",
	}

	if err := u.userRepo.CreateUser(ctx, user); err != nil {
		return err
	}

	return nil
}

func (u *authUsecase) Login(ctx context.Context, req dto.LoginRequest, userAgent, ipAddress string) (string, error) {
	user, err := u.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return "", apperror.NewInvalidInputErrorWithMessage("invalid username or password")
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", apperror.NewInvalidInputErrorWithMessage("invalid username or password")
	}

	sessionID := uuid.New().String()

	if err := u.sessionRepo.CreateSession(ctx, sessionID, *user, userAgent, ipAddress); err != nil {
		return "", err
	}

	return sessionID, nil
}

func (u *authUsecase) Logout(ctx context.Context, sessionID string) error {
	return u.sessionRepo.DeleteSession(ctx, sessionID)
}

func (u *authUsecase) Me(ctx context.Context, sessionID string) (*dto.SessionResponse, error) {
	data, err := u.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, apperror.ErrNotFound
	}

	return &dto.SessionResponse{
		UserID:       data["user_id"],
		Role:         data["role"],
		UserAgent:    data["user_agent"],
		IPAddress:    data["ip_address"],
		LastActivity: data["last_activity"],
	}, nil
}

func (u *authUsecase) RefreshSession(ctx context.Context, sessionID string) error {
	return u.sessionRepo.RefreshTTL(ctx, sessionID)
}
