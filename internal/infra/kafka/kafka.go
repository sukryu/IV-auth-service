package kafka

import (
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/sukryu/IV-auth-services/internal/config"
	"github.com/sukryu/IV-auth-services/internal/domain"
)

// Producer는 Kafka 연결과 이벤트 발행을 관리하는 구조체입니다.
// sarama 라이브러리를 사용하여 비동기 프로듀서를 구현합니다.
type Producer struct {
	producer sarama.AsyncProducer
	config   *sarama.Config
}

// NewProducer는 새로운 Kafka Producer 인스턴스를 생성하여 반환합니다.
// config.Config에서 Kafka 브로커 설정을 받아 연결을 초기화합니다.
func NewProducer(cfg *config.Config) *Producer {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForLocal       // 로컬 브로커만 확인
	kafkaConfig.Producer.Retry.Max = 5                            // 최대 5회 재시도
	kafkaConfig.Producer.Retry.Backoff = 100 * time.Millisecond   // 재시도 간격 100ms
	kafkaConfig.Producer.Return.Successes = true                  // 성공 메시지 반환
	kafkaConfig.Producer.Return.Errors = true                     // 에러 메시지 반환
	kafkaConfig.Producer.Flush.Frequency = 500 * time.Millisecond // 500ms마다 플러시

	producer, err := sarama.NewAsyncProducer([]string{cfg.Kafka.Broker}, kafkaConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Kafka at %s: %v", cfg.Kafka.Broker, err)
	}

	// 비동기 에러 처리 고루틴
	go func() {
		for err := range producer.Errors() {
			log.Printf("Failed to publish message to Kafka: %v", err)
		}
	}()

	log.Printf("Connected to Kafka at %s", cfg.Kafka.Broker)
	return &Producer{
		producer: producer,
		config:   kafkaConfig,
	}
}

// Close는 Kafka 프로듀서를 안전하게 종료합니다.
// 리소스 정리를 위해 호출되며, 모든 대기 중인 메시지가 플러시됩니다.
func (p *Producer) Close() error {
	return p.producer.Close()
}

// TopicForEvent는 이벤트 타입에 따라 적절한 Kafka 토픽을 반환합니다.
// config.Kafka.TopicPrefix를 기반으로 토픽을 생성합니다.
func (p *Producer) TopicForEvent(event interface{}) string {
	prefix := "auth.events." // 기본값, config에서 가져오면 더 좋음
	switch event.(type) {
	case *domain.LoginSucceeded:
		return prefix + "login_succeeded"
	case *domain.LoginFailed:
		return prefix + "login_failed"
	case *domain.TokenBlacklisted:
		return prefix + "token_blacklisted"
	case *domain.UserCreated:
		return prefix + "user_created"
	case *domain.UserUpdated:
		return prefix + "user_updated"
	case *domain.UserDeleted:
		return prefix + "user_deleted"
	case *domain.PlatformConnected:
		return prefix + "platform_connected"
	case *domain.PlatformDisconnected:
		return prefix + "platform_disconnected"
	case *domain.PlatformTokenRefreshed:
		return prefix + "platform_token_refreshed"
	case *domain.PlatformConnectionFailed:
		return prefix + "platform_connection_failed"
	case *domain.PlatformTokenRefreshFailed:
		return prefix + "platform_token_refresh_failed"
	default:
		return prefix + "unknown"
	}
}
