package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/apperror"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/auth/entity"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/platform/database"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (username, password, role) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := r.db.QueryRow(ctx, query, user.Username, user.Password, user.Role).
		Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == database.ErrCodeUniqueViolation {
			return apperror.NewInvalidInputErrorWithMessage("username already exists")
		}
		return err
	}
	return nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `SELECT id, username, password, role, created_at FROM users WHERE username = $1`

	var user entity.User
	err := r.db.QueryRow(ctx, query, username).
		Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}
