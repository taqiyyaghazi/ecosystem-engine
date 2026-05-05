-- +goose Up
CREATE TABLE partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    service_type VARCHAR(50) NOT NULL, -- e.g., 'AC', 'Clean', 'Massage'
    is_active BOOLEAN DEFAULT true,
    rating DECIMAL(2, 1) DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE partners;
