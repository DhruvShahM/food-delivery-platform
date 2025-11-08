package kafka

import (
	"context"
	"encoding/json"

	"food-delivery-platform/common/proto"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func Publish(w *kafka.Writer, evt *proto.OrderEvent, logger *zap.Logger) error {
	data, err := json.Marshal(evt)
	if err != nil {
		logger.Error("Marshal error", zap.Error(err))
		return err
	}
	msg := kafka.Message{Value: data}
	return w.WriteMessages(context.Background(), msg)
}

func NewReader(brokers []string, topic string, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
}

func ConsumeLoop(r *kafka.Reader, fn func(*kafka.Message) error, logger *zap.Logger) {
	defer r.Close()
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			logger.Error("Read error", zap.Error(err))
			break
		}
		if err := fn(&m); err != nil {
			logger.Error("Process error", zap.Error(err))
		}
	}
}
