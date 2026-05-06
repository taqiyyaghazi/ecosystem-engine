package usecase

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/notifications/dto"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/notifications/repository"
)

type NotificationUsecase interface {
	Send(ctx context.Context, channel string, message dto.NotificationMessage) error
	RunSubscriberLoop(ctx context.Context, channelPattern string)
	Shutdown()
}

type notificationUsecase struct {
	repo repository.NotificationRepository
	wg   sync.WaitGroup
}

func NewNotificationUsecase(repo repository.NotificationRepository) NotificationUsecase {
	return &notificationUsecase{
		repo: repo,
	}
}

func (u *notificationUsecase) Send(ctx context.Context, channel string, message dto.NotificationMessage) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return u.repo.Publish(ctx, channel, payload)
}

func (u *notificationUsecase) RunSubscriberLoop(ctx context.Context, channelPattern string) {
	pubsub := u.repo.Subscribe(ctx, channelPattern)
	defer pubsub.Close()

	ch := pubsub.Channel()

	slog.Info("Notification subscriber loop started", "pattern", channelPattern)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Notification subscriber loop stopping due to context cancellation")
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}

			// Process message asynchronously
			u.wg.Add(1)
			go func(payload string) {
				defer u.wg.Done()
				u.processMessage(payload)
			}(msg.Payload)
		}
	}
}

func (u *notificationUsecase) Shutdown() {
	u.wg.Wait()
}

func (u *notificationUsecase) processMessage(payload string) {
	var message dto.NotificationMessage
	if err := json.Unmarshal([]byte(payload), &message); err != nil {
		slog.Error("Failed to unmarshal notification message", "error", err, "payload", payload)
		return
	}

	// Use background context for db operation since caller context might be cancelled during shutdown
	// We want to ensure save completes
	// We want to ensure save completes but with a reasonable timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := u.repo.SaveHistory(ctx, message.RecipientID, message.Title, message.Message); err != nil {
		slog.Error("Failed to save notification history", "error", err, "recipient_id", message.RecipientID)
		return
	}

	slog.Debug("Notification processed successfully", "recipient_id", message.RecipientID)
}
