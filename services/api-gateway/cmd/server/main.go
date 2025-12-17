// Package main - точка входа API Gateway
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/egrul-system/services/api-gateway/internal/cache"
	"github.com/egrul-system/services/api-gateway/internal/config"
	"github.com/egrul-system/services/api-gateway/internal/graph"
	"github.com/egrul-system/services/api-gateway/internal/middleware"
	"github.com/egrul-system/services/api-gateway/internal/repository/clickhouse"
	"github.com/egrul-system/services/api-gateway/internal/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Version информация о версии (устанавливается при сборке)
var Version = "dev"

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Инициализация логгера
	logger, err := initLogger(cfg.Log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting EGRUL API Gateway",
		zap.String("version", Version),
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
	)

	// Подключение к ClickHouse
	chClient, err := clickhouse.NewClient(&cfg.ClickHouse, logger)
	if err != nil {
		logger.Fatal("failed to connect to clickhouse", zap.Error(err))
	}
	defer chClient.Close()

	// Инициализация репозиториев
	companyRepo := clickhouse.NewCompanyRepository(chClient, logger)
	entrepreneurRepo := clickhouse.NewEntrepreneurRepository(chClient, logger)
	founderRepo := clickhouse.NewFounderRepository(chClient, logger)
	licenseRepo := clickhouse.NewLicenseRepository(chClient, logger)
	branchRepo := clickhouse.NewBranchRepository(chClient, logger)
	historyRepo := clickhouse.NewHistoryRepository(chClient, logger)
	statsRepo := clickhouse.NewStatisticsRepository(chClient, logger)

	// Инициализация Redis кэша
	redisCache := cache.NewRedisCache(cfg.Redis, logger)
	defer redisCache.Close()

	// Инициализация сервисов
	companyService := service.NewCompanyService(companyRepo, founderRepo, licenseRepo, branchRepo, historyRepo, logger)
	entrepreneurService := service.NewEntrepreneurService(entrepreneurRepo, licenseRepo, historyRepo, logger)
	statsService := service.NewStatisticsService(statsRepo, logger)
	searchService := service.NewSearchService(companyService, entrepreneurService, logger)

	// Инициализация GraphQL резолвера
	resolver := graph.NewResolver(companyService, entrepreneurService, statsService, searchService, redisCache, logger)

	// Создание роутера
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(chimiddleware.Compress(5))
	r.Use(middleware.Timeout(30 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", healthHandler(chClient))
	r.Get("/ready", readyHandler(chClient))

	// GraphQL endpoint
	graphqlHandler := graph.NewManualHandler(resolver)
	r.Group(func(r chi.Router) {
		// DataLoader middleware для GraphQL эндпоинтов
		r.Use(graph.DataLoaderMiddleware(resolver))
		r.Handle("/graphql", graphqlHandler)
		r.Handle("/query", graphqlHandler) // Alias
	})

	// GraphQL Playground
	if cfg.GraphQL.PlaygroundEnabled {
		r.Get("/", playground.Handler("EGRUL GraphQL Playground", "/graphql"))
		r.Get("/playground", playground.Handler("EGRUL GraphQL Playground", "/graphql"))
		logger.Info("GraphQL Playground enabled at /playground")
	}

	// REST API compatibility endpoints
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/companies/{ogrn}", restCompanyHandler(companyService))
		r.Get("/entrepreneurs/{ogrnip}", restEntrepreneurHandler(entrepreneurService))
		r.Get("/search", restSearchHandler(searchService))
	})

	// Создание HTTP сервера
	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Запуск сервера в горутине
	go func() {
		logger.Info("HTTP server starting",
			zap.String("addr", srv.Addr),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped")
}

func initLogger(cfg config.LogConfig) (*zap.Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	var zapCfg zap.Config
	if cfg.Format == "json" {
		zapCfg = zap.NewProductionConfig()
	} else {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	return zapCfg.Build()
}

func healthHandler(chClient *clickhouse.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"api-gateway","version":"` + Version + `"}`))
	}
}

func readyHandler(chClient *clickhouse.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Проверяем подключение к ClickHouse
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := chClient.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(fmt.Sprintf(`{"status":"error","error":"clickhouse: %s"}`, err.Error())))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready","clickhouse":"connected"}`))
	}
}

// REST API handlers for backward compatibility

func restCompanyHandler(svc *service.CompanyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ogrn := chi.URLParam(r, "ogrn")
		company, err := svc.GetByOGRN(r.Context(), ogrn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if company == nil {
			http.Error(w, "company not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		// Simple JSON response - in production use encoding/json
		fmt.Fprintf(w, `{"ogrn":"%s","inn":"%s","fullName":"%s","status":"%s"}`,
			company.Ogrn, company.Inn, company.FullName, company.Status)
	}
}

func restEntrepreneurHandler(svc *service.EntrepreneurService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ogrnip := chi.URLParam(r, "ogrnip")
		entrepreneur, err := svc.GetByOGRNIP(r.Context(), ogrnip)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if entrepreneur == nil {
			http.Error(w, "entrepreneur not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"ogrnip":"%s","inn":"%s","lastName":"%s","firstName":"%s","status":"%s"}`,
			entrepreneur.Ogrnip, entrepreneur.Inn, entrepreneur.LastName, entrepreneur.FirstName, entrepreneur.Status)
	}
}

func restSearchHandler(svc *service.SearchService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
			return
		}

		result, err := svc.Search(r.Context(), query, 10)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"totalCompanies":%d,"totalEntrepreneurs":%d}`,
			result.TotalCompanies, result.TotalEntrepreneurs)
	}
}
