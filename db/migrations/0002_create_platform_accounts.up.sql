-- 0002_create_platform_accounts.up.sql
CREATE TABLE IF NOT EXISTS platform_accounts (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    platform VARCHAR(20) NOT NULL,
    platform_user_id VARCHAR(100) NOT NULL,
    platform_username VARCHAR(100) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 인덱스 생성
CREATE INDEX idx_platform_accounts_user_id ON platform_accounts(user_id);
CREATE UNIQUE INDEX idx_platform_accounts_platform_user ON platform_accounts(platform, platform_user_id);