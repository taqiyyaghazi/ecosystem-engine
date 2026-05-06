package dto

type PointRequest struct {
	PartnerID string `json:"partner_id" binding:"required,uuid"`
	Amount    int    `json:"amount" binding:"required,gt=0"`
	Reason    string `json:"reason" binding:"required,min=5,max=255"`
}

type LeaderboardEntry struct {
	Rank      int64   `json:"rank"`
	PartnerID string  `json:"partner_id"`
	Score     float64 `json:"score"`
}

type LeaderboardResponse struct {
	Entries []LeaderboardEntry `json:"entries"`
}

type LeaderboardQuery struct {
	ServiceID string `form:"service_id" binding:"required,uuid"`
	Limit     int64  `form:"limit" binding:"omitempty,gt=0"`
}

type PartnerRankQuery struct {
	ServiceID string `form:"service_id" binding:"required,uuid"`
}
