package elasticsearch

import (
	"bytes"
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/yourusername/egrul/services/sync-service/internal/config"
	"github.com/yourusername/egrul/services/sync-service/internal/mapper"
	"go.uber.org/zap"
)

type Writer struct {
	client *elasticsearch.Client
	logger *zap.Logger
}

func NewWriter(cfg config.ElasticsearchConfig, logger *zap.Logger) (*Writer, error) {
	esCfg := elasticsearch.Config{
		Addresses: []string{cfg.URL},
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Проверить подключение
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get Elasticsearch info: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch returned error: %s", res.String())
	}

	logger.Info("Connected to Elasticsearch", zap.String("url", cfg.URL))

	return &Writer{
		client: client,
		logger: logger,
	}, nil
}

// BulkIndexCompanies индексирует компании bulk запросом
func (w *Writer) BulkIndexCompanies(ctx context.Context, companies []mapper.CompanyRow) (int, error) {
	if len(companies) == 0 {
		return 0, nil
	}

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client:     w.client,
		Index:      "egrul_companies",
		NumWorkers: 4,
		FlushBytes: 5 * 1024 * 1024, // 5MB
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create bulk indexer: %w", err)
	}

	successCount := 0
	errorCount := 0

	for _, company := range companies {
		docJSON, err := mapper.MapCompanyToES(company)
		if err != nil {
			w.logger.Error("Failed to map company",
				zap.String("ogrn", company.OGRN),
				zap.Error(err))
			errorCount++
			continue
		}

		err = bi.Add(ctx, esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: company.OGRN,
			Body:       bytes.NewReader(docJSON),
			OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
				successCount++
			},
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				if err != nil {
					w.logger.Error("Bulk indexer item failed",
						zap.String("document_id", item.DocumentID),
						zap.Error(err))
				} else {
					w.logger.Error("Bulk indexer item failed",
						zap.String("document_id", item.DocumentID),
						zap.String("error_type", res.Error.Type),
						zap.String("error_reason", res.Error.Reason))
				}
				errorCount++
			},
		})
		if err != nil {
			w.logger.Error("Failed to add item to bulk indexer",
				zap.String("ogrn", company.OGRN),
				zap.Error(err))
			errorCount++
		}
	}

	if err := bi.Close(ctx); err != nil {
		return successCount, fmt.Errorf("bulk indexer close error: %w", err)
	}

	stats := bi.Stats()
	w.logger.Info("Bulk index companies completed",
		zap.Int("total", len(companies)),
		zap.Uint64("indexed", stats.NumIndexed),
		zap.Uint64("failed", stats.NumFailed),
		zap.Int("errors", errorCount))

	return successCount, nil
}

// BulkIndexEntrepreneurs индексирует предпринимателей bulk запросом
func (w *Writer) BulkIndexEntrepreneurs(ctx context.Context, entrepreneurs []mapper.EntrepreneurRow) (int, error) {
	if len(entrepreneurs) == 0 {
		return 0, nil
	}

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client:     w.client,
		Index:      "egrul_entrepreneurs",
		NumWorkers: 4,
		FlushBytes: 5 * 1024 * 1024, // 5MB
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create bulk indexer: %w", err)
	}

	successCount := 0
	errorCount := 0

	for _, entr := range entrepreneurs {
		docJSON, err := mapper.MapEntrepreneurToES(entr)
		if err != nil {
			w.logger.Error("Failed to map entrepreneur",
				zap.String("ogrnip", entr.OGRNIP),
				zap.Error(err))
			errorCount++
			continue
		}

		err = bi.Add(ctx, esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: entr.OGRNIP,
			Body:       bytes.NewReader(docJSON),
			OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
				successCount++
			},
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				if err != nil {
					w.logger.Error("Bulk indexer item failed",
						zap.String("document_id", item.DocumentID),
						zap.Error(err))
				} else {
					w.logger.Error("Bulk indexer item failed",
						zap.String("document_id", item.DocumentID),
						zap.String("error_type", res.Error.Type),
						zap.String("error_reason", res.Error.Reason))
				}
				errorCount++
			},
		})
		if err != nil {
			w.logger.Error("Failed to add item to bulk indexer",
				zap.String("ogrnip", entr.OGRNIP),
				zap.Error(err))
			errorCount++
		}
	}

	if err := bi.Close(ctx); err != nil {
		return successCount, fmt.Errorf("bulk indexer close error: %w", err)
	}

	stats := bi.Stats()
	w.logger.Info("Bulk index entrepreneurs completed",
		zap.Int("total", len(entrepreneurs)),
		zap.Uint64("indexed", stats.NumIndexed),
		zap.Uint64("failed", stats.NumFailed),
		zap.Int("errors", errorCount))

	return successCount, nil
}

// DeleteDocument удаляет документ из индекса
func (w *Writer) DeleteDocument(ctx context.Context, index string, documentID string) error {
	res, err := w.client.Delete(index, documentID, w.client.Delete.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("delete request failed: %s", res.String())
	}

	return nil
}

// RefreshIndex принудительно обновляет индекс (для тестов)
func (w *Writer) RefreshIndex(ctx context.Context, index string) error {
	res, err := w.client.Indices.Refresh(
		w.client.Indices.Refresh.WithContext(ctx),
		w.client.Indices.Refresh.WithIndex(index),
	)
	if err != nil {
		return fmt.Errorf("failed to refresh index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("refresh request failed: %s", res.String())
	}

	w.logger.Debug("Index refreshed", zap.String("index", index))
	return nil
}
