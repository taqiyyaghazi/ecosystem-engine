package dto

type ActivityEvent struct {
	ActorID   string                 `json:"actor_id"`
	Action    string                 `json:"action"`
	Payload   map[string]interface{} `json:"payload"`
	IPAddress string                 `json:"ip_address"`
}
