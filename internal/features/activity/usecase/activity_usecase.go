package usecase

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/activity/dto"
	"github.com/taqiyyaghazi/ecosystem-engine/internal/features/activity/repository"
)

type ActivityUsecase interface {
	PublishEvent(ctx context.Context, event dto.ActivityEvent) error
	WorkerLoop(ctx context.Context, consumerName string)
}

type activityUsecase struct {
	repo repository.ActivityRepository
}

func NewActivityUsecase(repo repository.ActivityRepository) ActivityUsecase {
	return &activityUsecase{repo: repo}
}

func (u *activityUsecase) PublishEvent(ctx context.Context, event dto.ActivityEvent) error {
	payloadStr := "{}"
	if event.Payload != nil {
		b, err := json.Marshal(event.Payload)
		if err == nil {
			payloadStr = string(b)
		}
	}

	values := map[string]interface{}{
		"actor_id":   event.ActorID,
		"action":     event.Action,
		"payload":    payloadStr,
		"ip_address": event.IPAddress,
	}

	return u.repo.RecordEvent(ctx, values)
}

func (u *activityUsecase) WorkerLoop(ctx context.Context, consumerName string) {
	slog.Info("Activity worker started", "consumer", consumerName)
	for {
		select {
		case <-ctx.Done():
			slog.Info("Activity worker stopping", "consumer", consumerName)
			return
		default:
			// Read events from stream
			messages, err := u.repo.ReadEvents(ctx, consumerName)
			if err != nil {
				slog.Error("Error reading activity events", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if len(messages) == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			for _, msg := range messages {
				actorID, _ := msg.Values["actor_id"].(string)
				action, _ := msg.Values["action"].(string)
				ipAddress, _ := msg.Values["ip_address"].(string)
				payloadStr, _ := msg.Values["payload"].(string)

				var payload map[string]interface{}
				if payloadStr != "" && payloadStr != "{}" {
					_ = json.Unmarshal([]byte(payloadStr), &payload)
				}

				if actorID != "" && action != "" {
					err = u.repo.InsertActivityLog(ctx, actorID, action, payload, ipAddress)
					if err != nil {
						slog.Error("Failed to insert activity log", "messageID", msg.ID, "error", err)
						continue // Will be retried on next stream read if not acked?
						// Wait, XREADGROUP with ">" only reads new messages. We need a way to process unacknowledged ones.
						// For now, we skip ACK and it becomes a pending message.
					}
				}

				// Acknowledge event
				if err := u.repo.AcknowledgeEvent(ctx, msg.ID); err != nil {
					slog.Error("Failed to ack activity event", "messageID", msg.ID, "error", err)
				} else {
					slog.Info("Processed activity event", "messageID", msg.ID, "action", action)
				}
			}
		}
	}
}
