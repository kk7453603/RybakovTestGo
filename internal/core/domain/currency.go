package domain

import (
	"time"
)

// Currency представляет доменную модель криптовалюты
type Currency struct {
	ID        int64     `json:"id"`
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CurrencyPrice представляет цену криптовалюты в определенный момент времени
type CurrencyPrice struct {
	ID        int64     `json:"id"`
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// ValidateCurrency выполняет валидацию данных криптовалюты
func (c *Currency) Validate() error {
	if c.Symbol == "" {
		return ErrInvalidCurrencySymbol
	}

	if c.Name == "" {
		return ErrInvalidCurrencyName
	}

	return nil
}

// IsValidPrice проверяет корректность цены
func (cp *CurrencyPrice) IsValidPrice() bool {
	return cp.Price > 0
}
