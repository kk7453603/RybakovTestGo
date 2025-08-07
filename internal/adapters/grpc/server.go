package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/kk7453603/RybakovTestGo/internal/core/ports"
	currencyv1 "github.com/kk7453603/RybakovTestGo/pkg/api/gen"
)

// GRPCServer представляет gRPC сервер
type GRPCServer struct {
	server   *grpc.Server
	handler  *CurrencyHandler
	grpcPort int
	httpPort int
}

// NewGRPCServer создает новый экземпляр gRPC сервера
func NewGRPCServer(service ports.CurrencyService, grpcPort, httpPort int) *GRPCServer {
	server := grpc.NewServer()
	handler := NewCurrencyHandler(service)

	// Регистрируем сервис
	currencyv1.RegisterCurrencyServiceServer(server, handler)

	// Включаем reflection для grpcurl и других инструментов
	reflection.Register(server)

	return &GRPCServer{
		server:   server,
		handler:  handler,
		grpcPort: grpcPort,
		httpPort: httpPort,
	}
}

// Start запускает gRPC сервер
func (s *GRPCServer) Start() error {
	// Запускаем gRPC сервер в горутине
	go func() {
		if err := s.startGRPCServer(); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Запускаем HTTP gateway
	return s.startHTTPGateway()
}

// startGRPCServer запускает gRPC сервер
func (s *GRPCServer) startGRPCServer() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.grpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.grpcPort, err)
	}

	log.Printf("gRPC server listening on port %d", s.grpcPort)
	return s.server.Serve(listener)
}

// startHTTPGateway запускает HTTP gateway
func (s *GRPCServer) startHTTPGateway() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Создаем mux для gRPC gateway
	mux := runtime.NewServeMux()

	// Подключаемся к gRPC серверу
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := currencyv1.RegisterCurrencyServiceHandlerFromEndpoint(
		ctx,
		mux,
		fmt.Sprintf("localhost:%d", s.grpcPort),
		opts,
	)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	// Добавляем CORS middleware
	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			h.ServeHTTP(w, r)
		})
	}

	log.Printf("HTTP gateway listening on port %d", s.httpPort)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.httpPort), corsHandler(mux))
}

// Stop останавливает сервер
func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}
