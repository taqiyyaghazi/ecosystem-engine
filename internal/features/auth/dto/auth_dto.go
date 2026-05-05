package dto

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

type SessionResponse struct {
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
	UserAgent    string `json:"user_agent"`
	IPAddress    string `json:"ip_address"`
	LastActivity string `json:"last_activity"`
}
