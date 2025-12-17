package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Настройка логирования
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Настройка Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Роуты
	setupRoutes(router)

	// Сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Info().Str("port", port).Msg("Запуск Search Service")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Ошибка запуска сервера")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Остановка сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Ошибка остановки сервера")
	}

	log.Info().Msg("Сервер остановлен")
}

func setupRoutes(r *gin.Engine) {
	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "search-service",
		})
	})

	// Search API
	r.POST("/search", handleSearch)
	r.POST("/index", handleIndex)
	r.DELETE("/index/:id", handleDeleteFromIndex)
}

// SearchRequest - запрос на поиск
type SearchRequest struct {
	Query    string   `json:"query" binding:"required"`
	Type     string   `json:"type,omitempty"` // legal_entity, entrepreneur, all
	Filters  []Filter `json:"filters,omitempty"`
	Page     int      `json:"page,omitempty"`
	PageSize int      `json:"page_size,omitempty"`
}

// Filter - фильтр поиска
type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, in, contains
	Value    interface{} `json:"value"`
}

// IndexRequest - запрос на индексацию
type IndexRequest struct {
	ID   string      `json:"id" binding:"required"`
	Type string      `json:"type" binding:"required"`
	Data interface{} `json:"data" binding:"required"`
}

func handleSearch(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Реализовать поиск через Elasticsearch
	log.Info().Str("query", req.Query).Msg("Поиск")

	c.JSON(http.StatusOK, gin.H{
		"query":   req.Query,
		"results": []interface{}{},
		"total":   0,
		"page":    req.Page,
	})
}

func handleIndex(c *gin.Context) {
	var req IndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Реализовать индексацию в Elasticsearch
	log.Info().Str("id", req.ID).Str("type", req.Type).Msg("Индексация документа")

	c.JSON(http.StatusOK, gin.H{
		"status": "indexed",
		"id":     req.ID,
	})
}

func handleDeleteFromIndex(c *gin.Context) {
	id := c.Param("id")

	// TODO: Реализовать удаление из индекса
	log.Info().Str("id", id).Msg("Удаление из индекса")

	c.JSON(http.StatusOK, gin.H{
		"status": "deleted",
		"id":     id,
	})
}

