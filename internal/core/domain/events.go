package domain

import "time"

type TicketSoldEvent struct {
	EventID   string
	ConcertID string
	Quantity  int
	Timestamp time.Time
}
