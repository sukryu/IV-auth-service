package events

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sukryu/IV-auth-services/internal/config"
	"github.com/sukryu/IV-auth-services/internal/core/domain"
	"github.com/sukryu/IV-auth-services/pkg/logger"
	"go.uber.org/zap"
)

// KafkaEventPublisher implements domain.EventPublisher for Kafka.
type KafkaEventPublisher struct {
	writer *kafka.Writer
	logger *logger.Logger
	mutex  sync.Mutex
}

// NewKafkaEventPublisher creates a new KafkaEventPublisher instance.
func NewKafkaEventPublisher(cfg *config.Config, log *logger.Logger) domain.EventPublisher {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Kafka.Broker),
		Topic:        "auth.events", // 단일 토픽 사용, 필요 시 동적 설정 가능
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond, // 성능 최적화
		RequiredAcks: kafka.RequireOne,      // 최소 1개 브로커 확인
	}

	return &KafkaEventPublisher{
		writer: writer,
		logger: log.With(zap.String("component", "kafka_event_publisher")),
	}
}

// Publish sends an event to the Kafka topic.
func (p *KafkaEventPublisher) Publish(event domain.Event) error {
	if event == nil {
		return errors.New("event must not be nil")
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 이벤트 직렬화
	payload, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("Failed to marshal event", zap.Error(err), zap.String("event_name", event.EventName()))
		return errors.New("failed to marshal event: " + err.Error())
	}

	// Kafka 메시지 생성
	msg := kafka.Message{
		Key:   []byte(event.EventName()), // 이벤트 이름으로 키 설정
		Value: payload,
	}

	// 메시지 발행
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Error("Failed to publish event to Kafka", zap.Error(err), zap.String("event_name", event.EventName()))
		return errors.New("failed to publish event: " + err.Error())
	}

	p.logger.Debug("Event published successfully", zap.String("event_name", event.EventName()))
	return nil
}

// Close shuts down the Kafka writer.
func (p *KafkaEventPublisher) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if err := p.writer.Close(); err != nil {
		p.logger.Error("Failed to close Kafka writer", zap.Error(err))
		return errors.New("failed to close Kafka writer: " + err.Error())
	}
	p.logger.Info("Kafka writer closed successfully")
	return nil
}
