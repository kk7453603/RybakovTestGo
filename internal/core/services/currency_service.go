package services

import (
	"context"
	"time"

	"github.com/kk7453603/RybakovTestGo/internal/core/domain"
	"github.com/kk7453603/RybakovTestGo/internal/core/ports"
)

type currencyService struct {
	currencyRepo  ports.CurrencyRepository
	priceRepo     ports.PriceRepository
	priceProvider ports.ExternalPriceProvider
}

// NewCurrencyService создает новый экземпляр сервиса
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
	// Проверяем, не существует ли уже такая криптовалюта
	existing, err := s.currencyRepo.GetBySymbol(ctx, symbol)
	if err == nil && existing != nil {
		return nil, domain.ErrCurrencyAlreadyExists
	}

	// Создаем новую криптовалюту
	currency := &domain.Currency{
		Symbol:    symbol,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Валидируем данные
	if err := currency.Validate(); err != nil {
		return nil, err
	}

	// Сохраняем в базу данных
	if err := s.currencyRepo.Create(ctx, currency); err != nil {
		return nil, err
	}

	// Пытаемся получить текущую цену и сохранить ее
	if currentPrice, err := s.priceProvider.GetCurrentPrice(ctx, symbol); err == nil {
		_ = s.priceRepo.SavePrice(ctx, currentPrice)
	}

	return currency, nil
}

func (s *currencyService) RemoveCurrency(ctx context.Context, symbol string) error {
	// Проверяем существование криптовалюты
	existing, err := s.currencyRepo.GetBySymbol(ctx, symbol)
	if err != nil || existing == nil {
		return domain.ErrCurrencyNotFound
	}

	return s.currencyRepo.Delete(ctx, symbol)
}

func (s *currencyService) GetCurrencyPrice(ctx context.Context, symbol string, timestamp time.Time) (*domain.CurrencyPrice, error) {
	// Проверяем, что криптовалюта отслеживается
	_, err := s.currencyRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, domain.ErrCurrencyNotFound
	}

	// Если timestamp не указан, возвращаем последнюю цену
	if timestamp.IsZero() {
		return s.priceRepo.GetLatestPrice(ctx, symbol)
	}

	// Ищем цену по конкретному времени
	price, err := s.priceRepo.GetPriceByTimestamp(ctx, symbol, timestamp)
	if err != nil {
		// Если не найдено в локальной БД, пытаемся получить из внешнего API
		return s.priceProvider.GetCurrentPrice(ctx, symbol)
	}

	return price, nil
}

func (s *currencyService) ListCurrencies(ctx context.Context) ([]*domain.Currency, error) {
	return s.currencyRepo.List(ctx)
}

func (s *currencyService) GetPriceHistory(ctx context.Context, symbol string, startTime, endTime time.Time, limit int) ([]*domain.CurrencyPrice, error) {
	// Проверяем, что криптовалюта отслеживается
	_, err := s.currencyRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, domain.ErrCurrencyNotFound
	}

	// Получаем исторические данные из локальной БД
	localPrices, err := s.priceRepo.GetPriceHistory(ctx, symbol, startTime, endTime, limit)
	if err != nil || len(localPrices) == 0 {
		// Если данных нет локально, пытаемся получить из внешнего API
		externalPrices, extErr := s.priceProvider.GetHistoricalPrices(ctx, symbol, startTime, endTime)
		if extErr != nil {
			return nil, extErr
		}

		// Сохраняем полученные данные в локальную БД
		for _, price := range externalPrices {
			_ = s.priceRepo.SavePrice(ctx, price)
		}

		return externalPrices, nil
	}

	return localPrices, nil
}
