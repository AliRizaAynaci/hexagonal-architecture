package ports

import "hexagonal/internal/core/domain"

// Driven Port (Secondary): Veritabanı işlemleri için gerekli sözleşme.
// Domain katmanı veritabanı detayını bilmez, sadece bu metodları bilir.
// İleride buraya Postgres, Mongo veya In-Memory ne bağlarsan bağla, Domain değişmez.
type ConcertRepository interface {
	Get(id string) (*domain.Concert, error)
	Save(concert domain.Concert) error
}

// Driving Port (Primary): Dış dünyanın (HTTP, CLI, gRPC) kullanacağı servis sözleşmesi.
// Service katmanı bu interface'i implemente edecek.
type ConcertService interface {
	CreateConcert(id string, name string, capacity int) (*domain.Concert, error)
	BuyTicket(concertID string, quantity int) error
}

// Service katmanı bu arayüzü kullanarak event fırlatır.
// Arkada Kafka mı var, RabbitMQ mu var bilmez.
type EventPublisher interface {
	PublishTicketSold(event domain.TicketSoldEvent) error
}
