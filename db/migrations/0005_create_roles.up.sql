-- 0005_create_roles.up.sql
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id VARCHAR(36) NOT NULL,
    role_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

-- 기본 역할 추가
INSERT INTO roles (id, name, description) VALUES
    ('role-admin', 'ADMIN', '시스템 관리자 역할'),
    ('role-user', 'USER', '일반 사용자 역할'),
    ('role-streamer', 'STREAMER', '스트리머 역할');