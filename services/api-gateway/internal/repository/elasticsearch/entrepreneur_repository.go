package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"github.com/elastic/go-elasticsearch/v8"
	"go.uber.org/zap"
)

// ESEntrepreneurRepository репозиторий для работы с ИП через Elasticsearch
type ESEntrepreneurRepository struct {
	client *elasticsearch.Client
	logger *zap.Logger
}

// NewESEntrepreneurRepository создает новый Elasticsearch репозиторий ИП
func NewESEntrepreneurRepository(client *elasticsearch.Client, logger *zap.Logger) *ESEntrepreneurRepository {
	return &ESEntrepreneurRepository{
		client: client,
		logger: logger.Named("es_entrepreneur_repo"),
	}
}

// Search выполняет полнотекстовый поиск ИП в Elasticsearch
func (r *ESEntrepreneurRepository) Search(ctx context.Context, query string, limit, offset int) ([]*model.Entrepreneur, error) {
	// Построение multi-match запроса с boost
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					// Exact match по ОГРНИП (boost 100)
					{
						"term": map[string]interface{}{
							"ogrnip": map[string]interface{}{
								"value": query,
								"boost": 100,
							},
						},
					},
					// Exact match по ИНН (boost 100)
					{
						"term": map[string]interface{}{
							"inn": map[string]interface{}{
								"value": query,
								"boost": 100,
							},
						},
					},
					// Морфологический поиск по полному имени (boost 10)
					{
						"match": map[string]interface{}{
							"full_name": map[string]interface{}{
								"query": query,
								"boost": 10,
							},
						},
					},
					// Поиск по фамилии (boost 8)
					{
						"match": map[string]interface{}{
							"last_name": map[string]interface{}{
								"query": query,
								"boost": 8,
							},
						},
					},
					// Поиск по имени (boost 5)
					{
						"match": map[string]interface{}{
							"first_name": map[string]interface{}{
								"query": query,
								"boost": 5,
							},
						},
					},
					// Поиск по отчеству (boost 3)
					{
						"match": map[string]interface{}{
							"middle_name": map[string]interface{}{
								"query": query,
								"boost": 3,
							},
						},
					},
				},
			},
		},
		"from": offset,
		"size": limit,
	}

	// Сериализация запроса
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, fmt.Errorf("encode search query: %w", err)
	}

	r.logger.Debug("Elasticsearch search query",
		zap.String("query", query),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.String("body", buf.String()))

	// Выполнение запроса
	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("egrul_entrepreneurs"),
		r.client.Search.WithBody(&buf),
		r.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		r.logger.Error("Elasticsearch search error",
			zap.String("status", res.Status()),
			zap.String("response", string(bodyBytes)))
		return nil, fmt.Errorf("elasticsearch returned error: %s", res.Status())
	}

	// Парсинг ответа
	entrepreneurs, err := r.parseSearchResponse(res.Body)
	if err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	r.logger.Info("Elasticsearch search completed",
		zap.String("query", query),
		zap.Int("results", len(entrepreneurs)))

	return entrepreneurs, nil
}

// parseSearchResponse парсит ответ Elasticsearch в модели Entrepreneur
func (r *ESEntrepreneurRepository) parseSearchResponse(body io.Reader) ([]*model.Entrepreneur, error) {
	entrepreneurs, _, err := r.parseSearchResponseWithTotal(body)
	return entrepreneurs, err
}

// esEntrepreneurDocument структура документа ИП в Elasticsearch
type esEntrepreneurDocument struct {
	OGRNIP           string    `json:"ogrnip"`
	INN              string    `json:"inn"`
	LastName         string    `json:"last_name"`
	FirstName        string    `json:"first_name"`
	MiddleName       *string   `json:"middle_name"`
	FullName         string    `json:"full_name"`
	Status           string    `json:"status"`
	RegionCode       *string   `json:"region_code"`
	OkvedMainCode    *string   `json:"okved_main_code"`
	OkvedMainName    *string   `json:"okved_main_name"`
	OkvedAdditional  []string  `json:"okved_additional"`
	RegistrationDate *string   `json:"registration_date"`
	UpdatedAt        string    `json:"updated_at"`
}

// toModel преобразует ES документ в модель GraphQL
func (doc *esEntrepreneurDocument) toModel() *model.Entrepreneur {
	entrepreneur := &model.Entrepreneur{
		Ogrnip:     doc.OGRNIP,
		Inn:        doc.INN,
		LastName:   doc.LastName,
		FirstName:  doc.FirstName,
		Status:     model.ParseEntityStatus(doc.Status),
		Activities: make([]*model.Activity, 0),
	}

	// Nullable fields
	if doc.MiddleName != nil {
		entrepreneur.MiddleName = doc.MiddleName
	}

	// Address
	if doc.RegionCode != nil {
		entrepreneur.Address = &model.Address{
			RegionCode: doc.RegionCode,
		}
	}

	// Main activity
	if doc.OkvedMainCode != nil {
		entrepreneur.MainActivity = &model.Activity{
			Code:   *doc.OkvedMainCode,
			IsMain: true,
		}
		if doc.OkvedMainName != nil {
			entrepreneur.MainActivity.Name = doc.OkvedMainName
		}
		entrepreneur.Activities = append(entrepreneur.Activities, entrepreneur.MainActivity)
	}

	// Additional activities
	for _, code := range doc.OkvedAdditional {
		activity := &model.Activity{
			Code:   code,
			IsMain: false,
		}
		entrepreneur.Activities = append(entrepreneur.Activities, activity)
	}

	// Dates
	if doc.RegistrationDate != nil {
		if t, err := time.Parse(time.RFC3339, *doc.RegistrationDate); err == nil {
			entrepreneur.RegistrationDate = &model.Date{Time: t}
		}
	}
	if t, err := time.Parse(time.RFC3339, doc.UpdatedAt); err == nil {
		entrepreneur.UpdatedAt = t
	}

	// Placeholder values для обязательных полей
	entrepreneur.VersionDate = model.Date{Time: time.Now()}
	entrepreneur.CreatedAt = time.Now()

	return entrepreneur
}

// GetByOGRNIP не поддерживается в Elasticsearch репозитории
// Для точных совпадений используется ClickHouse
func (r *ESEntrepreneurRepository) GetByOGRNIP(ctx context.Context, ogrnip string) (*model.Entrepreneur, error) {
	return nil, fmt.Errorf("GetByOGRNIP not supported in Elasticsearch repository, use ClickHouse instead")
}

// GetByINN не поддерживается в Elasticsearch репозитории
// Для точных совпадений используется ClickHouse
func (r *ESEntrepreneurRepository) GetByINN(ctx context.Context, inn string) (*model.Entrepreneur, error) {
	return nil, fmt.Errorf("GetByINN not supported in Elasticsearch repository, use ClickHouse instead")
}

// List не поддерживается в Elasticsearch репозитории
// Для фильтрации и пагинации используется ClickHouse
func (r *ESEntrepreneurRepository) List(ctx context.Context, filter *model.EntrepreneurFilter, pagination *model.Pagination, sort *model.EntrepreneurSort) ([]*model.Entrepreneur, int, error) {
	return nil, 0, fmt.Errorf("List not supported in Elasticsearch repository, use ClickHouse instead")
}

// SearchWithTotal выполняет поиск и возвращает ИП с общим количеством найденных
func (r *ESEntrepreneurRepository) SearchWithTotal(ctx context.Context, query string, limit, offset int) ([]*model.Entrepreneur, int, error) {
	return r.SearchWithTotalAndFilters(ctx, query, nil, limit, offset)
}

// SearchWithTotalAndFilters выполняет поиск с фильтрами и возвращает ИП с общим количеством
func (r *ESEntrepreneurRepository) SearchWithTotalAndFilters(ctx context.Context, query string, filter *model.EntrepreneurFilter, limit, offset int) ([]*model.Entrepreneur, int, error) {
	// Build bool query with should (for text search) and must (for filters)
	boolQuery := map[string]interface{}{}

	// Если query задан, добавляем текстовый поиск
	if query != "" {
		boolQuery["should"] = []map[string]interface{}{
			// Exact match по ОГРНИП (boost 100)
			{
				"term": map[string]interface{}{
					"ogrnip": map[string]interface{}{
						"value": query,
						"boost": 100,
					},
				},
			},
			// Exact match по ИНН (boost 100)
			{
				"term": map[string]interface{}{
					"inn": map[string]interface{}{
						"value": query,
						"boost": 100,
					},
				},
			},
			// Морфологический поиск по ФИО (boost 10)
			{
				"multi_match": map[string]interface{}{
					"query":  query,
					"fields": []string{"last_name^3", "first_name^2", "middle_name"},
					"boost":  10,
				},
			},
		}
		boolQuery["minimum_should_match"] = 1
	}

	// Add filters if provided
	if filter != nil {
		mustClauses := []map[string]interface{}{}

		if filter.RegionCode != nil && *filter.RegionCode != "" {
			mustClauses = append(mustClauses, map[string]interface{}{
				"term": map[string]interface{}{
					"region_code": *filter.RegionCode,
				},
			})
		}

		if filter.Status != nil {
			mustClauses = append(mustClauses, map[string]interface{}{
				"term": map[string]interface{}{
					// Статусы в ES хранятся в lowercase
					"status": strings.ToLower(string(*filter.Status)),
				},
			})
		}

		if filter.StatusIn != nil && len(filter.StatusIn) > 0 {
			statuses := make([]string, len(filter.StatusIn))
			for i, s := range filter.StatusIn {
				// Статусы в ES хранятся в lowercase
				statuses[i] = strings.ToLower(string(s))
			}
			mustClauses = append(mustClauses, map[string]interface{}{
				"terms": map[string]interface{}{
					"status": statuses,
				},
			})
		}

		if filter.Okved != nil && *filter.Okved != "" {
			mustClauses = append(mustClauses, map[string]interface{}{
				"bool": map[string]interface{}{
					"should": []map[string]interface{}{
						{
							"prefix": map[string]interface{}{
								"okved_main_code": *filter.Okved,
							},
						},
						{
							"prefix": map[string]interface{}{
								"okved_additional": *filter.Okved,
							},
						},
					},
					"minimum_should_match": 1,
				},
			})
		}

		// Date range filters for registration_date
		if filter.RegisteredAfter != nil || filter.RegisteredBefore != nil {
			rangeClause := map[string]interface{}{}
			if filter.RegisteredAfter != nil {
				rangeClause["gte"] = filter.RegisteredAfter.Time.Format("2006-01-02")
			}
			if filter.RegisteredBefore != nil {
				rangeClause["lte"] = filter.RegisteredBefore.Time.Format("2006-01-02")
			}
			mustClauses = append(mustClauses, map[string]interface{}{
				"range": map[string]interface{}{
					"registration_date": rangeClause,
				},
			})
		}

		if len(mustClauses) > 0 {
			boolQuery["must"] = mustClauses
		}
	}

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": boolQuery,
		},
		"from": offset,
		"size": limit,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, 0, fmt.Errorf("encode search query: %w", err)
	}

	// Execute search
	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("egrul_entrepreneurs"),
		r.client.Search.WithBody(&buf),
		r.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("elasticsearch search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		r.logger.Error("Elasticsearch search error",
			zap.String("status", res.Status()),
			zap.String("response", string(bodyBytes)))
		return nil, 0, fmt.Errorf("elasticsearch returned error: %s", res.Status())
	}

	// Parse response with total
	entrepreneurs, total, err := r.parseSearchResponseWithTotal(res.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("parse search response: %w", err)
	}

	r.logger.Info("Elasticsearch search with filters completed",
		zap.String("query", query),
		zap.Int("results", len(entrepreneurs)),
		zap.Int("total", total))

	return entrepreneurs, total, nil
}

// parseSearchResponseWithTotal парсит ответ Elasticsearch в модели Entrepreneur с общим количеством
func (r *ESEntrepreneurRepository) parseSearchResponseWithTotal(body io.Reader) ([]*model.Entrepreneur, int, error) {
	var esResponse struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source esEntrepreneurDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(body).Decode(&esResponse); err != nil {
		return nil, 0, fmt.Errorf("decode elasticsearch response: %w", err)
	}

	entrepreneurs := make([]*model.Entrepreneur, 0, len(esResponse.Hits.Hits))
	for _, hit := range esResponse.Hits.Hits {
		entrepreneur := hit.Source.toModel()
		entrepreneurs = append(entrepreneurs, entrepreneur)
	}

	return entrepreneurs, esResponse.Hits.Total.Value, nil
}
