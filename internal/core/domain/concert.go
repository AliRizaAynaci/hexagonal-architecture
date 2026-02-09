package domain

import (
	"errors"
	"time"
)

// NOTE: Service katmani hatatnin ne oldugunu anlayip ona gore HTTP 409 vs donmesi icin
var ErrCapacityExceeded = errors.New("concert capacity exceeded")

// Dikkat: Burada `json:"..."` veya `db:"..."` tagleri YOK.
// Çünkü Domain, veritabanından veya HTTP'den habersiz olmalı!
// Yarın öbür gün JSON yerine gRPC kullanırsak burası değişmemeli
type Concert struct {
	ID          string
	Name        string
	Capacity    int
	SoldTickets int
	Date        time.Time
}

// NewConcert - Bir Constructor
// Geçerli bir konser nesnesi oluşturmak için kuralları buraya koyarız.
// Örneğin: Kapasite negatif olamaz kontrolü buraya eklenebilir.
func NewConcert(id string, name string, capacity int, date time.Time) *Concert {
	return &Concert{
		ID:          id,
		Name:        name,
		Capacity:    capacity,
		SoldTickets: 0,
		Date:        date,
	}
}

// CanSell - Business Rule buradadır.
// Database'e bakmaz, sadece elindeki veriye göre karar verir.
func (c *Concert) CanSell(quantity int) error {
	if c.SoldTickets+quantity > c.Capacity {
		return ErrCapacityExceeded
	}
	return nil
}

func (c *Concert) Sell(quantity int) error {
	if c.SoldTickets+quantity > c.Capacity {
		return ErrCapacityExceeded
	}
	c.SoldTickets += quantity
	return nil
}
