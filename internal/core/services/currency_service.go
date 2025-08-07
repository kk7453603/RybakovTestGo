package services

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/kk7453603/RybakovTestGo/internal/core/domain"
	"github.com/kk7453603/RybakovTestGo/internal/core/ports"
)

type currencyService struct {
	currencyRepo  ports.CurrencyRepository
	priceRepo     ports.PriceRepository
	priceProvider ports.ExternalPriceProvider
}

func NewCurrencyService(
	currencyRepo ports.CurrencyRepository,
	priceRepo ports.PriceRepository,
	priceProvider ports.ExternalPriceProvider,
) ports.CurrencyService {
	return &currencyService{
		currencyRepo:  currencyRepo,
		priceRepo:     priceRepo,
		priceProvider: priceProvider,
	}
}

func (s *currencyService) AddCurrency(ctx context.Context, symbol, name string) (*domain.Currency, error) {
	symbol = strings.ToUpper(symbol)
	existing, err := s.currencyRepo.GetBySymbol(ctx, symbol)
	if err == nil && existing != nil {
		return nil, domain.ErrCurrencyAlreadyExists
	}

	currency := &domain.Currency{
		Symbol:    symbol,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := currency.Validate(); err != nil {
		return nil, err
	}

	if err := s.currencyRepo.Create(ctx, currency); err != nil {
		return nil, err
	}

	if currentPrice, err := s.priceProvider.GetCurrentPrice(ctx, symbol); err == nil {
		_ = s.priceRepo.SavePrice(ctx, currentPrice)
	}

	return currency, nil
}

func (s *currencyService) RemoveCurrency(ctx context.Context, symbol string) error {
	existing, err := s.currencyRepo.GetBySymbol(ctx, symbol)
	if err != nil || existing == nil {
		return domain.ErrCurrencyNotFound
	}

	return s.currencyRepo.Delete(ctx, symbol)
}

func (s *currencyService) GetCurrencyPrice(ctx context.Context, symbol string, timestamp time.Time) (*domain.CurrencyPrice, error) {
	symbol = strings.ToUpper(symbol)
	_, err := s.currencyRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, domain.ErrCurrencyNotFound
	}

	if timestamp.IsZero() {
		return s.priceRepo.GetLatestPrice(ctx, symbol)
	}

	price, err := s.priceRepo.GetPriceByTimestamp(ctx, symbol, timestamp)
	if err != nil {
		return s.priceProvider.GetCurrentPrice(ctx, symbol)
	}

	return price, nil
}

func (s *currencyService) ListCurrencies(ctx context.Context) ([]*domain.Currency, error) {
	return s.currencyRepo.List(ctx)
}

func (s *currencyService) GetPriceHistory(ctx context.Context, symbol string, startTime, endTime time.Time, limit int) ([]*domain.CurrencyPrice, error) {
	_, err := s.currencyRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, err
	}

	prices, err := s.priceRepo.GetPriceHistory(ctx, symbol, startTime, endTime, limit)

	if err != nil || len(prices) == 0 {
		log.Printf("📡 Запрашиваем исторические данные у внешнего провайдера для %s", symbol)
		externalPrices, extErr := s.priceProvider.GetHistoricalPrices(ctx, symbol, startTime, endTime)
		if extErr == nil && len(externalPrices) > 0 {
			for _, price := range externalPrices {
				s.priceRepo.SavePrice(ctx, price)
			}
			return externalPrices[:min(len(externalPrices), limit)], nil
		}
	}

	return prices, err
}
