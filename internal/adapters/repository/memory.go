package repository

import (
	"errors"
	"hexagonal/internal/core/domain"
	"sync"
)

// InMemoryRepository - Veritabanı yerine geçecek geçici yapı.
// Neden sync.RWMutex var? -> Eşzamanlı (Concurrent) isteklerde map patlamasın diye.
type InMemoryRepository struct {
	db map[string]domain.Concert
	mu sync.RWMutex // Read Lock icin RWMutex kullanilmali
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		db: make(map[string]domain.Concert),
	}
}

// Get - veriyi okuma islemi
// Read Lock (RLock) kullanıyoruz çünkü okurken başkaları da okuyabilir, sadece yazan olmasın yeter.
func (r *InMemoryRepository) Get(id string) (*domain.Concert, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	concert, ok := r.db[id]
	if !ok {
		return nil, errors.New("concert not found")
	}

	return &concert, nil
}

// Write Lock (Lock) kullanıyoruz. Biz yazarken kimse okuyamaz ve yazamaz (Mutual Exclusion).
func (r *InMemoryRepository) Save(concert domain.Concert) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.db[concert.ID] = concert
	return nil
}
