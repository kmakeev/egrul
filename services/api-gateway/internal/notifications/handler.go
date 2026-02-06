package notifications

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/egrul-system/services/api-gateway/internal/auth"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// ServeSSE обрабатывает SSE подключения клиентов
func (h *Hub) ServeSSE(w http.ResponseWriter, r *http.Request) {
	email := auth.GetEmailFromContext(r.Context())
	userID := auth.GetUserIDFromContext(r.Context())

	if email == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.logger.Info("New SSE connection",
		zap.String("email", email),
		zap.String("user_id", userID),
		zap.String("remote_addr", r.RemoteAddr),
	)

	// Установить SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no") // Для nginx

	// Создать клиента и зарегистрировать в Hub
	client := NewClient(email, userID, h.config.BufferSize)
	h.RegisterClient(client)

	// Уведомить об успешном подключении
	h.sendSSEEvent(w, &NotificationEvent{
		Type:      "connected",
		Timestamp: time.Now(),
	})

	// Отправить initial batch (последние непрочитанные уведомления)
	// Пропускаем для простоты первой версии - клиент может загрузить через REST API

	// Основной цикл отправки событий
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		h.UnregisterClient(client)
		return
	}

	ticker := time.NewTicker(h.config.HeartbeatInterval)
	defer ticker.Stop()
	defer h.UnregisterClient(client)

	for {
		select {
		case msg, ok := <-client.Messages:
			if !ok {
				// Канал закрыт, клиент отключен
				h.logger.Info("Client channel closed",
					zap.String("email", email),
				)
				return
			}
			// Отправить уведомление
			if err := h.sendSSEEvent(w, msg); err != nil {
				h.logger.Error("Failed to send SSE event",
					zap.String("email", email),
					zap.Error(err),
				)
				return
			}
			flusher.Flush()

		case <-ticker.C:
			// Отправить heartbeat (пустой комментарий)
			if err := h.sendHeartbeat(w); err != nil {
				h.logger.Error("Failed to send heartbeat",
					zap.String("email", email),
					zap.Error(err),
				)
				return
			}
			flusher.Flush()
			client.UpdateLastSeen()

		case <-client.Done:
			h.logger.Info("Client disconnected (Done channel)",
				zap.String("email", email),
			)
			return

		case <-r.Context().Done():
			h.logger.Info("Client disconnected (context cancelled)",
				zap.String("email", email),
			)
			return
		}
	}
}

// sendSSEEvent отправляет событие через SSE
func (h *Hub) sendSSEEvent(w http.ResponseWriter, event *NotificationEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	// SSE формат: id, event, data
	if event.ID != "" {
		if _, err := fmt.Fprintf(w, "id: %s\n", event.ID); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "event: %s\n", event.Type); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}

	return nil
}

// sendHeartbeat отправляет heartbeat (SSE comment)
func (h *Hub) sendHeartbeat(w http.ResponseWriter) error {
	_, err := fmt.Fprintf(w, ": heartbeat\n\n")
	return err
}

// ServeHistory обрабатывает запросы истории уведомлений
func (h *Hub) ServeHistory(w http.ResponseWriter, r *http.Request) {
	email := auth.GetEmailFromContext(r.Context())
	if email == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Параметры пагинации
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	offset := 0

	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}
	if offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			offset = val
		}
	}

	// Получить историю уведомлений из БД
	ctx := r.Context()
	notifications, err := h.subscriptionRepo.GetNotificationHistoryByEmail(ctx, email, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get notification history",
			zap.String("email", email),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notifications); err != nil {
		h.logger.Error("Failed to encode response",
			zap.Error(err),
		)
	}
}

// MarkAsRead отмечает уведомление как прочитанное
func (h *Hub) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	email := auth.GetEmailFromContext(r.Context())
	if email == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	notificationID := chi.URLParam(r, "id")
	if notificationID == "" {
		http.Error(w, "Missing notification ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.subscriptionRepo.MarkNotificationAsRead(ctx, notificationID, email); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Notification not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to mark notification as read",
			zap.String("email", email),
			zap.String("notification_id", notificationID),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// MarkAllAsRead отмечает все уведомления как прочитанные
func (h *Hub) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	email := auth.GetEmailFromContext(r.Context())
	if email == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	count, err := h.subscriptionRepo.MarkAllNotificationsAsRead(ctx, email)
	if err != nil {
		h.logger.Error("Failed to mark all notifications as read",
			zap.String("email", email),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"marked_count": count,
	})
}

// StatsHandler возвращает статистику Hub (для debug)
func (h *Hub) StatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := h.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
