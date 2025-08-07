package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kk7453603/RybakovTestGo/internal/core/domain"
	"github.com/kk7453603/RybakovTestGo/internal/core/ports"
)

// Внешний провайдер цен (например, CoinGecko API)
type externalPriceProvider struct {
	client  *http.Client
	baseURL string
}

// NewExternalPriceProvider создает новый провайдер внешних цен
func NewExternalPriceProvider() ports.ExternalPriceProvider {
	return &externalPriceProvider{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://api.coingecko.com/api/v3",
	}
}

type coinGeckoResponse struct {
	Price     float64 `json:"current_price"`
	Symbol    string  `json:"symbol"`
	UpdatedAt string  `json:"last_updated"`
}

func (p *externalPriceProvider) GetCurrentPrice(ctx context.Context, symbol string) (*domain.CurrencyPrice, error) {
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd&include_last_updated_at=true", p.baseURL, symbol)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, domain.ErrExternalAPIUnavailable
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, domain.ErrExternalAPIUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, domain.ErrExternalAPIUnavailable
	}

	var data map[string]coinGeckoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, domain.ErrExternalAPIUnavailable
	}

	priceData, exists := data[symbol]
	if !exists {
		return nil, domain.ErrCurrencyNotFound
	}

	return &domain.CurrencyPrice{
		Symbol:    symbol,
		Price:     priceData.Price,
		Timestamp: time.Now(),
	}, nil
}

func (p *externalPriceProvider) GetHistoricalPrices(ctx context.Context, symbol string, startTime, endTime time.Time) ([]*domain.CurrencyPrice, error) {
	// Реализация получения исторических данных
	// Для простоты возвращаем пустой массив
	// В реальном проекте здесь будет запрос к историческому API
	return []*domain.CurrencyPrice{}, nil
}
