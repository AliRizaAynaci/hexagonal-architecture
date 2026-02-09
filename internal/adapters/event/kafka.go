package event

import (
	"encoding/json"
	"fmt"
	"hexagonal/internal/core/domain"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// KafkaPublisher - Gerçek Kafka bağlantısını tutan struct.
type KafkaPublisher struct {
	producer sarama.SyncProducer // SyncProducer: Mesaj gitmezse hata döner (Güvenli).
	topic    string
}

// NewKafkaPublisher - Kafka bağlantısını başlatır.
func NewKafkaPublisher(brokers []string, topic string) (*KafkaPublisher, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true          // Başarılı olursa haber ver
	config.Producer.RequiredAcks = sarama.WaitForAll // En güvenli mod (Veri kaybolmaz)
	config.Producer.Retry.Max = 5                    // Hata olursa 5 kere dene

	// Producer'ı oluştur (Bağlantı burada kurulur)
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		// Hata fırlatıyoruz, main.go bunu yakalayıp zap.L().Fatal basacak.
		// fmt.Errorf burada hatayı "wrap" (sarmalamak) için kullanılıyor, ekrana basmak için değil.
		return nil, fmt.Errorf("failed to start kafka producer: %w", err)
	}

	// Bağlantı başarılı logu
	zap.L().Info("kafka producer connected successfully",
		zap.Strings("brokers", brokers),
		zap.String("topic", topic),
	)

	return &KafkaPublisher{
		producer: producer,
		topic:    topic,
	}, nil
}

// kafkaEventDTO - DTO (Data Transfer Object): Kafka'ya gidecek JSON formatı.
type kafkaEventDTO struct {
	EventID   string    `json:"event_id"`
	ConcertID string    `json:"concert_id"`
	Quantity  int       `json:"qty"`
	SentAt    time.Time `json:"sent_at"`
}

// PublishTicketSold - Interface implementasyonu.
func (k *KafkaPublisher) PublishTicketSold(event domain.TicketSoldEvent) error {
	// 1. Domain -> DTO Dönüşümü
	dto := kafkaEventDTO{
		EventID:   event.EventID,
		ConcertID: event.ConcertID,
		Quantity:  event.Quantity,
		SentAt:    event.Timestamp,
	}

	// 2. JSON'a çevir
	val, err := json.Marshal(dto)
	if err != nil {
		zap.L().Error("failed to marshal event dto", zap.Error(err))
		return err
	}

	// 3. Kafka Mesajını Hazırla
	msg := &sarama.ProducerMessage{
		Topic: k.topic,
		// Key: Partitioning için önemli (Aynı konseri aynı partition'a atar ki sıra bozulmasın)
		Key:   sarama.StringEncoder(event.ConcertID),
		Value: sarama.ByteEncoder(val),
	}

	// 4. Gönder!
	partition, offset, err := k.producer.SendMessage(msg)
	if err != nil {
		zap.L().Error("failed to send message to kafka",
			zap.String("topic", k.topic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send message to kafka: %w", err)
	}

	// 5. Başarılı gönderim logu (fmt yerine zap kullandık)
	zap.L().Info("event published to kafka",
		zap.String("event_id", event.EventID),
		zap.String("topic", k.topic),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
	)

	return nil
}

// Close - Uygulama kapanırken bağlantıyı kapatmak için.
func (k *KafkaPublisher) Close() error {
	zap.L().Info("closing kafka producer...")
	return k.producer.Close()
}
