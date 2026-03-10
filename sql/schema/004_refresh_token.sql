-- +goose Up
CREATE TABLE IF NOT EXISTS refresh_tokens (
    token VARCHAR(255) Primary Key,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at  TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP 
);



-- +goose Down
DROP TABLE refresh_tokens;