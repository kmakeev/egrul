package service

import (
	"context"
	"fmt"
	"time"

	"github.com/egrul/change-detection-service/internal/detector"
	"github.com/egrul/change-detection-service/internal/kafka"
	"github.com/egrul/change-detection-service/internal/model"
	"github.com/egrul/change-detection-service/internal/repository"
	"go.uber.org/zap"
)

// DetectionService отвечает за детектирование изменений в данных компаний и ИП
type DetectionService struct {
	companyRepo      repository.CompanyRepository
	entrepreneurRepo repository.EntrepreneurRepository
	changeRepo       repository.ChangeRepository
	comparator       *detector.Comparator
	kafkaProducer    *kafka.Producer
	logger           *zap.Logger
}

// NewDetectionService создает новый экземпляр DetectionService
func NewDetectionService(
	companyRepo repository.CompanyRepository,
	entrepreneurRepo repository.EntrepreneurRepository,
	changeRepo repository.ChangeRepository,
	comparator *detector.Comparator,
	kafkaProducer *kafka.Producer,
	logger *zap.Logger,
) *DetectionService {
	return &DetectionService{
		companyRepo:      companyRepo,
		entrepreneurRepo: entrepreneurRepo,
		changeRepo:       changeRepo,
		comparator:       comparator,
		kafkaProducer:    kafkaProducer,
		logger:           logger,
	}
}

// DetectCompanyChanges детектирует изменения для списка компаний
func (s *DetectionService) DetectCompanyChanges(ctx context.Context, ogrns []string) error {
	s.logger.Info("starting company change detection",
		zap.Int("count", len(ogrns)),
	)

	startTime := time.Now()
	totalChanges := 0
	processedCount := 0
	errorCount := 0

	// Обрабатываем компании батчами
	batchSize := 100
	for i := 0; i < len(ogrns); i += batchSize {
		end := i + batchSize
		if end > len(ogrns) {
			end = len(ogrns)
		}

		batch := ogrns[i:end]
		s.logger.Debug("processing company batch",
			zap.Int("batch_num", i/batchSize+1),
			zap.Int("batch_size", len(batch)),
		)

		// Получаем текущие данные компаний
		companies, err := s.companyRepo.GetByOGRNs(ctx, batch)
		if err != nil {
			s.logger.Error("failed to get companies batch", zap.Error(err))
			errorCount += len(batch)
			continue
		}

		// Для каждой компании детектируем изменения
		for _, newCompany := range companies {
			// Получаем предыдущую версию компании (с максимальной extract_date меньше текущей)
			oldCompany, err := s.companyRepo.GetPreviousByOGRN(ctx, newCompany.OGRN, newCompany.ExtractDate.Format("2006-01-02"))
			if err != nil {
				s.logger.Error("failed to get previous company version",
					zap.String("ogrn", newCompany.OGRN),
					zap.Error(err),
				)
				errorCount++
				continue
			}

			// Если нет предыдущей версии - это первая загрузка, пропускаем
			if oldCompany == nil {
				s.logger.Debug("no previous version found, skipping",
					zap.String("ogrn", newCompany.OGRN),
				)
				processedCount++
				continue
			}

			// Детектируем изменения между старой и новой версиями
			err = s.DetectCompanyChange(ctx, oldCompany, newCompany)
			if err != nil {
				s.logger.Error("failed to detect company changes",
					zap.String("ogrn", newCompany.OGRN),
					zap.Error(err),
				)
				errorCount++
				continue
			}

			processedCount++
			// Подсчитываем количество обнаруженных изменений (можно улучшить, сохраняя результат DetectCompanyChange)
			// Пока просто инкрементируем processedCount
		}
	}

	duration := time.Since(startTime)
	s.logger.Info("company change detection completed",
		zap.Int("total", len(ogrns)),
		zap.Int("processed", processedCount),
		zap.Int("errors", errorCount),
		zap.Int("changes_detected", totalChanges),
		zap.Duration("duration", duration),
	)

	return nil
}

// DetectCompanyChange детектирует изменения для одной компании
// oldCompany - старая версия, newCompany - новая версия
func (s *DetectionService) DetectCompanyChange(ctx context.Context, oldCompany, newCompany *model.Company) error {
	if oldCompany == nil {
		return fmt.Errorf("old company is nil")
	}
	if newCompany == nil {
		return fmt.Errorf("new company is nil")
	}

	s.logger.Debug("detecting company changes",
		zap.String("ogrn", oldCompany.OGRN),
		zap.String("name", oldCompany.FullName),
	)

	// Сравниваем старую и новую версии
	changes, err := s.comparator.CompareCompany(oldCompany, newCompany)
	if err != nil {
		return fmt.Errorf("failed to compare companies: %w", err)
	}

	if len(changes) == 0 {
		s.logger.Debug("no changes detected",
			zap.String("ogrn", oldCompany.OGRN),
		)
		return nil
	}

	// Проставляем timestamp детектирования
	now := time.Now()
	for _, change := range changes {
		change.DetectedAt = now
	}

	// Сохраняем изменения в ClickHouse
	err = s.changeRepo.SaveCompanyChanges(ctx, changes)
	if err != nil {
		s.logger.Error("failed to save company changes",
			zap.String("ogrn", oldCompany.OGRN),
			zap.Int("changes_count", len(changes)),
			zap.Error(err),
		)
		// Не возвращаем ошибку, продолжаем с отправкой в Kafka
	}

	// Отправляем события в Kafka
	err = s.kafkaProducer.SendCompanyChanges(ctx, changes)
	if err != nil {
		s.logger.Error("failed to send company changes to Kafka",
			zap.String("ogrn", oldCompany.OGRN),
			zap.Int("changes_count", len(changes)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send changes to Kafka: %w", err)
	}

	s.logger.Info("company changes detected and sent",
		zap.String("ogrn", oldCompany.OGRN),
		zap.String("name", oldCompany.FullName),
		zap.Int("changes_count", len(changes)),
	)

	return nil
}

// DetectEntrepreneurChanges детектирует изменения для списка ИП
func (s *DetectionService) DetectEntrepreneurChanges(ctx context.Context, ogrnips []string) error {
	s.logger.Info("starting entrepreneur change detection",
		zap.Int("count", len(ogrnips)),
	)

	startTime := time.Now()
	totalChanges := 0
	processedCount := 0
	errorCount := 0

	// Обрабатываем ИП батчами
	batchSize := 100
	for i := 0; i < len(ogrnips); i += batchSize {
		end := i + batchSize
		if end > len(ogrnips) {
			end = len(ogrnips)
		}

		batch := ogrnips[i:end]
		s.logger.Debug("processing entrepreneur batch",
			zap.Int("batch_num", i/batchSize+1),
			zap.Int("batch_size", len(batch)),
		)

		// Получаем текущие данные ИП
		entrepreneurs, err := s.entrepreneurRepo.GetByOGRNIPs(ctx, batch)
		if err != nil {
			s.logger.Error("failed to get entrepreneurs batch", zap.Error(err))
			errorCount += len(batch)
			continue
		}

		// Для каждого ИП детектируем изменения
		for _, newEntrepreneur := range entrepreneurs {
			// Получаем предыдущую версию ИП (с максимальной extract_date меньше текущей)
			oldEntrepreneur, err := s.entrepreneurRepo.GetPreviousByOGRNIP(ctx, newEntrepreneur.OGRNIP, newEntrepreneur.ExtractDate.Format("2006-01-02"))
			if err != nil {
				s.logger.Error("failed to get previous entrepreneur version",
					zap.String("ogrnip", newEntrepreneur.OGRNIP),
					zap.Error(err),
				)
				errorCount++
				continue
			}

			// Если нет предыдущей версии - это первая загрузка, пропускаем
			if oldEntrepreneur == nil {
				s.logger.Debug("no previous version found, skipping",
					zap.String("ogrnip", newEntrepreneur.OGRNIP),
				)
				processedCount++
				continue
			}

			// Детектируем изменения между старой и новой версиями
			err = s.DetectEntrepreneurChange(ctx, oldEntrepreneur, newEntrepreneur)
			if err != nil {
				s.logger.Error("failed to detect entrepreneur changes",
					zap.String("ogrnip", newEntrepreneur.OGRNIP),
					zap.Error(err),
				)
				errorCount++
				continue
			}

			processedCount++
		}
	}

	duration := time.Since(startTime)
	s.logger.Info("entrepreneur change detection completed",
		zap.Int("total", len(ogrnips)),
		zap.Int("processed", processedCount),
		zap.Int("errors", errorCount),
		zap.Int("changes_detected", totalChanges),
		zap.Duration("duration", duration),
	)

	return nil
}

// DetectEntrepreneurChange детектирует изменения для одного ИП
// oldEntrepreneur - старая версия, newEntrepreneur - новая версия
func (s *DetectionService) DetectEntrepreneurChange(ctx context.Context, oldEntrepreneur, newEntrepreneur *model.Entrepreneur) error {
	if oldEntrepreneur == nil {
		return fmt.Errorf("old entrepreneur is nil")
	}
	if newEntrepreneur == nil {
		return fmt.Errorf("new entrepreneur is nil")
	}

	s.logger.Debug("detecting entrepreneur changes",
		zap.String("ogrnip", oldEntrepreneur.OGRNIP),
		zap.String("name", oldEntrepreneur.FullName),
	)

	// Сравниваем старую и новую версии
	changes, err := s.comparator.CompareEntrepreneur(oldEntrepreneur, newEntrepreneur)
	if err != nil {
		return fmt.Errorf("failed to compare entrepreneurs: %w", err)
	}

	if len(changes) == 0 {
		s.logger.Debug("no changes detected",
			zap.String("ogrnip", oldEntrepreneur.OGRNIP),
		)
		return nil
	}

	// Проставляем timestamp детектирования
	now := time.Now()
	for _, change := range changes {
		change.DetectedAt = now
	}

	// Сохраняем изменения в ClickHouse
	err = s.changeRepo.SaveEntrepreneurChanges(ctx, changes)
	if err != nil {
		s.logger.Error("failed to save entrepreneur changes",
			zap.String("ogrnip", oldEntrepreneur.OGRNIP),
			zap.Int("changes_count", len(changes)),
			zap.Error(err),
		)
		// Не возвращаем ошибку, продолжаем с отправкой в Kafka
	}

	// Отправляем события в Kafka
	err = s.kafkaProducer.SendEntrepreneurChanges(ctx, changes)
	if err != nil {
		s.logger.Error("failed to send entrepreneur changes to Kafka",
			zap.String("ogrnip", oldEntrepreneur.OGRNIP),
			zap.Int("changes_count", len(changes)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send changes to Kafka: %w", err)
	}

	s.logger.Info("entrepreneur changes detected and sent",
		zap.String("ogrnip", oldEntrepreneur.OGRNIP),
		zap.String("name", oldEntrepreneur.FullName),
		zap.Int("changes_count", len(changes)),
	)

	return nil
}

// GetCompanyChanges возвращает историю изменений компании
func (s *DetectionService) GetCompanyChanges(ctx context.Context, ogrn string, limit int) ([]*model.ChangeEvent, error) {
	return s.changeRepo.GetCompanyChanges(ctx, ogrn, limit)
}

// GetEntrepreneurChanges возвращает историю изменений ИП
func (s *DetectionService) GetEntrepreneurChanges(ctx context.Context, ogrnip string, limit int) ([]*model.ChangeEvent, error) {
	return s.changeRepo.GetEntrepreneurChanges(ctx, ogrnip, limit)
}

// GetRecentChanges возвращает последние изменения за период
func (s *DetectionService) GetRecentChanges(ctx context.Context, entityType string, since int64) ([]*model.ChangeEvent, error) {
	return s.changeRepo.GetRecentChanges(ctx, entityType, since)
}

// GetStats возвращает статистику работы сервиса
func (s *DetectionService) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"kafka_producer": s.kafkaProducer.Stats(),
	}
}
