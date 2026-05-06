package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/tasks/entity"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, task *entity.Task) error
	UpdateTaskStatus(ctx context.Context, id string, status string, errMsg *string) error
	PushTask(ctx context.Context, queueName string, taskID string) error
	PopTask(ctx context.Context, queueName string) (string, error)
}

type taskRepository struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewTaskRepository(db *pgxpool.Pool, rdb *redis.Client) TaskRepository {
	return &taskRepository{
		db:    db,
		redis: rdb,
	}
}

func (r *taskRepository) CreateTask(ctx context.Context, task *entity.Task) error {
	query := `
		INSERT INTO tasks (id, task_type, payload, status, error_message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err := r.db.Exec(ctx, query, task.ID, task.TaskType, task.Payload, task.Status, task.ErrorMessage)
	return err
}

func (r *taskRepository) UpdateTaskStatus(ctx context.Context, id string, status string, errMsg *string) error {
	query := `
		UPDATE tasks
		SET status = $1, error_message = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, status, errMsg, id)
	return err
}

func (r *taskRepository) PushTask(ctx context.Context, queueName string, taskID string) error {
	return r.redis.LPush(ctx, queueName, taskID).Err()
}

func (r *taskRepository) PopTask(ctx context.Context, queueName string) (string, error) {
	// BRPop blocks until a task is available
	res, err := r.redis.BRPop(ctx, 0, queueName).Result()
	if err != nil {
		return "", err
	}
	if len(res) < 2 {
		return "", fmt.Errorf("unexpected brpop result format")
	}
	// res[0] is queue name, res[1] is the task payload
	return res[1], nil
}
