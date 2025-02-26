-- 0003_create_token_blacklist.up.sql
CREATE TABLE IF NOT EXISTS token_blacklist (
    token_id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    reason VARCHAR(50),
    blacklisted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 인덱스 생성
CREATE INDEX idx_token_blacklist_expires_at ON token_blacklist(expires_at);