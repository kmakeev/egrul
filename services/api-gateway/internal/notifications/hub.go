package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/egrul-system/services/api-gateway/internal/config"
	"github.com/egrul-system/services/api-gateway/internal/repository/postgresql"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Hub управляет SSE клиентами и распределяет уведомления из Kafka
type Hub struct {
	clients   map[string]*Client // email -> Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *NotificationEvent

	companyReader      *kafka.Reader
	entrepreneurReader *kafka.Reader
	subscriptionRepo   *postgresql.SubscriptionRepository

	config *config.NotificationHubConfig
	logger *zap.Logger
	mu     sync.RWMutex
}

// NewHub создает новый Notification Hub
func NewHub(
	db *sql.DB,
	pgSchema string,
	kafkaCfg config.KafkaConfig,
	hubCfg config.NotificationHubConfig,
	logger *zap.Logger,
) *Hub {
	subscriptionRepo := postgresql.NewSubscriptionRepository(db, pgSchema, logger)

	// Создать Kafka readers для двух топиков
	companyReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        kafkaCfg.Brokers,
		Topic:          kafkaCfg.CompanyTopic,
		GroupID:        kafkaCfg.ConsumerGroup,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset, // Читаем только новые события
		Logger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Debug(fmt.Sprintf(msg, args...))
		}),
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Error(fmt.Sprintf(msg, args...))
		}),
	})

	entrepreneurReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        kafkaCfg.Brokers,
		Topic:          kafkaCfg.EntrepreneurTopic,
		GroupID:        kafkaCfg.ConsumerGroup,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
		Logger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Debug(fmt.Sprintf(msg, args...))
		}),
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Error(fmt.Sprintf(msg, args...))
		}),
	})

	return &Hub{
		clients:            make(map[string]*Client),
		register:           make(chan *Client, 10),
		unregister:         make(chan *Client, 10),
		broadcast:          make(chan *NotificationEvent, hubCfg.BufferSize),
		companyReader:      companyReader,
		entrepreneurReader: entrepreneurReader,
		subscriptionRepo:   subscriptionRepo,
		config:             &hubCfg,
		logger:             logger,
	}
}

// Run запускает Hub в фоновом режиме
func (h *Hub) Run(ctx context.Context) {
	h.logger.Info("Starting Notification Hub",
		zap.Int("buffer_size", h.config.BufferSize),
		zap.Duration("heartbeat_interval", h.config.HeartbeatInterval),
		zap.Int("max_clients", h.config.MaxClients),
	)

	// Запустить Kafka consumers в отдельных горутинах
	go h.consumeKafka(ctx, h.companyReader, "company")
	go h.consumeKafka(ctx, h.entrepreneurReader, "entrepreneur")

	// Основной цикл обработки событий
	heartbeatTicker := time.NewTicker(h.config.HeartbeatInterval)
	defer heartbeatTicker.Stop()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case event := <-h.broadcast:
			h.broadcastEvent(event)

		case <-heartbeatTicker.C:
			h.sendHeartbeats()

		case <-ctx.Done():
			h.logger.Info("Stopping Notification Hub")
			h.shutdown()
			return
		}
	}
}

// RegisterClient добавляет нового SSE клиента
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient удаляет SSE клиента
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Проверить лимит клиентов
	if len(h.clients) >= h.config.MaxClients {
		h.logger.Warn("Max clients limit reached, rejecting client",
			zap.String("email", client.Email),
			zap.Int("current_clients", len(h.clients)),
		)
		client.Close()
		return
	}

	// Если клиент уже подключен (другая вкладка), закрываем старое соединение
	if existingClient, ok := h.clients[client.Email]; ok {
		h.logger.Info("Client already connected, closing old connection",
			zap.String("email", client.Email),
		)
		existingClient.Close()
	}

	h.clients[client.Email] = client
	h.logger.Info("Client registered",
		zap.String("email", client.Email),
		zap.Int("total_clients", len(h.clients)),
	)
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.Email]; ok {
		delete(h.clients, client.Email)
		client.Close()
		h.logger.Info("Client unregistered",
			zap.String("email", client.Email),
			zap.Int("total_clients", len(h.clients)),
		)
	}
}

func (h *Hub) broadcastEvent(event *NotificationEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.logger.Info("Broadcasting event",
		zap.String("event_id", event.ID),
		zap.String("entity_type", event.EntityType),
		zap.String("entity_id", event.EntityID),
		zap.String("change_type", event.ChangeType),
	)

	ctx := context.Background()

	// Получить все подписки для данной сущности
	var subscriptions []postgresql.EntitySubscription
	var err error

	if event.EntityType == "company" {
		subscriptions, err = h.subscriptionRepo.GetActiveSubscriptionsForEntity(ctx, "company", event.EntityID)
	} else if event.EntityType == "entrepreneur" {
		subscriptions, err = h.subscriptionRepo.GetActiveSubscriptionsForEntity(ctx, "entrepreneur", event.EntityID)
	}

	if err != nil {
		h.logger.Error("Failed to get subscriptions",
			zap.String("entity_type", event.EntityType),
			zap.String("entity_id", event.EntityID),
			zap.Error(err),
		)
		return
	}

	h.logger.Info("Found subscriptions for broadcast",
		zap.String("entity_type", event.EntityType),
		zap.String("entity_id", event.EntityID),
		zap.Int("count", len(subscriptions)),
		zap.Int("connected_clients", len(h.clients)),
	)

	// Отправить уведомление каждому подписчику
	sentCount := 0
	for _, sub := range subscriptions {
		// Проверить фильтры изменений
		if !h.shouldNotify(sub.ChangeFilters, event.ChangeType) {
			continue
		}

		// Найти клиента
		client, ok := h.clients[sub.UserEmail]
		if !ok {
			// Клиент не подключен через SSE, пропускаем
			continue
		}

		// Отправить событие клиенту
		if client.Send(event) {
			sentCount++
		} else {
			h.logger.Warn("Client buffer full, dropping notification",
				zap.String("email", client.Email),
				zap.String("event_id", event.ID),
			)
		}
	}

	if sentCount > 0 {
		h.logger.Info("Event broadcast to clients",
			zap.String("event_id", event.ID),
			zap.String("entity_type", event.EntityType),
			zap.String("entity_id", event.EntityID),
			zap.Int("sent_count", sentCount),
		)
	} else {
		h.logger.Warn("No clients received event",
			zap.String("event_id", event.ID),
			zap.String("entity_type", event.EntityType),
			zap.String("entity_id", event.EntityID),
			zap.Int("subscriptions_count", len(subscriptions)),
			zap.Int("connected_clients", len(h.clients)),
		)
	}
}

func (h *Hub) sendHeartbeats() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Отправить heartbeat всем подключенным клиентам
	// В реальной реализации можно отправить специальное событие :heartbeat
	// Но для SSE достаточно просто отправлять комментарий через handler
	// Heartbeat отправляется через SSE handler (ticker в ServeSSE)
}

func (h *Hub) consumeKafka(ctx context.Context, reader *kafka.Reader, entityType string) {
	h.logger.Info("Started consuming Kafka topic",
		zap.String("topic", reader.Config().Topic),
		zap.String("entity_type", entityType),
	)

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("Stopping Kafka consumer",
				zap.String("topic", reader.Config().Topic),
			)
			return
		default:
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				h.logger.Error("Failed to fetch Kafka message",
					zap.String("topic", reader.Config().Topic),
					zap.Error(err),
				)
				time.Sleep(time.Second)
				continue
			}

			// Десериализовать событие
			var changeEvent ChangeEvent
			if err := json.Unmarshal(msg.Value, &changeEvent); err != nil {
				h.logger.Error("Failed to unmarshal change event",
					zap.String("topic", reader.Config().Topic),
					zap.Error(err),
				)
				_ = reader.CommitMessages(ctx, msg)
				continue
			}

			// Преобразовать в NotificationEvent и отправить в broadcast
			notificationEvent := changeEvent.ToNotificationEvent()
			h.logger.Info("Received change event from Kafka",
				zap.String("change_id", changeEvent.ChangeID),
				zap.String("entity_type", changeEvent.EntityType),
				zap.String("entity_id", changeEvent.EntityID),
				zap.String("change_type", string(changeEvent.ChangeType)),
			)
			h.broadcast <- notificationEvent

			// Коммитить offset после успешной обработки
			if err := reader.CommitMessages(ctx, msg); err != nil {
				h.logger.Error("Failed to commit Kafka message",
					zap.String("topic", reader.Config().Topic),
					zap.Error(err),
				)
			}
		}
	}
}

func (h *Hub) shouldNotify(filters map[string]bool, changeType string) bool {
	// Если фильтры пусты, уведомлять о всех изменениях
	if len(filters) == 0 {
		return true
	}

	// Проверить конкретный фильтр
	notify, ok := filters[changeType]
	return ok && notify
}

func (h *Hub) shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Закрыть все клиентские соединения
	for _, client := range h.clients {
		client.Close()
	}
	h.clients = make(map[string]*Client)

	// Закрыть Kafka readers
	if err := h.companyReader.Close(); err != nil {
		h.logger.Error("Failed to close company reader", zap.Error(err))
	}
	if err := h.entrepreneurReader.Close(); err != nil {
		h.logger.Error("Failed to close entrepreneur reader", zap.Error(err))
	}

	h.logger.Info("Notification Hub stopped")
}

// GetStats возвращает статистику Hub
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"total_clients":    len(h.clients),
		"max_clients":      h.config.MaxClients,
		"buffer_size":      h.config.BufferSize,
		"broadcast_queue":  len(h.broadcast),
	}
}
