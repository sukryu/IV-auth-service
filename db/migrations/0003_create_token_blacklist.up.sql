CREATE TABLE token_blacklist (
    token_id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(36),
    expires_at TIMESTAMP NOT NULL,
    reason VARCHAR(50),
    blacklisted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);