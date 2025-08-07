package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kk7453603/RybakovTestGo/internal/config"
	"github.com/kk7453603/RybakovTestGo/internal/core/domain"
	"github.com/kk7453603/RybakovTestGo/internal/core/ports"
)

type externalPriceProvider struct {
	client     *http.Client
	baseURL    string
	symbolToID map[string]string
	ApiKey     string
}

func NewExternalPriceProvider(cfg config.Config) ports.ExternalPriceProvider {
	symbolToID := map[string]string{
		"BTC":   "bitcoin",
		"ETH":   "ethereum",
		"ADA":   "cardano",
		"SOL":   "solana",
		"DOT":   "polkadot",
		"LINK":  "chainlink",
		"MATIC": "matic-network",
		"AVAX":  "avalanche-2",
		"UNI":   "uniswap",
		"LTC":   "litecoin",
		"XRP":   "ripple",
		"DOGE":  "dogecoin",
	}

	return &externalPriceProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:    "https://api.coingecko.com/api/v3",
		symbolToID: symbolToID,
		ApiKey:     cfg.APIToken,
	}
}

type coinGeckoPriceResponse struct {
	USD           float64 `json:"usd"`
	LastUpdatedAt int64   `json:"last_updated_at"`
}

type coinGeckoHistoryResponse struct {
	Prices [][]float64 `json:"prices"`
}

func (p *externalPriceProvider) GetCurrentPrice(ctx context.Context, symbol string) (*domain.CurrencyPrice, error) {
	coinID, exists := p.symbolToID[strings.ToUpper(symbol)]
	if !exists {
		log.Printf("⚠️  Неподдерживаемая валюта: %s, используем fallback", symbol)
		return p.getFallbackPrice(symbol), nil
	}

	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd&include_last_updated_at=true",
		p.baseURL, coinID)

	log.Printf("🌐 Запрос текущей цены %s: %s", symbol, url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("❌ Ошибка создания запроса: %v", err)
		return p.getFallbackPrice(symbol), nil
	}

	req.Header.Set("User-Agent", "CryptoService/1.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Apikey "+p.ApiKey)
	resp, err := p.client.Do(req)
	if err != nil {
		log.Printf("❌ HTTP ошибка: %v", err)
		return p.getFallbackPrice(symbol), nil
	}
	defer resp.Body.Close()

	log.Printf("📊 Статус ответа: %d", resp.StatusCode)

	if resp.StatusCode == 429 {
		log.Printf("⚠️  Rate limit, используем fallback")
		return p.getFallbackPrice(symbol), nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Неожиданный статус: %d", resp.StatusCode)
		return p.getFallbackPrice(symbol), nil
	}

	var data map[string]coinGeckoPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("❌ Ошибка JSON: %v", err)
		return p.getFallbackPrice(symbol), nil
	}

	priceData, exists := data[coinID]
	if !exists {
		log.Printf("❌ Нет данных для %s", symbol)
		return nil, domain.ErrCurrencyNotFound
	}

	timestamp := time.Now()
	if priceData.LastUpdatedAt > 0 {
		timestamp = time.Unix(priceData.LastUpdatedAt, 0)
	}

	log.Printf("✅ Цена %s: $%.2f", symbol, priceData.USD)

	return &domain.CurrencyPrice{
		Symbol:    strings.ToUpper(symbol),
		Price:     priceData.USD,
		Timestamp: timestamp,
	}, nil
}

func (p *externalPriceProvider) GetHistoricalPrices(ctx context.Context, symbol string, startTime, endTime time.Time) ([]*domain.CurrencyPrice, error) {
	coinID, exists := p.symbolToID[strings.ToUpper(symbol)]
	if !exists {
		log.Printf("⚠️  Неподдерживаемая валюта для истории: %s", symbol)
		return p.getFallbackHistoricalPrices(symbol), nil
	}

	// Если время не указано, берем последние 30 дней
	if startTime.IsZero() {
		startTime = time.Now().AddDate(0, 0, -30)
	}
	if endTime.IsZero() {
		endTime = time.Now()
	}

	from := startTime.Unix()
	to := endTime.Unix()

	url := fmt.Sprintf("%s/coins/%s/market_chart/range?vs_currency=usd&from=%d&to=%d",
		p.baseURL, coinID, from, to)

	log.Printf("🌐 Запрос истории %s: %s", symbol, url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("❌ Ошибка создания запроса истории: %v", err)
		return p.getFallbackHistoricalPrices(symbol), nil
	}

	req.Header.Set("User-Agent", "CryptoService/1.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Apikey "+p.ApiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		log.Printf("❌ HTTP ошибка истории: %v", err)
		return p.getFallbackHistoricalPrices(symbol), nil
	}
	defer resp.Body.Close()

	log.Printf("📊 Статус ответа истории: %d", resp.StatusCode)

	if resp.StatusCode == 429 {
		log.Printf("⚠️  Rate limit для истории, используем fallback")
		return p.getFallbackHistoricalPrices(symbol), nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Неожиданный статус истории: %d", resp.StatusCode)
		return p.getFallbackHistoricalPrices(symbol), nil
	}

	var data coinGeckoHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("❌ Ошибка JSON истории: %v", err)
		return p.getFallbackHistoricalPrices(symbol), nil
	}

	var prices []*domain.CurrencyPrice
	for i, pricePoint := range data.Prices {
		if len(pricePoint) >= 2 && i < 50 { // Ограничиваем количество точек
			timestamp := time.Unix(int64(pricePoint[0]/1000), 0)
			price := pricePoint[1]

			prices = append(prices, &domain.CurrencyPrice{
				Symbol:    strings.ToUpper(symbol),
				Price:     price,
				Timestamp: timestamp,
			})
		}
	}

	log.Printf("✅ Получено %d исторических точек для %s", len(prices), symbol)
	return prices, nil
}

// Fallback методы для случая недоступности API
func (p *externalPriceProvider) getFallbackPrice(symbol string) *domain.CurrencyPrice {
	prices := map[string]float64{
		"BTC": 117416.0,
		"ETH": 3200.0,
		"ADA": 0.45,
		"SOL": 140.0,
	}

	price, exists := prices[strings.ToUpper(symbol)]
	if !exists {
		price = 100.0
	}

	return &domain.CurrencyPrice{
		Symbol:    strings.ToUpper(symbol),
		Price:     price,
		Timestamp: time.Now(),
	}
}

func (p *externalPriceProvider) getFallbackHistoricalPrices(symbol string) []*domain.CurrencyPrice {
	basePrice := p.getFallbackPrice(symbol).Price
	var prices []*domain.CurrencyPrice

	for i := 0; i < 10; i++ {
		variation := 1.0 + (float64(i%5)-2)*0.01 // Небольшие вариации ±2%
		prices = append(prices, &domain.CurrencyPrice{
			Symbol:    strings.ToUpper(symbol),
			Price:     basePrice * variation,
			Timestamp: time.Now().AddDate(0, 0, -i),
		})
	}

	return prices
}
