-- 테스트용 사용자 추가
INSERT INTO users (id, username, email, password_hash, status, subscription_tier, created_at, updated_at)
VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    'testuser',
    'testuser@example.com',
    '$argon2id$v=19$m=65536,t=3,p=1$randomsalt$hashedpassword', -- 예시 해시
    'ACTIVE',
    'FREE',
    NOW(),
    NOW()
);

-- 테스트용 플랫폼 계정 추가
INSERT INTO platform_accounts (id, user_id, platform, platform_user_id, platform_username, access_token, refresh_token, token_expires_at, created_at, updated_at)
VALUES (
    '550e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440000',
    'TWITCH',
    'twitch123',
    'TestTwitchUser',
    'mock_access_token',
    'mock_refresh_token',
    NOW() + INTERVAL '1 hour',
    NOW(),
    NOW()
);