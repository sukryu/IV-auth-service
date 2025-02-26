package repo

import (
	"context"

	"github.com/sukryu/IV-auth-services/internal/domain"
)

// PlatformRepository는 플랫폼 계정 데이터에 대한 접근을 정의하는 인터페이스입니다.
// 데이터베이스(PostgreSQL 등)와의 상호작용을 추상화하며, 플랫폼 연동 관리에 필요한 CRUD 작업을 제공합니다.
type PlatformRepository interface {
	// FindByID는 플랫폼 계정 ID로 계정을 조회합니다.
	// 계정이 존재하지 않을 경우 nil과 nil 에러를 반환하며, 조회 실패 시 에러를 반환합니다.
	FindByID(ctx context.Context, id string) (*domain.PlatformAccount, error)

	// Insert는 새로운 플랫폼 계정을 데이터베이스에 추가합니다.
	// 성공 시 nil을 반환하며, 삽입 실패(중복 키 등) 시 에러를 반환합니다.
	Insert(ctx context.Context, account *domain.PlatformAccount) error

	// Update는 기존 플랫폼 계정을 갱신합니다.
	// 토큰 갱신 등을 처리하며, 실패 시 에러를 반환합니다.
	Update(ctx context.Context, account *domain.PlatformAccount) error

	// Delete는 플랫폼 계정 ID로 계정을 삭제합니다.
	// 물리적 삭제로 구현하며, 실패 시 에러를 반환합니다.
	Delete(ctx context.Context, id string) error
}
