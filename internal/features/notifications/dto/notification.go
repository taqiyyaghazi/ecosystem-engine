package dto

type NotificationMessage struct {
	RecipientID string `json:"recipient_id"`
	Title       string `json:"title"`
	Message     string `json:"message"`
}
