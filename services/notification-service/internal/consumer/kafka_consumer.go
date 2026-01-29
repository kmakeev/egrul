package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/egrul/notification-service/internal/model"
	"github.com/egrul/notification-service/internal/service"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaConsumer отвечает за чтение событий изменений из Kafka
type KafkaConsumer struct {
	companyReader      *kafka.Reader
	entrepreneurReader *kafka.Reader
	service            *service.NotificationService
	logger             *zap.Logger
}

// ConsumerConfig конфигурация Kafka consumer
type ConsumerConfig struct {
	Brokers                  []string
	CompanyTopic             string
	EntrepreneurTopic        string
	GroupID                  string
}

// NewKafkaConsumer создает новый экземпляр Kafka Consumer
func NewKafkaConsumer(cfg ConsumerConfig, svc *service.NotificationService, logger *zap.Logger) *KafkaConsumer {
	companyReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.CompanyTopic,
		GroupID:        cfg.GroupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
		Logger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Debug(fmt.Sprintf(msg, args...))
		}),
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Error(fmt.Sprintf(msg, args...))
		}),
	})

	entrepreneurReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.EntrepreneurTopic,
		GroupID:        cfg.GroupID,
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

	return &KafkaConsumer{
		companyReader:      companyReader,
		entrepreneurReader: entrepreneurReader,
		service:            svc,
		logger:             logger,
	}
}

// Start запускает обработку событий из Kafka
func (c *KafkaConsumer) Start(ctx context.Context) error {
	c.logger.Info("Starting Kafka consumer",
		zap.String("company_topic", c.companyReader.Config().Topic),
		zap.String("entrepreneur_topic", c.entrepreneurReader.Config().Topic),
		zap.String("group_id", c.companyReader.Config().GroupID),
	)

	// Запускаем обработку обоих топиков в отдельных горутинах
	errChan := make(chan error, 2)

	go func() {
		errChan <- c.consumeTopic(ctx, c.companyReader, "company")
	}()

	go func() {
		errChan <- c.consumeTopic(ctx, c.entrepreneurReader, "entrepreneur")
	}()

	// Ждем первую ошибку
	return <-errChan
}

// consumeTopic обрабатывает сообщения из конкретного топика
func (c *KafkaConsumer) consumeTopic(ctx context.Context, reader *kafka.Reader, entityType string) error {
	c.logger.Info("Started consuming topic",
		zap.String("topic", reader.Config().Topic),
		zap.String("entity_type", entityType),
	)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping consumer",
				zap.String("topic", reader.Config().Topic),
			)
			return ctx.Err()
		default:
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				c.logger.Error("failed to fetch message",
					zap.String("topic", reader.Config().Topic),
					zap.Error(err),
				)
				time.Sleep(time.Second)
				continue
			}

			c.logger.Debug("received message",
				zap.String("topic", reader.Config().Topic),
				zap.String("key", string(msg.Key)),
				zap.Int("partition", msg.Partition),
				zap.Int64("offset", msg.Offset),
			)

			// Обрабатываем сообщение
			if err := c.processMessage(ctx, msg); err != nil {
				c.logger.Error("failed to process message",
					zap.String("topic", reader.Config().Topic),
					zap.String("key", string(msg.Key)),
					zap.Error(err),
				)
				// Не возвращаем ошибку, продолжаем обработку следующих сообщений
			}

			// Коммитим offset только после успешной обработки
			if err := reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("failed to commit message",
					zap.String("topic", reader.Config().Topic),
					zap.Error(err),
				)
			}
		}
	}
}

// processMessage обрабатывает одно сообщение из Kafka
func (c *KafkaConsumer) processMessage(ctx context.Context, msg kafka.Message) error {
	// Десериализуем событие изменения
	var changeEvent model.ChangeEvent
	if err := json.Unmarshal(msg.Value, &changeEvent); err != nil {
		return fmt.Errorf("failed to unmarshal change event: %w", err)
	}

	c.logger.Info("processing change event",
		zap.String("change_id", changeEvent.ChangeID),
		zap.String("entity_type", changeEvent.EntityType),
		zap.String("entity_id", changeEvent.EntityID),
		zap.String("change_type", changeEvent.ChangeType),
		zap.Bool("is_significant", changeEvent.IsSignificant),
	)

	// Отправляем через NotificationService
	if err := c.service.ProcessChangeEvent(ctx, &changeEvent); err != nil {
		return fmt.Errorf("failed to process change event: %w", err)
	}

	c.logger.Info("change event processed successfully",
		zap.String("change_id", changeEvent.ChangeID),
		zap.String("entity_id", changeEvent.EntityID),
	)

	return nil
}

// Close закрывает readers и освобождает ресурсы
func (c *KafkaConsumer) Close() error {
	var err1, err2 error

	if err1 = c.companyReader.Close(); err1 != nil {
		c.logger.Error("failed to close company reader", zap.Error(err1))
	}

	if err2 = c.entrepreneurReader.Close(); err2 != nil {
		c.logger.Error("failed to close entrepreneur reader", zap.Error(err2))
	}

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	c.logger.Info("Kafka consumer closed successfully")
	return nil
}

// Stats возвращает статистику consumer
func (c *KafkaConsumer) Stats() map[string]interface{} {
	companyStats := c.companyReader.Stats()
	entrepreneurStats := c.entrepreneurReader.Stats()

	return map[string]interface{}{
		"company_consumer": map[string]interface{}{
			"topic":    c.companyReader.Config().Topic,
			"messages": companyStats.Messages,
			"bytes":    companyStats.Bytes,
			"lag":      companyStats.Lag,
		},
		"entrepreneur_consumer": map[string]interface{}{
			"topic":    c.entrepreneurReader.Config().Topic,
			"messages": entrepreneurStats.Messages,
			"bytes":    entrepreneurStats.Bytes,
			"lag":      entrepreneurStats.Lag,
		},
	}
}
