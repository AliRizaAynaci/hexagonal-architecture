package event

import (
	"encoding/json"
	"fmt"
	"hexagonal/internal/core/domain"
	"time"
)

// Kafke yerine ekrana yazan adaptor
type ConsolePublisher struct{}

func NewConsolePublisher() *ConsolePublisher {
	return &ConsolePublisher{}
}

// eventDTO - Sadece bu adaptörün içinde yaşayan, dışarıya gönderilecek format.
// Domain nesnesini kirletmemek için buraya özel bir struct açtık.
type eventDTO struct {
	EventID   string    `json:"event_id"`
	ConcertID string    `json:"concert_id"`
	Quantity  int       `json:"qty"` // İstersem ismini 'qty' diye kısaltabilirim, Domain karışmaz.
	SentAt    time.Time `json:"sent_at"`
}

func (c *ConsolePublisher) PublishTicketSold(event domain.TicketSoldEvent) error {
	// MAPPING: Domain nesnesini -> Adapter DTO'suna çevir.
	// Buna "Mapping" veya "Transformation" denir.
	dto := eventDTO{
		EventID:   event.EventID,
		ConcertID: event.ConcertID,
		Quantity:  event.Quantity,
		SentAt:    event.Timestamp,
	}

	// Artık DTO'yu JSON'a çeviriyoruz.
	b, _ := json.MarshalIndent(dto, "", "  ") // Indent ile daha okunaklı olsun

	fmt.Printf("\n[EVENT BUS] >>> TicketSoldEvent Published:\n%s\n\n", string(b))
	return nil
}
