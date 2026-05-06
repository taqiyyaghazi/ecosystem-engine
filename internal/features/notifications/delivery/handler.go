package delivery

import "github.com/taqiyyaghazi/ecosystem-engine/internal/features/notifications/usecase"

// NotificationHandler handles HTTP requests for notifications (e.g. Inbox).
// Phase 7: Empty scaffold. Will be implemented in future phases.
type NotificationHandler struct {
	notificationUseCase usecase.NotificationUsecase
}

func NewNotificationHandler(notificationUseCase usecase.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{notificationUseCase: notificationUseCase}
}
