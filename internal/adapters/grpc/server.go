package grpc

import (
	"context"
	"encoding/json"
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

type GRPCServer struct {
	server       *grpc.Server
	handler      *CurrencyHandler
	grpcPort     int
	httpPort     int
	grpcStarted  chan struct{}
	healthServer *health.Server
}

func NewGRPCServer(service ports.CurrencyService, grpcPort, httpPort int) *GRPCServer {
	server := grpc.NewServer()
	handler := NewCurrencyHandler(service)
	healthServer := health.NewServer()

	currencyv1.RegisterCurrencyServiceServer(server, handler)

	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

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

func (s *GRPCServer) Start() error {
	var wg sync.WaitGroup
	var grpcErr, httpErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		grpcErr = s.startGRPCServer()
	}()

	select {
	case <-s.grpcStarted:
		log.Println("✅ gRPC server started successfully")
	case <-time.After(10 * time.Second):
		return fmt.Errorf("gRPC server failed to start within 10 seconds")
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		httpErr = s.startHTTPGateway()
	}()

	wg.Wait()

	if grpcErr != nil {
		return fmt.Errorf("gRPC server error: %w", grpcErr)
	}
	if httpErr != nil {
		return fmt.Errorf("HTTP gateway error: %w", httpErr)
	}

	return nil
}

func (s *GRPCServer) startGRPCServer() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.grpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.grpcPort, err)
	}

	log.Printf("🚀 Starting gRPC server on port %d", s.grpcPort)

	close(s.grpcStarted)

	return s.server.Serve(listener)
}

func (s *GRPCServer) startHTTPGateway() error {
	ctx := context.Background()

	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(s.customErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
		runtime.WithIncomingHeaderMatcher(s.customHeaderMatcher),
	)

	grpcEndpoint := fmt.Sprintf("localhost:%d", s.grpcPort)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	log.Printf("🔗 Connecting to gRPC server at %s", grpcEndpoint)

	conn, err := grpc.NewClient(grpcEndpoint, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	err = currencyv1.RegisterCurrencyServiceHandler(ctx, mux, conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	log.Printf("✅ Gateway registered successfully")

	handler := s.corsMiddleware(TimeValidationMiddleware(s.loggingMiddleware(mux)))

	log.Printf("🌐 Starting HTTP gateway on port %d", s.httpPort)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.httpPort), handler)
}

func (s *GRPCServer) customHeaderMatcher(key string) (string, bool) {
	switch key {
	case "starttime", "endtime", "timestamp":
		return fmt.Sprintf("grpc-gateway-%s", key), true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

func (s *GRPCServer) customErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("❌ Gateway error: %v", err)
	if timeErr, ok := err.(*InvalidTimeFormatError); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		errorResponse := map[string]interface{}{
			"error": map[string]interface{}{
				"code":    3,
				"message": timeErr.Error(),
				"details": map[string]interface{}{
					"field":             "timestamp",
					"provided_value":    timeErr.Value,
					"supported_formats": timeErr.SupportedFormats,
					"examples": []string{
						"2025-08-07T20:15:30Z",
						"2025-08-07T20:15:30.123Z",
						"2025-08-07T20:15:30+03:00",
						"2025-08-07",
					},
				},
			},
		}

		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
}

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

func (s *GRPCServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("📥 %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("📤 %s %s - %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *GRPCServer) Stop() {
	log.Println("🛑 Stopping servers...")
	if s.healthServer != nil {
		s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}
	s.server.GracefulStop()
	log.Println("✅ Servers stopped")
}
