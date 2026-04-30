package entity

import "time"

// SessionTTL is the single source of truth for session duration,
// shared by both the Redis session store and the HTTP cookie.
const SessionTTL = 24 * time.Hour
