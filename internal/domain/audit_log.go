package domain

import (
	"errors"
	"time"
)

// AuditLog는 시스템 내 주요 이벤트에 대한 감사 로그를 표현합니다.
type AuditLog struct {
	ID         string
	UserID     string
	Action     string
	EntityType string
	EntityID   string
	OldValues  map[string]interface{}
	NewValues  map[string]interface{}
	IPAddress  string
	UserAgent  string
	CreatedAt  time.Time
}

// NewAuditLog는 새로운 감사 로그 항목을 생성합니다.
func NewAuditLog(id, userID, action, entityType, entityID string) (*AuditLog, error) {
	if id == "" {
		return nil, errors.New("감사 로그 ID는 필수입니다")
	}
	if action == "" {
		return nil, errors.New("작업 유형은 필수입니다")
	}
	if entityType == "" {
		return nil, errors.New("엔티티 유형은 필수입니다")
	}

	return &AuditLog{
		ID:         id,
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		OldValues:  make(map[string]interface{}),
		NewValues:  make(map[string]interface{}),
		CreatedAt:  time.Now(),
	}, nil
}

// SetOldValues는 변경 전 값들을 설정합니다.
func (al *AuditLog) SetOldValues(values map[string]interface{}) {
	al.OldValues = values
}

// SetNewValues는 변경 후 값들을 설정합니다.
func (al *AuditLog) SetNewValues(values map[string]interface{}) {
	al.NewValues = values
}

// SetIPAddress는 요청 IP 주소를 설정합니다.
func (al *AuditLog) SetIPAddress(ip string) {
	al.IPAddress = ip
}

// SetUserAgent는 사용자 에이전트 정보를 설정합니다.
func (al *AuditLog) SetUserAgent(userAgent string) {
	al.UserAgent = userAgent
}
