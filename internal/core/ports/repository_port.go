package ports

import (
	"context"
	"time"

	"github.com/kk7453603/RybakovTestGo/internal/core/domain"
)

type CurrencyRepository interface {
	Create(ctx context.Context, currency *domain.Currency) error
	GetBySymbol(ctx context.Context, symbol string) (*domain.Currency, error)
	List(ctx context.Context) ([]*domain.Currency, error)
	Delete(ctx context.Context, symbol string) error
	Update(ctx context.Context, currency *domain.Currency) error
}

type PriceRepository interface {
	SavePrice(ctx context.Context, price *domain.CurrencyPrice) error
	GetLatestPrice(ctx context.Context, symbol string) (*domain.CurrencyPrice, error)
	GetPriceByTimestamp(ctx context.Context, symbol string, timestamp time.Time) (*domain.CurrencyPrice, error)
	GetPriceHistory(ctx context.Context, symbol string, startTime, endTime time.Time, limit int) ([]*domain.CurrencyPrice, error)
}

type ExternalPriceProvider interface {
	GetCurrentPrice(ctx context.Context, symbol string) (*domain.CurrencyPrice, error)
	GetHistoricalPrices(ctx context.Context, symbol string, startTime, endTime time.Time) ([]*domain.CurrencyPrice, error)
}
