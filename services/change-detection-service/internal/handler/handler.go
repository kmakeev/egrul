package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/egrul/change-detection-service/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Handler представляет HTTP обработчики
type Handler struct {
	service *service.DetectionService
	logger  *zap.Logger
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service *service.DetectionService, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// DetectRequest представляет запрос на детектирование изменений
type DetectRequest struct {
	EntityType string   `json:"entity_type"` // "company" или "entrepreneur"
	EntityIDs  []string `json:"entity_ids"`  // Список ОГРН или ОГРНИП
}

// DetectResponse представляет ответ на запрос детектирования
type DetectResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Count    int    `json:"count"`
	Duration string `json:"duration"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// HandleDetect обрабатывает POST /detect запрос
func (h *Handler) HandleDetect(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var req DetectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Валидация
	if req.EntityType != "company" && req.EntityType != "entrepreneur" {
		h.respondError(w, http.StatusBadRequest, "entity_type must be 'company' or 'entrepreneur'")
		return
	}

	if len(req.EntityIDs) == 0 {
		h.respondError(w, http.StatusBadRequest, "entity_ids cannot be empty")
		return
	}

	if len(req.EntityIDs) > 10000 {
		h.respondError(w, http.StatusBadRequest, "entity_ids cannot exceed 10000")
		return
	}

	h.logger.Info("received detect request",
		zap.String("entity_type", req.EntityType),
		zap.Int("count", len(req.EntityIDs)),
	)

	// Детектирование изменений
	var err error
	if req.EntityType == "company" {
		err = h.service.DetectCompanyChanges(r.Context(), req.EntityIDs)
	} else {
		err = h.service.DetectEntrepreneurChanges(r.Context(), req.EntityIDs)
	}

	if err != nil {
		h.logger.Error("failed to detect changes",
			zap.String("entity_type", req.EntityType),
			zap.Error(err),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to detect changes")
		return
	}

	duration := time.Since(startTime)
	h.respondJSON(w, http.StatusOK, DetectResponse{
		Success:  true,
		Message:  "Change detection completed successfully",
		Count:    len(req.EntityIDs),
		Duration: duration.String(),
	})
}

// HandleGetCompanyChanges обрабатывает GET /company/{ogrn}/changes запрос
func (h *Handler) HandleGetCompanyChanges(w http.ResponseWriter, r *http.Request) {
	ogrn := chi.URLParam(r, "ogrn")
	if ogrn == "" {
		h.respondError(w, http.StatusBadRequest, "ogrn is required")
		return
	}

	// Парсим limit из query параметров
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // по умолчанию
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	changes, err := h.service.GetCompanyChanges(r.Context(), ogrn, limit)
	if err != nil {
		h.logger.Error("failed to get company changes",
			zap.String("ogrn", ogrn),
			zap.Error(err),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to get company changes")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    changes,
		"count":   len(changes),
	})
}

// HandleGetEntrepreneurChanges обрабатывает GET /entrepreneur/{ogrnip}/changes запрос
func (h *Handler) HandleGetEntrepreneurChanges(w http.ResponseWriter, r *http.Request) {
	ogrnip := chi.URLParam(r, "ogrnip")
	if ogrnip == "" {
		h.respondError(w, http.StatusBadRequest, "ogrnip is required")
		return
	}

	// Парсим limit из query параметров
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // по умолчанию
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	changes, err := h.service.GetEntrepreneurChanges(r.Context(), ogrnip, limit)
	if err != nil {
		h.logger.Error("failed to get entrepreneur changes",
			zap.String("ogrnip", ogrnip),
			zap.Error(err),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to get entrepreneur changes")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    changes,
		"count":   len(changes),
	})
}

// HandleGetRecentChanges обрабатывает GET /changes/recent запрос
func (h *Handler) HandleGetRecentChanges(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	if entityType == "" {
		entityType = "company"
	}

	if entityType != "company" && entityType != "entrepreneur" {
		h.respondError(w, http.StatusBadRequest, "entity_type must be 'company' or 'entrepreneur'")
		return
	}

	// Парсим since из query параметров (timestamp)
	sinceStr := r.URL.Query().Get("since")
	since := time.Now().Add(-24 * time.Hour).Unix() // по умолчанию последние 24 часа
	if sinceStr != "" {
		parsedSince, err := strconv.ParseInt(sinceStr, 10, 64)
		if err == nil {
			since = parsedSince
		}
	}

	changes, err := h.service.GetRecentChanges(r.Context(), entityType, since)
	if err != nil {
		h.logger.Error("failed to get recent changes",
			zap.String("entity_type", entityType),
			zap.Error(err),
		)
		h.respondError(w, http.StatusInternalServerError, "failed to get recent changes")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"data":        changes,
		"count":       len(changes),
		"entity_type": entityType,
		"since":       since,
	})
}

// HandleHealth обрабатывает GET /health запрос
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "change-detection",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// HandleStats обрабатывает GET /stats запрос
func (h *Handler) HandleStats(w http.ResponseWriter, r *http.Request) {
	stats := h.service.GetStats()
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

// respondJSON отправляет JSON ответ
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", zap.Error(err))
	}
}

// respondError отправляет JSON ответ с ошибкой
func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, ErrorResponse{
		Success: false,
		Error:   message,
	})
}
