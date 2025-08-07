package domain

import "errors"

var (
	// ErrCurrencyNotFound возвращается когда криптовалюта не найдена
	ErrCurrencyNotFound = errors.New("currency not found")

	// ErrCurrencyAlreadyExists возвращается при попытке добавить существующую криптовалюту
	ErrCurrencyAlreadyExists = errors.New("currency already exists")

	// ErrInvalidCurrencySymbol возвращается при некорректном символе
	ErrInvalidCurrencySymbol = errors.New("invalid currency symbol")

	// ErrInvalidCurrencyName возвращается при некорректном названии
	ErrInvalidCurrencyName = errors.New("invalid currency name")

	// ErrInvalidPrice возвращается при некорректной цене
	ErrInvalidPrice = errors.New("invalid price")

	// ErrDatabaseConnection возвращается при ошибке подключения к БД
	ErrDatabaseConnection = errors.New("database connection error")

	// ErrExternalAPIUnavailable возвращается при недоступности внешнего API
	ErrExternalAPIUnavailable = errors.New("external API unavailable")
)
