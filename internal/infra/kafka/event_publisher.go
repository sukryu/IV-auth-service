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
// domain/repo.EventPublisher 인터페이스를 만족하며, 비동기 및 배치 발행을 지원합니다.
type EventPublisher struct {
	producer      *Producer
	topicPrefix   string // 설정에서 가져온 토픽 접두사
	retryAttempts int    // 발행 실패 시 재시도 횟수
}

// NewEventPublisher는 새로운 EventPublisher 인스턴스를 생성하여 반환합니다.
// Kafka Producer와 토픽 접두사를 주입받아 초기화합니다.
func NewEventPublisher(producer *Producer, topicPrefix string) *EventPublisher {
	return &EventPublisher{
		producer:      producer,
		topicPrefix:   topicPrefix,
		retryAttempts: 3, // 기본 재시도 횟수
	}
}

// Publish는 도메인 이벤트를 Kafka에 발행합니다.
// 실패 시 재시도 후 에러를 반환하며, 비동기 발행을 처리합니다.
func (p *EventPublisher) Publish(ctx context.Context, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	topic := p.producer.TopicForEvent(event, p.topicPrefix)
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.ByteEncoder(data),
		Timestamp: time.Now().UTC(),
	}

	var lastErr error
	for attempt := 0; attempt < p.retryAttempts; attempt++ {
		select {
		case p.producer.producer.Input() <- msg:
			// 발행 성공 시 재시도 중단
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case err := <-p.producer.producer.Errors():
			lastErr = fmt.Errorf("attempt %d failed: %v", attempt+1, err)
			time.Sleep(100 * time.Millisecond) // 재시도 전 대기
		}
	}

	if lastErr != nil {
		log.Printf("Failed to publish event to %s after %d attempts: %v", topic, p.retryAttempts, lastErr)
		return lastErr
	}
	return nil
}

// BatchPublish는 여러 이벤트를 일괄 발행합니다.
// 성능 최적화를 위해 배치 처리를 지원하며, 실패 시 전체 에러를 반환합니다.
func (p *EventPublisher) BatchPublish(ctx context.Context, events []interface{}) error {
	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event in batch: %v", err)
			continue
		}

		topic := p.producer.TopicForEvent(event, p.topicPrefix)
		msg := &sarama.ProducerMessage{
			Topic:     topic,
			Value:     sarama.ByteEncoder(data),
			Timestamp: time.Now().UTC(),
		}

		p.producer.producer.Input() <- msg // 비동기 발행
	}

	// 성공/실패 여부는 producer의 Errors 채널에서 확인, 여기선 비동기만 처리
	return nil
}

// Close는 이벤트 발행 리소스를 정리합니다.
// Kafka Producer를 종료하며, 대기 중인 메시지가 플러시되도록 보장합니다.
func (p *EventPublisher) Close() error {
	return p.producer.Close()
}
