package domain

import (
	"time"
)


type Currency struct {
	ID        int64     `json:"id"`
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}


type CurrencyPrice struct {
	ID        int64     `json:"id"`
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}


func (c *Currency) Validate() error {
	if c.Symbol == "" {
		return ErrInvalidCurrencySymbol
	}

	if c.Name == "" {
		return ErrInvalidCurrencyName
	}

	return nil
}


func (cp *CurrencyPrice) IsValidPrice() bool {
	return cp.Price > 0
}
