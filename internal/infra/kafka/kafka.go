package kafka

import (
	"context"
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
	ctx      context.Context    // 고루틴 제어용 컨텍스트
	cancel   context.CancelFunc // 컨텍스트 취소 함수
}

// NewProducer는 새로운 Kafka Producer 인스턴스를 생성하여 반환합니다.
// config.Config에서 Kafka 브로커 설정을 받아 연결을 초기화합니다.
func NewProducer(cfg *config.Config) *Producer {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForLocal
	kafkaConfig.Producer.Retry.Max = 5
	kafkaConfig.Producer.Retry.Backoff = 100 * time.Millisecond
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.Return.Errors = true
	kafkaConfig.Producer.Flush.Frequency = 500 * time.Millisecond

	producer, err := sarama.NewAsyncProducer([]string{cfg.Kafka.Broker}, kafkaConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Kafka at %s: %v", cfg.Kafka.Broker, err)
	}

	// 고루틴 제어용 컨텍스트 생성
	ctx, cancel := context.WithCancel(context.Background())

	// 비동기 에러 처리 및 성공 로그
	go func() {
		defer log.Printf("Kafka producer goroutine terminated")
		for {
			select {
			case success, ok := <-producer.Successes():
				if !ok {
					return
				}
				log.Printf("Message published to %s, partition: %d, offset: %d", success.Topic, success.Partition, success.Offset)
			case err, ok := <-producer.Errors():
				if !ok {
					return
				}
				if err != nil {
					log.Printf("Failed to publish message to Kafka: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Printf("Connected to Kafka at %s", cfg.Kafka.Broker)
	return &Producer{
		producer: producer,
		config:   kafkaConfig,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Close는 Kafka 프로듀서를 안전하게 종료합니다.
// 모든 대기 중인 메시지가 플러시되며, 고루틴이 종료됩니다.
func (p *Producer) Close() error {
	p.cancel() // 고루틴 종료 신호
	return p.producer.Close()
}

// TopicForEvent는 이벤트 타입에 따라 적절한 Kafka 토픽을 반환합니다.
// config.Kafka.TopicPrefix를 기반으로 토픽을 생성합니다.
func (p *Producer) TopicForEvent(event interface{}, prefix string) string {
	switch event.(type) {
	case domain.LoginSucceeded:
		return prefix + "login_succeeded"
	case domain.LoginFailed:
		return prefix + "login_failed"
	case domain.TokenBlacklisted:
		return prefix + "token_blacklisted"
	case domain.UserCreated:
		return prefix + "user_created"
	case domain.UserUpdated:
		return prefix + "user_updated"
	case domain.UserDeleted:
		return prefix + "user_deleted"
	case domain.PlatformConnected:
		return prefix + "platform_connected"
	case domain.PlatformDisconnected:
		return prefix + "platform_disconnected"
	case domain.PlatformTokenRefreshed:
		return prefix + "platform_token_refreshed"
	case domain.PlatformConnectionFailed:
		return prefix + "platform_connection_failed"
	case domain.PlatformTokenRefreshFailed:
		return prefix + "platform_token_refresh_failed"
	default:
		log.Printf("Unknown event type: %T", event)
		return prefix + "unknown"
	}
}
