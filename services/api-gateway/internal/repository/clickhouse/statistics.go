package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"go.uber.org/zap"
)

// StatisticsRepository репозиторий для работы со статистикой
type StatisticsRepository struct {
	client *Client
	logger *zap.Logger
}

// NewStatisticsRepository создает новый репозиторий статистики
func NewStatisticsRepository(client *Client, logger *zap.Logger) *StatisticsRepository {
	return &StatisticsRepository{
		client: client,
		logger: logger.Named("statistics_repo"),
	}
}

// GetStatistics получает общую статистику
func (r *StatisticsRepository) GetStatistics(ctx context.Context, filter *model.StatsFilter) (*model.Statistics, error) {
	stats := &model.Statistics{
		ByRegion:   make([]*model.RegionStatistics, 0),
		ByActivity: make([]*model.ActivityStatistics, 0),
	}

	// Общее количество компаний
	if err := r.getCompanyCounts(ctx, stats, filter); err != nil {
		return nil, err
	}

	// Общее количество ИП
	if err := r.getEntrepreneurCounts(ctx, stats, filter); err != nil {
		return nil, err
	}

	// Статистика регистраций
	if err := r.getRegistrationStats(ctx, stats); err != nil {
		return nil, err
	}

	// Статистика по регионам
	if err := r.getRegionStats(ctx, stats, filter); err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *StatisticsRepository) getCompanyCounts(ctx context.Context, stats *model.Statistics, filter *model.StatsFilter) error {
	baseQuery := `
		SELECT 
			count() as total,
			countIf(status = 'active') as active,
			countIf(status = 'liquidated') as liquidated
		FROM egrul.companies FINAL
	`

	whereClause := ""
	var args []interface{}

	if filter != nil {
		if filter.RegionCode != nil && *filter.RegionCode != "" {
			whereClause = "WHERE region_code = ?"
			args = append(args, *filter.RegionCode)
		}
	}

	query := baseQuery + whereClause

	row := r.client.conn.QueryRow(ctx, query, args...)
	
	var total, active, liquidated uint64
	if err := row.Scan(&total, &active, &liquidated); err != nil {
		return fmt.Errorf("get company counts: %w", err)
	}

	stats.TotalCompanies = int(total)
	stats.ActiveCompanies = int(active)
	stats.LiquidatedCompanies = int(liquidated)

	return nil
}

func (r *StatisticsRepository) getEntrepreneurCounts(ctx context.Context, stats *model.Statistics, filter *model.StatsFilter) error {
	baseQuery := `
		SELECT 
			count() as total,
			countIf(status = 'active') as active,
			countIf(status = 'liquidated') as liquidated
		FROM egrul.entrepreneurs FINAL
	`

	whereClause := ""
	var args []interface{}

	if filter != nil {
		if filter.RegionCode != nil && *filter.RegionCode != "" {
			whereClause = "WHERE region_code = ?"
			args = append(args, *filter.RegionCode)
		}
	}

	query := baseQuery + whereClause

	row := r.client.conn.QueryRow(ctx, query, args...)
	
	var total, active, liquidated uint64
	if err := row.Scan(&total, &active, &liquidated); err != nil {
		return fmt.Errorf("get entrepreneur counts: %w", err)
	}

	stats.TotalEntrepreneurs = int(total)
	stats.ActiveEntrepreneurs = int(active)
	stats.LiquidatedEntrepreneurs = int(liquidated)

	return nil
}

func (r *StatisticsRepository) getRegistrationStats(ctx context.Context, stats *model.Statistics) error {
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	query := `
		SELECT 
			countIf(registration_date >= ?) as today,
			countIf(registration_date >= ?) as month,
			countIf(registration_date >= ?) as year
		FROM (
			SELECT registration_date FROM egrul.companies FINAL
			WHERE registration_date IS NOT NULL
			UNION ALL
			SELECT registration_date FROM egrul.entrepreneurs FINAL
			WHERE registration_date IS NOT NULL
		)
	`

	row := r.client.conn.QueryRow(ctx, query, today, monthStart, yearStart)
	
	var todayCount, monthCount, yearCount uint64
	if err := row.Scan(&todayCount, &monthCount, &yearCount); err != nil {
		return fmt.Errorf("get registration stats: %w", err)
	}

	stats.RegisteredToday = int(todayCount)
	stats.RegisteredThisMonth = int(monthCount)
	stats.RegisteredThisYear = int(yearCount)

	return nil
}

func (r *StatisticsRepository) getRegionStats(ctx context.Context, stats *model.Statistics, filter *model.StatsFilter) error {
	query := `
		SELECT 
			region_code,
			any(region) as region_name,
			count() as total,
			countIf(status = 'active') as active,
			countIf(status = 'liquidated') as liquidated
		FROM egrul.companies FINAL
		WHERE region_code IS NOT NULL AND region_code != ''
		GROUP BY region_code
		ORDER BY total DESC
		LIMIT 20
	`

	rows, err := r.client.conn.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("get region stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var regionCode, regionName string
		var total, active, liquidated uint64
		
		if err := rows.Scan(&regionCode, &regionName, &total, &active, &liquidated); err != nil {
			return fmt.Errorf("scan region stats: %w", err)
		}

		stats.ByRegion = append(stats.ByRegion, &model.RegionStatistics{
			RegionCode:         regionCode,
			RegionName:         regionName,
			CompaniesCount:     int(total),
			EntrepreneursCount: 0, // TODO: добавить подсчет ИП
			ActiveCount:        int(active),
			LiquidatedCount:    int(liquidated),
		})
	}

	return nil
}

// GetActivityStats получает статистику по видам деятельности
func (r *StatisticsRepository) GetActivityStats(ctx context.Context, limit int) ([]*model.ActivityStatistics, error) {
	query := `
		SELECT 
			okved_main_code,
			any(okved_main_name) as okved_name,
			count() as companies_count
		FROM egrul.companies FINAL
		WHERE okved_main_code IS NOT NULL AND okved_main_code != ''
		GROUP BY okved_main_code
		ORDER BY companies_count DESC
		LIMIT ?
	`

	rows, err := r.client.conn.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get activity stats: %w", err)
	}
	defer rows.Close()

	var result []*model.ActivityStatistics
	for rows.Next() {
		var code, name string
		var count uint64
		
		if err := rows.Scan(&code, &name, &count); err != nil {
			return nil, fmt.Errorf("scan activity stats: %w", err)
		}

		result = append(result, &model.ActivityStatistics{
			OkvedCode:          code,
			OkvedName:          name,
			CompaniesCount:     int(count),
			EntrepreneursCount: 0,
		})
	}

	return result, nil
}

