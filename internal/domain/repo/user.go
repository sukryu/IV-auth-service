package repo

import (
	"context"

	"github.com/sukryu/IV-auth-services/internal/domain"
)

// UserRepository는 사용자 데이터에 대한 접근을 정의하는 인터페이스입니다.
// 데이터베이스(PostgreSQL 등)와의 상호작용을 추상화하며, 인증 및 사용자 관리에 필요한 CRUD 작업을 제공합니다.
type UserRepository interface {
	// FindByUsername은 사용자명을 기준으로 사용자 엔티티를 조회합니다.
	// 사용자가 존재하지 않을 경우 nil과 nil 에러를 반환하며, 조회 실패 시 에러를 반환합니다.
	FindByUsername(ctx context.Context, username string) (*domain.User, error)

	// FindByID는 사용자 ID를 기준으로 사용자 엔티티를 조회합니다.
	// 사용자가 존재하지 않을 경우 nil과 nil 에러를 반환하며, 조회 실패 시 에러를 반환합니다.
	FindByID(ctx context.Context, id string) (*domain.User, error)

	// ExistsByUsername은 주어진 사용자명이 이미 존재하는지 확인합니다.
	// 존재 여부를 bool로 반환하며, 쿼리 실패 시 에러를 반환합니다.
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// Insert는 새로운 사용자 엔티티를 데이터베이스에 추가합니다.
	// 성공 시 nil을 반환하며, 삽입 실패(중복 키 등) 시 에러를 반환합니다.
	Insert(ctx context.Context, user *domain.User) error

	// Update는 기존 사용자 엔티티를 갱신합니다.
	// 예: 마지막 로그인 시간, 비밀번호 해시, 상태 등을 업데이트하며, 실패 시 에러를 반환합니다.
	Update(ctx context.Context, user *domain.User) error

	// Delete는 사용자 ID로 사용자 엔티티를 삭제합니다.
	// 논리적 삭제(상태를 DELETED로 변경) 또는 물리적 삭제를 구현 가능하며, 실패 시 에러를 반환합니다.
	Delete(ctx context.Context, id string) error
}
