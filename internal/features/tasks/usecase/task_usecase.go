package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/tasks/dto"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/tasks/entity"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/tasks/repository"
)

type TaskUsecase interface {
	DispatchInvoiceTask(ctx context.Context, req *dto.CreateInvoiceTaskRequest) error
	WorkerLoop(ctx context.Context, queueName string)
}

type taskUsecase struct {
	repo repository.TaskRepository
}

func NewTaskUsecase(repo repository.TaskRepository) TaskUsecase {
	return &taskUsecase{
		repo: repo,
	}
}

func (u *taskUsecase) DispatchInvoiceTask(ctx context.Context, req *dto.CreateInvoiceTaskRequest) error {
	taskID := uuid.New()

	// Create payload
	payload := dto.TaskPayload{
		TaskID: taskID.String(),
		Type:   "GENERATE_INVOICE",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := &entity.Task{
		ID:       taskID,
		TaskType: "GENERATE_INVOICE",
		Payload:  payloadBytes,
		Status:   "PENDING",
	}

	// 1. Save to Postgres
	if err := u.repo.CreateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to create task in db: %w", err)
	}

	// 2. Push to Redis
	if err := u.repo.PushTask(ctx, "tasks:default", taskID.String()); err != nil {
		return fmt.Errorf("failed to push task to redis: %w", err)
	}

	slog.Info("Successfully dispatched invoice task", "taskID", taskID.String())
	return nil
}

func (u *taskUsecase) WorkerLoop(ctx context.Context, queueName string) {
	slog.Info("Task worker started listening to queue", "queueName", queueName)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Task worker shutting down")
			return
		default:
			// BRPOP blocks until an item is available
			taskID, err := u.repo.PopTask(ctx, queueName)
			if err != nil {
				// Don't log if context is canceled
				if ctx.Err() != nil {
					return
				}
				slog.Error("Failed to pop task", "error", err)
				time.Sleep(1 * time.Second) // backoff
				continue
			}

			slog.Info("Picked up task", "taskID", taskID)

			// Process task in background or synchronously. For simple queue, we do it synchronously.
			u.processTask(ctx, taskID)
		}
	}
}

func (u *taskUsecase) processTask(ctx context.Context, taskID string) {
	slog.Info("Processing task", "taskID", taskID)

	// Update status to PROCESSING
	if err := u.repo.UpdateTaskStatus(ctx, taskID, "PROCESSING", nil); err != nil {
		slog.Error("Failed to update task status to PROCESSING", "taskID", taskID, "error", err)
		return
	}

	// Generate PDF
	err := generatePDFInvoice(taskID)

	if err != nil {
		slog.Error("Failed to process task", "taskID", taskID, "error", err)
		errMsg := err.Error()
		// Update status to FAILED
		_ = u.repo.UpdateTaskStatus(ctx, taskID, "FAILED", &errMsg)
		return
	}

	// Update status to COMPLETED
	if err := u.repo.UpdateTaskStatus(ctx, taskID, "COMPLETED", nil); err != nil {
		slog.Error("Failed to update task status to COMPLETED", "taskID", taskID, "error", err)
		return
	}

	slog.Info("Successfully completed task", "taskID", taskID)
}

func generatePDFInvoice(taskID string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, fmt.Sprintf("Invoice for Task: %s", taskID))
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Date: %s", time.Now().Format("2006-01-02 15:04:05")))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Total: $100.00")

	// Ensure the invoices directory exists
	dir := "invoices"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	filePath := filepath.Join(dir, fmt.Sprintf("invoice_%s.pdf", taskID))
	if err := pdf.OutputFileAndClose(filePath); err != nil {
		return fmt.Errorf("failed to save pdf: %w", err)
	}

	return nil
}
