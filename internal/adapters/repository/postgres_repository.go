package repository

import (
	"context"
	"time"

	"github.com/kk7453603/RybakovTestGo/internal/core/domain"
	"github.com/kk7453603/RybakovTestGo/internal/core/ports"
	"gorm.io/gorm"
)

// Модели для GORM
type CurrencyModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Symbol    string    `gorm:"uniqueIndex;not null;size:10"`
	Name      string    `gorm:"not null;size:100"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (CurrencyModel) TableName() string {
	return "currencies"
}

type CurrencyPriceModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Symbol    string    `gorm:"index;not null;size:10"`
	Price     float64   `gorm:"not null"`
	Timestamp time.Time `gorm:"index;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (CurrencyPriceModel) TableName() string {
	return "currency_prices"
}

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) (ports.CurrencyRepository, ports.PriceRepository) {
	repo := &postgresRepository{db: db}

	db.AutoMigrate(&CurrencyModel{}, &CurrencyPriceModel{})

	return repo, repo
}

func (r *postgresRepository) Create(ctx context.Context, currency *domain.Currency) error {
	model := &CurrencyModel{
		Symbol:    currency.Symbol,
		Name:      currency.Name,
		CreatedAt: currency.CreatedAt,
		UpdatedAt: currency.UpdatedAt,
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return domain.ErrDatabaseConnection
	}

	currency.ID = model.ID
	return nil
}

func (r *postgresRepository) GetBySymbol(ctx context.Context, symbol string) (*domain.Currency, error) {
	var model CurrencyModel
	result := r.db.WithContext(ctx).Where("symbol = ?", symbol).First(&model)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domain.ErrCurrencyNotFound
		}
		return nil, domain.ErrDatabaseConnection
	}

	return &domain.Currency{
		ID:        model.ID,
		Symbol:    model.Symbol,
		Name:      model.Name,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}

func (r *postgresRepository) List(ctx context.Context) ([]*domain.Currency, error) {
	var models []CurrencyModel
	result := r.db.WithContext(ctx).Find(&models)

	if result.Error != nil {
		return nil, domain.ErrDatabaseConnection
	}

	currencies := make([]*domain.Currency, len(models))
	for i, model := range models {
		currencies[i] = &domain.Currency{
			ID:        model.ID,
			Symbol:    model.Symbol,
			Name:      model.Name,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		}
	}

	return currencies, nil
}

func (r *postgresRepository) Delete(ctx context.Context, symbol string) error {
	result := r.db.WithContext(ctx).Where("symbol = ?", symbol).Delete(&CurrencyModel{})

	if result.Error != nil {
		return domain.ErrDatabaseConnection
	}

	if result.RowsAffected == 0 {
		return domain.ErrCurrencyNotFound
	}

	return nil
}

func (r *postgresRepository) Update(ctx context.Context, currency *domain.Currency) error {
	model := &CurrencyModel{
		Symbol:    currency.Symbol,
		Name:      currency.Name,
		UpdatedAt: time.Now(),
	}

	result := r.db.WithContext(ctx).Where("symbol = ?", currency.Symbol).Updates(model)

	if result.Error != nil {
		return domain.ErrDatabaseConnection
	}

	if result.RowsAffected == 0 {
		return domain.ErrCurrencyNotFound
	}

	return nil
}

func (r *postgresRepository) SavePrice(ctx context.Context, price *domain.CurrencyPrice) error {
	model := &CurrencyPriceModel{
		Symbol:    price.Symbol,
		Price:     price.Price,
		Timestamp: price.Timestamp,
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return domain.ErrDatabaseConnection
	}

	price.ID = model.ID
	return nil
}

func (r *postgresRepository) GetLatestPrice(ctx context.Context, symbol string) (*domain.CurrencyPrice, error) {
	var model CurrencyPriceModel
	result := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		Order("timestamp DESC").
		First(&model)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domain.ErrCurrencyNotFound
		}
		return nil, domain.ErrDatabaseConnection
	}

	return &domain.CurrencyPrice{
		ID:        model.ID,
		Symbol:    model.Symbol,
		Price:     model.Price,
		Timestamp: model.Timestamp,
	}, nil
}

func (r *postgresRepository) GetPriceByTimestamp(ctx context.Context, symbol string, timestamp time.Time) (*domain.CurrencyPrice, error) {
	var model CurrencyPriceModel
	result := r.db.WithContext(ctx).
		Where("symbol = ? AND timestamp <= ?", symbol, timestamp).
		Order("timestamp DESC").
		First(&model)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domain.ErrCurrencyNotFound
		}
		return nil, domain.ErrDatabaseConnection
	}

	return &domain.CurrencyPrice{
		ID:        model.ID,
		Symbol:    model.Symbol,
		Price:     model.Price,
		Timestamp: model.Timestamp,
	}, nil
}

func (r *postgresRepository) GetPriceHistory(ctx context.Context, symbol string, startTime, endTime time.Time, limit int) ([]*domain.CurrencyPrice, error) {
	var models []CurrencyPriceModel

	query := r.db.WithContext(ctx).
		Where("symbol = ? AND timestamp BETWEEN ? AND ?", symbol, startTime, endTime).
		Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&models)

	if result.Error != nil {
		return nil, domain.ErrDatabaseConnection
	}

	prices := make([]*domain.CurrencyPrice, len(models))
	for i, model := range models {
		prices[i] = &domain.CurrencyPrice{
			ID:        model.ID,
			Symbol:    model.Symbol,
			Price:     model.Price,
			Timestamp: model.Timestamp,
		}
	}

	return prices, nil
}
