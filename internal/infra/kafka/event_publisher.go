package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

// EventPublisher는 Kafka를 통해 도메인 이벤트를 발행하는 구현입니다.
// domain/repo.EventPublisher 인터페이스를 만족하며, 비동기 이벤트 발행을 처리합니다.
type EventPublisher struct {
	producer *Producer
}

// NewEventPublisher는 새로운 EventPublisher 인스턴스를 생성하여 반환합니다.
// Kafka Producer를 주입받아 초기화합니다.
func NewEventPublisher(producer *Producer) *EventPublisher {
	return &EventPublisher{producer: producer}
}

// Publish는 도메인 이벤트를 Kafka에 발행합니다.
// 이벤트 타입별로 적절한 토픽에 JSON 직렬화된 메시지를 전송하며, 실패 시 에러를 반환합니다.
func (p *EventPublisher) Publish(ctx context.Context, event interface{}) error {
	// 이벤트 직렬화
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// 토픽 결정
	topic := p.producer.TopicForEvent(event)

	// Kafka 메시지 생성
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.ByteEncoder(data),
		Timestamp: time.Now().UTC(),
	}

	// 비동기 발행
	select {
	case p.producer.producer.Input() <- msg:
		// 성공 시 채널로 전송
	case <-ctx.Done():
		return ctx.Err() // 컨텍스트 타임아웃/취소
	}

	// 성공 확인 (비동기이므로 선택적, 여기선 로그만)
	go func() {
		select {
		case <-p.producer.producer.Successes():
			log.Printf("Successfully published event to %s", topic)
		case <-time.After(5 * time.Second):
			log.Printf("Timeout waiting for publish confirmation to %s", topic)
		}
	}()

	return nil
}

// BatchPublish는 여러 이벤트를 일괄 발행합니다.
// 성능 최적화를 위해 배치 처리를 지원하며, 하나라도 실패 시 에러를 반환합니다.
func (p *EventPublisher) BatchPublish(ctx context.Context, events []interface{}) error {
	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to batch publish events: %w", err)
		}
	}
	return nil
}

// Close는 이벤트 발행 리소스를 정리합니다.
// Kafka Producer를 종료하며, 대기 중인 메시지가 플러시되도록 합니다.
func (p *EventPublisher) Close() error {
	return p.producer.Close()
}
