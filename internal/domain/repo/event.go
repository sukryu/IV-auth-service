package repo

import (
	"context"

	"github.com/sukryu/IV-auth-services/internal/domain"
)

// EventPublisher는 도메인 이벤트를 외부 메시지 시스템(Kafka 등)에 발행하는 인터페이스입니다.
// 인증 성공/실패, 토큰 무효화 등의 비동기 이벤트를 처리하며, 느슨한 결합을 유지합니다.
type EventPublisher interface {
	// Publish는 도메인 이벤트를 발행합니다.
	// 이벤트 타입별로 적절한 토픽(예: auth.events.login_succeeded)에 전송하며, 실패 시 에러를 반환합니다.
	Publish(ctx context.Context, event interface{}) error

	// BatchPublish는 여러 이벤트를 일괄 발행합니다.
	// 성능 최적화를 위해 배치 처리를 지원하며, 하나라도 실패 시 에러를 반환합니다.
	BatchPublish(ctx context.Context, events []interface{}) error

	// Close는 이벤트 발행 리소스를 정리합니다.
	// Kafka Producer 종료 등 graceful shutdown을 위해 사용하며, 실패 시 에러를 반환합니다.
	Close() error
}

var (
	_ interface{} = (*domain.LoginSucceeded)(nil)
	_ interface{} = (*domain.LoginFailed)(nil)
	_ interface{} = (*domain.TokenBlacklisted)(nil)
	_ interface{} = (*domain.UserCreated)(nil)
	_ interface{} = (*domain.UserUpdated)(nil)
	_ interface{} = (*domain.UserDeleted)(nil)
	_ interface{} = (*domain.PlatformConnected)(nil)
	_ interface{} = (*domain.PlatformDisconnected)(nil)
	_ interface{} = (*domain.PlatformTokenRefreshed)(nil)
	_ interface{} = (*domain.PlatformConnectionFailed)(nil)
	_ interface{} = (*domain.PlatformTokenRefreshFailed)(nil)
)
