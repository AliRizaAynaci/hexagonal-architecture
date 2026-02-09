package services

import (
	"errors"
	"hexagonal/internal/core/domain"
	"hexagonal/internal/core/ports"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// service - ConcertService portunu implemente eden yapımız.
// Küçük harfle başlattık çünkü dışarıdan direkt erişilmesin, Constructor ile verilsin.
type service struct {
	repo      ports.ConcertRepository
	publisher ports.EventPublisher
}

// NewConcertService - Servis oluşturmak için kullanılan Constructor fonksiyonu.
// Bu fonksiyon bize bir ports.ConcertService döner.
func NewConcertService(r ports.ConcertRepository, p ports.EventPublisher) ports.ConcertService {
	return &service{
		repo:      r,
		publisher: p, // Publisher'ı içeri aldık
	}
}

func (s *service) CreateConcert(id string, name string, capacity int) (*domain.Concert, error) {
	newConcert := domain.NewConcert(id, name, capacity, time.Now().AddDate(0, 1, 0))

	if err := s.repo.Save(*newConcert); err != nil {
		zap.L().Error("failed to create concert",
			zap.String("concert_id", id),
			zap.Error(err),
		)
		return nil, err
	}

	zap.L().Info("concert created successfully",
		zap.String("concert_id", id),
		zap.String("name", name),
		zap.Int("capacity", capacity),
	)

	return newConcert, nil
}

func (s *service) BuyTicket(concertID string, quantity int) error {
	// 1. Repo işlemleri
	concert, err := s.repo.Get(concertID)
	if err != nil {
		zap.L().Error("failed to fetch concert", zap.String("id", concertID), zap.Error(err))
		return err
	}
	if concert == nil {
		return errors.New("concert not found")
	}

	// 2. Logic
	if err := concert.CanSell(quantity); err != nil {
		zap.L().Warn("ticket sale rejected: capacity full",
			zap.String("concert_id", concertID),
			zap.Int("requested", quantity),
			zap.Int("remaining", concert.Capacity-concert.SoldTickets),
		)
		return err
	}

	// 3. State Change
	concert.Sell(quantity)

	if err := s.repo.Save(*concert); err != nil {
		zap.L().Error("failed to save ticket sale", zap.String("id", concertID), zap.Error(err))
		return err
	}

	// 4. EVENT PUBLISHING
	// Satış başarılı oldu, şimdi Kafka'ya mesaj atıyoruz.
	event := domain.TicketSoldEvent{
		EventID:   uuid.New().String(),
		ConcertID: concertID,
		Quantity:  quantity,
		Timestamp: time.Now(),
	}

	// Event'i fırlat
	if err := s.publisher.PublishTicketSold(event); err != nil {
		// Mesaj gitmese bile satış iptal edilmez (Genelde), o yüzden sadece log basıyoruz.
		zap.L().Error("failed to publish event to kafka", zap.Error(err))
	} else {
		zap.L().Info("ticket sold and event published",
			zap.String("event_id", event.EventID),
			zap.String("concert_id", concertID),
		)
	}

	return nil
}
