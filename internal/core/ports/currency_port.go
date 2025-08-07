package ports

import (
	"context"
	"time"

	"github.com/kk7453603/RybakovTestGo/internal/core/domain"
)


type CurrencyService interface {
	AddCurrency(ctx context.Context, symbol, name string) (*domain.Currency, error)
	RemoveCurrency(ctx context.Context, symbol string) error
	GetCurrencyPrice(ctx context.Context, symbol string, timestamp time.Time) (*domain.CurrencyPrice, error)
	ListCurrencies(ctx context.Context) ([]*domain.Currency, error)
	GetPriceHistory(ctx context.Context, symbol string, startTime, endTime time.Time, limit int) ([]*domain.CurrencyPrice, error)
}
