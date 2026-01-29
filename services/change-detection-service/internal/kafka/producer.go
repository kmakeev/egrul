package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/egrul/change-detection-service/internal/model"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Producer отвечает за отправку событий изменений в Kafka
type Producer struct {
	companyWriter      *kafka.Writer
	entrepreneurWriter *kafka.Writer
	logger             *zap.Logger
}

// ProducerConfig конфигурация Kafka producer
type ProducerConfig struct {
	Brokers               []string
	CompanyTopic          string
	EntrepreneurTopic     string
	RequiredAcks          int
	BatchSize             int
	BatchTimeout          time.Duration
}

// NewProducer создает новый экземпляр Kafka Producer
func NewProducer(cfg ProducerConfig, logger *zap.Logger) *Producer {
	companyWriter := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.CompanyTopic,
		Balancer:     &kafka.Hash{}, // Используем Hash балансировку по ключу (OGRN)
		RequiredAcks: kafka.RequiredAcks(cfg.RequiredAcks),
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.BatchTimeout,
		MaxAttempts:  3,
		Logger:       kafka.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Debug(fmt.Sprintf(msg, args...))
		}),
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Error(fmt.Sprintf(msg, args...))
		}),
	}

	entrepreneurWriter := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.EntrepreneurTopic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequiredAcks(cfg.RequiredAcks),
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.BatchTimeout,
		MaxAttempts:  3,
		Logger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Debug(fmt.Sprintf(msg, args...))
		}),
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Error(fmt.Sprintf(msg, args...))
		}),
	}

	return &Producer{
		companyWriter:      companyWriter,
		entrepreneurWriter: entrepreneurWriter,
		logger:             logger,
	}
}

// SendCompanyChange отправляет событие изменения компании в Kafka
func (p *Producer) SendCompanyChange(ctx context.Context, change *model.ChangeEvent) error {
	if !change.IsCompany() {
		return fmt.Errorf("change event is not for a company")
	}

	return p.sendChange(ctx, p.companyWriter, change)
}

// SendCompanyChanges отправляет несколько событий изменений компаний (батчинг)
func (p *Producer) SendCompanyChanges(ctx context.Context, changes []*model.ChangeEvent) error {
	if len(changes) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(changes))
	for _, change := range changes {
		if !change.IsCompany() {
			p.logger.Warn("skipping non-company change event", zap.String("entity_type", change.EntityType))
			continue
		}

		msg, err := p.buildMessage(change)
		if err != nil {
			p.logger.Error("failed to build message", zap.Error(err))
			continue
		}

		messages = append(messages, msg)
	}

	if len(messages) == 0 {
		return nil
	}

	err := p.companyWriter.WriteMessages(ctx, messages...)
	if err != nil {
		p.logger.Error("failed to send company changes batch",
			zap.Int("count", len(messages)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send company changes batch: %w", err)
	}

	p.logger.Info("sent company changes batch to Kafka",
		zap.Int("count", len(messages)),
		zap.String("topic", p.companyWriter.Topic),
	)

	return nil
}

// SendEntrepreneurChange отправляет событие изменения ИП в Kafka
func (p *Producer) SendEntrepreneurChange(ctx context.Context, change *model.ChangeEvent) error {
	if !change.IsEntrepreneur() {
		return fmt.Errorf("change event is not for an entrepreneur")
	}

	return p.sendChange(ctx, p.entrepreneurWriter, change)
}

// SendEntrepreneurChanges отправляет несколько событий изменений ИП (батчинг)
func (p *Producer) SendEntrepreneurChanges(ctx context.Context, changes []*model.ChangeEvent) error {
	if len(changes) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(changes))
	for _, change := range changes {
		if !change.IsEntrepreneur() {
			p.logger.Warn("skipping non-entrepreneur change event", zap.String("entity_type", change.EntityType))
			continue
		}

		msg, err := p.buildMessage(change)
		if err != nil {
			p.logger.Error("failed to build message", zap.Error(err))
			continue
		}

		messages = append(messages, msg)
	}

	if len(messages) == 0 {
		return nil
	}

	err := p.entrepreneurWriter.WriteMessages(ctx, messages...)
	if err != nil {
		p.logger.Error("failed to send entrepreneur changes batch",
			zap.Int("count", len(messages)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send entrepreneur changes batch: %w", err)
	}

	p.logger.Info("sent entrepreneur changes batch to Kafka",
		zap.Int("count", len(messages)),
		zap.String("topic", p.entrepreneurWriter.Topic),
	)

	return nil
}

// sendChange отправляет одно событие изменения
func (p *Producer) sendChange(ctx context.Context, writer *kafka.Writer, change *model.ChangeEvent) error {
	msg, err := p.buildMessage(change)
	if err != nil {
		return fmt.Errorf("failed to build message: %w", err)
	}

	err = writer.WriteMessages(ctx, msg)
	if err != nil {
		p.logger.Error("failed to send change event",
			zap.String("entity_id", change.EntityID),
			zap.String("change_type", string(change.ChangeType)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send change event: %w", err)
	}

	p.logger.Debug("sent change event to Kafka",
		zap.String("entity_id", change.EntityID),
		zap.String("change_type", string(change.ChangeType)),
		zap.String("change_id", change.ChangeID),
		zap.String("topic", writer.Topic),
	)

	return nil
}

// buildMessage создает Kafka сообщение из события изменения
func (p *Producer) buildMessage(change *model.ChangeEvent) (kafka.Message, error) {
	// Сериализуем событие в JSON
	value, err := json.Marshal(change)
	if err != nil {
		return kafka.Message{}, fmt.Errorf("failed to marshal change event: %w", err)
	}

	// Ключ сообщения - entity_id (OGRN или OGRNIP)
	// Это обеспечивает упорядоченность событий для одной сущности
	key := []byte(change.EntityID)

	// Заголовки сообщения (метаданные)
	headers := []kafka.Header{
		{Key: "entity_type", Value: []byte(change.EntityType)},
		{Key: "change_type", Value: []byte(change.ChangeType)},
		{Key: "change_id", Value: []byte(change.ChangeID)},
		{Key: "is_significant", Value: []byte(fmt.Sprintf("%t", change.IsSignificant))},
		{Key: "region_code", Value: []byte(change.RegionCode)},
	}

	msg := kafka.Message{
		Key:     key,
		Value:   value,
		Headers: headers,
		Time:    change.DetectedAt,
	}

	return msg, nil
}

// Close закрывает writers и освобождает ресурсы
func (p *Producer) Close() error {
	var err1, err2 error

	if err1 = p.companyWriter.Close(); err1 != nil {
		p.logger.Error("failed to close company writer", zap.Error(err1))
	}

	if err2 = p.entrepreneurWriter.Close(); err2 != nil {
		p.logger.Error("failed to close entrepreneur writer", zap.Error(err2))
	}

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	p.logger.Info("Kafka producer closed successfully")
	return nil
}

// Stats возвращает статистику producer
func (p *Producer) Stats() map[string]interface{} {
	companyStats := p.companyWriter.Stats()
	entrepreneurStats := p.entrepreneurWriter.Stats()

	return map[string]interface{}{
		"company_producer": map[string]interface{}{
			"topic":          p.companyWriter.Topic,
			"writes":         companyStats.Writes,
			"messages":       companyStats.Messages,
			"bytes":          companyStats.Bytes,
			"errors":         companyStats.Errors,
			"batch_time_avg": companyStats.BatchTime.Avg,
			"batch_size_avg": companyStats.BatchSize.Avg,
		},
		"entrepreneur_producer": map[string]interface{}{
			"topic":          p.entrepreneurWriter.Topic,
			"writes":         entrepreneurStats.Writes,
			"messages":       entrepreneurStats.Messages,
			"bytes":          entrepreneurStats.Bytes,
			"errors":         entrepreneurStats.Errors,
			"batch_time_avg": entrepreneurStats.BatchTime.Avg,
			"batch_size_avg": entrepreneurStats.BatchSize.Avg,
		},
	}
}
