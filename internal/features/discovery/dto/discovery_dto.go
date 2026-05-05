package dto

// LocationRequest is the payload for POST /v1/discovery/location.
// It carries the partner's current GPS coordinates.
type LocationRequest struct {
	Latitude  float64 `json:"latitude"  binding:"required,latitude"`
	Longitude float64 `json:"longitude" binding:"required,longitude"`
}

// PartnerDetail contains extra information about a partner fetched from PostgreSQL.
type PartnerDetail struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Rating      float64 `json:"rating"`
	ServiceType string  `json:"service_type"`
}

// NearbyPartnerResponse represents a single partner entry returned by
// GET /v1/discovery/nearby. Distance is a human-readable string
// (e.g. "1.2 km").
type NearbyPartnerResponse struct {
	PartnerID   string  `json:"partner_id"`
	Name        string  `json:"name"`
	Rating      float64 `json:"rating"`
	ServiceType string  `json:"service_type"`
	Distance    string  `json:"distance"`
	RawDist     float64 `json:"-"` // internal; used for sorting / formatting
}

// NearbyRequest defines the query parameters for GET /v1/discovery/nearby.
type NearbyRequest struct {
	Lat    float64 `form:"lat"    binding:"required,latitude"`
	Lon    float64 `form:"lon"    binding:"required,longitude"`
	Radius float64 `form:"radius" binding:"omitempty,gt=0"`
	Unit   string  `form:"unit"   binding:"omitempty,oneof=m km mi ft"`
	Limit  int     `form:"limit"  binding:"omitempty,gt=0"`
}
