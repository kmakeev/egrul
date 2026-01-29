package service

import (
	"context"
	"fmt"
	"time"

	"github.com/egrul/notification-service/internal/channels"
	"github.com/egrul/notification-service/internal/model"
	"github.com/egrul/notification-service/internal/repository"
	"go.uber.org/zap"
)

// NotificationService отвечает за обработку событий изменений и отправку уведомлений
type NotificationService struct {
	subscriptionRepo      repository.SubscriptionRepository
	notificationLogRepo   repository.NotificationLogRepository
	channels              map[string]channels.NotificationChannel
	logger                *zap.Logger
}

// NewNotificationService создает новый экземпляр NotificationService
func NewNotificationService(
	subscriptionRepo repository.SubscriptionRepository,
	notificationLogRepo repository.NotificationLogRepository,
	emailChannel channels.NotificationChannel,
	logger *zap.Logger,
) *NotificationService {
	// Инициализируем map каналов
	channelsMap := make(map[string]channels.NotificationChannel)
	channelsMap["email"] = emailChannel

	return &NotificationService{
		subscriptionRepo:    subscriptionRepo,
		notificationLogRepo: notificationLogRepo,
		channels:            channelsMap,
		logger:              logger,
	}
}

// ProcessChangeEvent обрабатывает событие изменения и отправляет уведомления
func (s *NotificationService) ProcessChangeEvent(ctx context.Context, event *model.ChangeEvent) error {
	s.logger.Info("processing change event",
		zap.String("change_id", event.ChangeID),
		zap.String("entity_type", event.EntityType),
		zap.String("entity_id", event.EntityID),
		zap.String("change_type", event.ChangeType),
	)

	// Получаем все подписки на эту сущность
	subscriptions, err := s.subscriptionRepo.GetByEntity(ctx, event.EntityType, event.EntityID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	if len(subscriptions) == 0 {
		s.logger.Debug("no subscriptions found for entity",
			zap.String("entity_type", event.EntityType),
			zap.String("entity_id", event.EntityID),
		)
		return nil
	}

	s.logger.Info("found subscriptions",
		zap.Int("count", len(subscriptions)),
		zap.String("entity_id", event.EntityID),
	)

	// Отправляем уведомления для каждой подписки
	successCount := 0
	errorCount := 0

	for _, subscription := range subscriptions {
		if err := s.processSubscription(ctx, subscription, event); err != nil {
			s.logger.Error("failed to process subscription",
				zap.String("subscription_id", subscription.ID),
				zap.String("user_email", subscription.UserEmail),
				zap.Error(err),
			)
			errorCount++
		} else {
			successCount++
		}
	}

	s.logger.Info("change event processing completed",
		zap.String("change_id", event.ChangeID),
		zap.Int("success", successCount),
		zap.Int("errors", errorCount),
	)

	return nil
}

// processSubscription обрабатывает одну подписку
func (s *NotificationService) processSubscription(
	ctx context.Context,
	subscription *model.EntitySubscription,
	event *model.ChangeEvent,
) error {
	// Проверяем, нужно ли уведомлять по этому типу изменения
	if !subscription.ShouldNotify(event.ChangeType) {
		s.logger.Debug("change type filtered out",
			zap.String("subscription_id", subscription.ID),
			zap.String("change_type", event.ChangeType),
		)
		return nil
	}

	// Проверяем дубликаты (idempotency)
	isDuplicate, err := s.notificationLogRepo.CheckDuplicate(ctx, subscription.ID, event.ChangeID)
	if err != nil {
		return fmt.Errorf("failed to check duplicate: %w", err)
	}

	if isDuplicate {
		s.logger.Debug("notification already sent (duplicate)",
			zap.String("subscription_id", subscription.ID),
			zap.String("change_event_id", event.ChangeID),
		)
		return nil
	}

	// Отправляем уведомления по всем включенным каналам
	if subscription.HasEmailChannel() {
		if err := s.sendNotification(ctx, subscription, event, "email"); err != nil {
			return fmt.Errorf("failed to send email notification: %w", err)
		}
	}

	// Обновляем время последнего уведомления
	if err := s.subscriptionRepo.UpdateLastNotified(ctx, subscription.ID); err != nil {
		s.logger.Warn("failed to update last_notified_at",
			zap.String("subscription_id", subscription.ID),
			zap.Error(err),
		)
		// Не возвращаем ошибку, так как уведомление уже отправлено
	}

	return nil
}

// sendNotification отправляет уведомление через конкретный канал
func (s *NotificationService) sendNotification(
	ctx context.Context,
	subscription *model.EntitySubscription,
	event *model.ChangeEvent,
	channelName string,
) error {
	channel, ok := s.channels[channelName]
	if !ok {
		return fmt.Errorf("channel not found: %s", channelName)
	}

	// Создаем объект уведомления
	notification := &model.Notification{
		SubscriptionID: subscription.ID,
		ChangeEvent:    event,
		UserEmail:      subscription.UserEmail,
		Channel:        channelName,
		Status:         model.NotificationStatusPending,
		CreatedAt:      time.Now(),
	}

	s.logger.Info("sending notification",
		zap.String("subscription_id", subscription.ID),
		zap.String("user_email", subscription.UserEmail),
		zap.String("channel", channelName),
		zap.String("change_type", event.ChangeType),
	)

	// Отправляем через канал
	err := channel.Send(ctx, notification)

	// Обновляем статус
	if err != nil {
		notification.Status = model.NotificationStatusFailed
		notification.ErrorMessage = err.Error()
		s.logger.Error("notification failed",
			zap.String("subscription_id", subscription.ID),
			zap.String("channel", channelName),
			zap.Error(err),
		)
	} else {
		notification.Status = model.NotificationStatusSent
		now := time.Now()
		notification.SentAt = &now
		s.logger.Info("notification sent successfully",
			zap.String("subscription_id", subscription.ID),
			zap.String("user_email", subscription.UserEmail),
			zap.String("channel", channelName),
		)
	}

	// Сохраняем в лог
	if logErr := s.notificationLogRepo.Save(ctx, notification); logErr != nil {
		s.logger.Error("failed to save notification log",
			zap.String("subscription_id", subscription.ID),
			zap.Error(logErr),
		)
		// Не возвращаем ошибку, так как само уведомление было обработано
	}

	return err
}

// GetStats возвращает статистику сервиса
func (s *NotificationService) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"channels": s.getChannelStats(),
	}
}

// getChannelStats возвращает статистику каналов
func (s *NotificationService) getChannelStats() map[string]string {
	stats := make(map[string]string)
	for name := range s.channels {
		stats[name] = "active"
	}
	return stats
}
