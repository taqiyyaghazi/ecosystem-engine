-- +goose Up
CREATE TABLE notification_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_history_recipient_id ON notification_history (recipient_id);

-- +goose Down
DROP INDEX IF EXISTS idx_notification_history_recipient_id;
DROP TABLE notification_history;
