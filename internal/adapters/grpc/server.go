package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/kk7453603/RybakovTestGo/internal/core/ports"
	currencyv1 "github.com/kk7453603/RybakovTestGo/pkg/api/gen"
)

// GRPCServer представляет gRPC сервер
type GRPCServer struct {
	server       *grpc.Server
	handler      *CurrencyHandler
	grpcPort     int
	httpPort     int
	grpcStarted  chan struct{}
	healthServer *health.Server
}

// NewGRPCServer создает новый экземпляр gRPC сервера
func NewGRPCServer(service ports.CurrencyService, grpcPort, httpPort int) *GRPCServer {
	server := grpc.NewServer()
	handler := NewCurrencyHandler(service)
	healthServer := health.NewServer()

	// Регистрируем сервис
	currencyv1.RegisterCurrencyServiceServer(server, handler)

	// Регистрируем health check
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Включаем reflection для grpcurl и других инструментов
	reflection.Register(server)

	return &GRPCServer{
		server:       server,
		handler:      handler,
		grpcPort:     grpcPort,
		httpPort:     httpPort,
		grpcStarted:  make(chan struct{}),
		healthServer: healthServer,
	}
}

// Start запускает gRPC сервер и HTTP gateway
func (s *GRPCServer) Start() error {
	var wg sync.WaitGroup
	var grpcErr, httpErr error

	// Запускаем gRPC сервер
	wg.Add(1)
	go func() {
		defer wg.Done()
		grpcErr = s.startGRPCServer()
	}()

	// Ждем запуска gRPC сервера
	select {
	case <-s.grpcStarted:
		log.Println("✅ gRPC server started successfully")
	case <-time.After(10 * time.Second):
		return fmt.Errorf("gRPC server failed to start within 10 seconds")
	}

	// Запускаем HTTP gateway
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpErr = s.startHTTPGateway()
	}()

	// Ждем завершения обоих серверов
	wg.Wait()

	if grpcErr != nil {
		return fmt.Errorf("gRPC server error: %w", grpcErr)
	}
	if httpErr != nil {
		return fmt.Errorf("HTTP gateway error: %w", httpErr)
	}

	return nil
}

// startGRPCServer запускает gRPC сервер
func (s *GRPCServer) startGRPCServer() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.grpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.grpcPort, err)
	}

	log.Printf("🚀 Starting gRPC server on port %d", s.grpcPort)

	// Сигнализируем о готовности gRPC сервера
	close(s.grpcStarted)

	return s.server.Serve(listener)
}

// startHTTPGateway запускает HTTP gateway
func (s *GRPCServer) startHTTPGateway() error {
	// Используем контекст который НЕ отменяется
	ctx := context.Background()

	// Создаем mux для gRPC gateway
	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(s.customErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
	)

	// Подключаемся к gRPC серверу с правильными опциями
	grpcEndpoint := fmt.Sprintf("localhost:%d", s.grpcPort)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// Убираем WithBlock() - он может вызывать проблемы
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	log.Printf("🔗 Connecting to gRPC server at %s", grpcEndpoint)

	// Используем RegisterCurrencyServiceHandler вместо RegisterCurrencyServiceHandlerFromEndpoint
	conn, err := grpc.DialContext(ctx, grpcEndpoint, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	// Регистрируем handler с существующим соединением
	err = currencyv1.RegisterCurrencyServiceHandler(ctx, mux, conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	log.Printf("✅ Gateway registered successfully")

	// Добавляем middleware
	handler := s.corsMiddleware(s.loggingMiddleware(mux))

	log.Printf("🌐 Starting HTTP gateway on port %d", s.httpPort)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.httpPort), handler)
}

// corsMiddleware добавляет CORS заголовки
func (s *GRPCServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware логирует HTTP запросы
func (s *GRPCServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("📥 %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("📤 %s %s - %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// customErrorHandler обрабатывает ошибки Gateway
func (s *GRPCServer) customErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("❌ Gateway error: %v", err)
	runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
}

// Stop останавливает сервер
func (s *GRPCServer) Stop() {
	log.Println("🛑 Stopping servers...")
	if s.healthServer != nil {
		s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}
	s.server.GracefulStop()
	log.Println("✅ Servers stopped")
}
