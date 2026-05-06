package dto

type PartnerResponse struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	ServiceID string  `json:"service_id"`
	IsActive  bool    `json:"is_active"`
	Rating    float64 `json:"rating"`
	CreatedAt string  `json:"created_at"`
}
