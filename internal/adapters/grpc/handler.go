package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kk7453603/RybakovTestGo/internal/core/domain"
	"github.com/kk7453603/RybakovTestGo/internal/core/ports"
	currencyv1 "github.com/kk7453603/RybakovTestGo/pkg/api/gen"
)

type CurrencyHandler struct {
	currencyv1.UnimplementedCurrencyServiceServer
	service ports.CurrencyService
}

func NewCurrencyHandler(service ports.CurrencyService) *CurrencyHandler {
	return &CurrencyHandler{
		service: service,
	}
}

func (h *CurrencyHandler) AddCurrency(ctx context.Context, req *currencyv1.AddCurrencyRequest) (*currencyv1.CurrencyResponse, error) {
	if req.Symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	currency, err := h.service.AddCurrency(ctx, req.Symbol, req.Name)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &currencyv1.CurrencyResponse{
		Currency: h.domainToProtoCurrency(currency),
	}, nil
}

func (h *CurrencyHandler) RemoveCurrency(ctx context.Context, req *currencyv1.RemoveCurrencyRequest) (*emptypb.Empty, error) {
	if req.Symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}

	err := h.service.RemoveCurrency(ctx, req.Symbol)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *CurrencyHandler) GetCurrencyPrice(ctx context.Context, req *currencyv1.GetCurrencyPriceRequest) (*currencyv1.CurrencyPriceResponse, error) {
	if req.Symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}

	var timestamp time.Time
	if req.Timestamp != nil {
		timestamp = req.Timestamp.AsTime()
	}

	price, err := h.service.GetCurrencyPrice(ctx, req.Symbol, timestamp)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &currencyv1.CurrencyPriceResponse{
		Price: h.domainToProtoCurrencyPrice(price),
	}, nil
}

func (h *CurrencyHandler) ListCurrencies(ctx context.Context, req *emptypb.Empty) (*currencyv1.ListCurrenciesResponse, error) {
	currencies, err := h.service.ListCurrencies(ctx)
	if err != nil {
		return nil, h.handleError(err)
	}

	protoCurrencies := make([]*currencyv1.Currency, len(currencies))
	for i, currency := range currencies {
		protoCurrencies[i] = h.domainToProtoCurrency(currency)
	}

	return &currencyv1.ListCurrenciesResponse{
		Currencies: protoCurrencies,
	}, nil
}

func (h *CurrencyHandler) GetPriceHistory(ctx context.Context, req *currencyv1.GetPriceHistoryRequest) (*currencyv1.PriceHistoryResponse, error) {
	if req.Symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}

	var startTime, endTime time.Time
	if req.StartTime != nil {
		startTime = req.StartTime.AsTime()
	}
	if req.EndTime != nil {
		endTime = req.EndTime.AsTime()
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 100
	}

	prices, err := h.service.GetPriceHistory(ctx, req.Symbol, startTime, endTime, limit)
	if err != nil {
		return nil, h.handleError(err)
	}

	protoPrices := make([]*currencyv1.CurrencyPrice, len(prices))
	for i, price := range prices {
		protoPrices[i] = h.domainToProtoCurrencyPrice(price)
	}

	return &currencyv1.PriceHistoryResponse{
		Prices: protoPrices,
	}, nil
}

func (h *CurrencyHandler) domainToProtoCurrency(currency *domain.Currency) *currencyv1.Currency {
	return &currencyv1.Currency{
		Id:        currency.ID,
		Symbol:    currency.Symbol,
		Name:      currency.Name,
		CreatedAt: timestamppb.New(currency.CreatedAt),
		UpdatedAt: timestamppb.New(currency.UpdatedAt),
	}
}

func (h *CurrencyHandler) domainToProtoCurrencyPrice(price *domain.CurrencyPrice) *currencyv1.CurrencyPrice {
	return &currencyv1.CurrencyPrice{
		Id:        price.ID,
		Symbol:    price.Symbol,
		Price:     price.Price,
		Timestamp: timestamppb.New(price.Timestamp),
	}
}

func (h *CurrencyHandler) handleError(err error) error {
	switch err {
	case domain.ErrCurrencyNotFound:
		return status.Error(codes.NotFound, err.Error())
	case domain.ErrCurrencyAlreadyExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case domain.ErrInvalidCurrencySymbol, domain.ErrInvalidCurrencyName, domain.ErrInvalidPrice:
		return status.Error(codes.InvalidArgument, err.Error())
	case domain.ErrDatabaseConnection:
		return status.Error(codes.Internal, "internal server error")
	case domain.ErrExternalAPIUnavailable:
		return status.Error(codes.Unavailable, "external service unavailable")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
