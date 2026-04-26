package dto

type CreateServiceRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required"`
}

// UpdateServiceRequest uses pointer fields so the usecase can distinguish
// between "field not sent" (nil) and "field sent with zero value" (e.g. price=0).
type UpdateServiceRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
}

type ServiceResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}
