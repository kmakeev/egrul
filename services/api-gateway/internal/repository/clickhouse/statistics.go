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
	// Проверяем нужна ли фильтрация по ОКВЭД
	hasOkvedFilter := filter != nil && filter.Okved != nil && *filter.Okved != ""

	var query string
	var args []interface{}

	if hasOkvedFilter {
		// Если установлен ОКВЭД фильтр, используем прямой запрос к таблице
		// Логика статусов согласно company-status-badge.tsx:
		// - active: нет termination_date И код НЕ в списке недействующих
		// - liquidated: есть termination_date ИЛИ код в списке недействующих
		query = `
			SELECT
				count() as total,
				countIf(termination_date IS NULL AND status_code NOT IN ('101', '105', '106', '107', '113', '114', '115', '116', '117', '701', '702', '801', '802')) as active,
				countIf(termination_date IS NOT NULL OR status_code IN ('101', '105', '106', '107', '113', '114', '115', '116', '117', '701', '702', '801', '802')) as liquidated
			FROM egrul.companies_local FINAL
			WHERE (okved_main_code = ? OR has(okved_additional, ?))
		`
		args = append(args, *filter.Okved, *filter.Okved)

		// Добавляем регион если указан
		if filter.RegionCode != nil && *filter.RegionCode != "" {
			query += " AND region_code = ?"
			args = append(args, *filter.RegionCode)
		}
	} else {
		// Используем Materialized View для быстрой агрегации
		query = `
			SELECT
				countMerge(count) as total,
				countMergeIf(count, status = 'active') as active,
				countMergeIf(count, status IN ('liquidated', 'bankrupt')) as liquidated
			FROM egrul.stats_companies_by_region
		`

		// Добавляем регион если указан
		if filter != nil && filter.RegionCode != nil && *filter.RegionCode != "" {
			query += " WHERE region_code = ?"
			args = append(args, *filter.RegionCode)
		}
	}

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
	// Проверяем нужна ли фильтрация по ОКВЭД
	hasOkvedFilter := filter != nil && filter.Okved != nil && *filter.Okved != ""

	var query string
	var args []interface{}

	if hasOkvedFilter {
		// Если установлен ОКВЭД фильтр, используем прямой запрос к таблице
		// Логика статусов согласно entrepreneur-status-badge.tsx (миграция 013):
		// - active: НЕТ termination_date И НЕТ status_code (NULL)
		// - liquidated: ЕСТЬ termination_date ИЛИ ЕСТЬ любой status_code
		query = `
			SELECT
				count() as total,
				countIf(termination_date IS NULL AND status_code IS NULL) as active,
				countIf(termination_date IS NOT NULL OR status_code IS NOT NULL) as liquidated
			FROM egrul.entrepreneurs_local FINAL
			WHERE (okved_main_code = ? OR has(okved_additional, ?))
		`
		args = append(args, *filter.Okved, *filter.Okved)

		// Добавляем регион если указан
		if filter.RegionCode != nil && *filter.RegionCode != "" {
			query += " AND region_code = ?"
			args = append(args, *filter.RegionCode)
		}
	} else {
		// Используем Materialized View для быстрой агрегации
		query = `
			SELECT
				countMerge(count) as total,
				countMergeIf(count, status = 'active') as active,
				countMergeIf(count, status = 'liquidated') as liquidated
			FROM egrul.stats_entrepreneurs_by_region
		`

		// Добавляем регион если указан
		if filter != nil && filter.RegionCode != nil && *filter.RegionCode != "" {
			query += " WHERE region_code = ?"
			args = append(args, *filter.RegionCode)
		}
	}

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

// GetRegistrationsByMonth получает временной ряд регистраций и ликвидаций по месяцам
func (r *StatisticsRepository) GetRegistrationsByMonth(
	ctx context.Context,
	dateFrom, dateTo *time.Time,
	entityType *model.EntityType,
	filter *model.StatsFilter,
) ([]*model.TimeSeriesPoint, error) {
	// Устанавливаем дефолтные значения если не указаны
	var from, to time.Time
	if dateFrom != nil {
		from = *dateFrom
	} else {
		// По умолчанию - последний год
		from = time.Now().AddDate(-1, 0, 0)
	}
	if dateTo != nil {
		to = *dateTo
	} else {
		to = time.Now()
	}

	// Проверяем нужна ли фильтрация по региону или ОКВЭД
	hasRegionFilter := filter != nil && filter.RegionCode != nil && *filter.RegionCode != ""
	hasOkvedFilter := filter != nil && filter.Okved != nil && *filter.Okved != ""

	// Строим запрос в зависимости от фильтров
	var query string
	var args []interface{}

	// Если установлен любой фильтр (регион или ОКВЭД), используем прямые запросы к таблицам
	if hasRegionFilter || hasOkvedFilter {
		// Запрос напрямую из таблиц с фильтрами (медленнее, но с фильтрацией)
		var regionCode, okvedCode string
		if hasRegionFilter {
			regionCode = *filter.RegionCode
		}
		if hasOkvedFilter {
			okvedCode = *filter.Okved
		}

		if entityType != nil && *entityType == model.EntityTypeCompany {
			// Только компании
			query = `
				WITH
					registrations AS (
						SELECT
							toStartOfMonth(registration_date) as registration_month,
							count() as count
						FROM egrul.companies_local FINAL
						WHERE registration_date IS NOT NULL
						  AND (region_code = ? OR ? = '')
						  AND ((okved_main_code = ? OR has(okved_additional, ?)) OR ? = '')
						  AND registration_date >= ?
						  AND registration_date <= ?
						GROUP BY registration_month
					),
					terminations AS (
						SELECT
							toStartOfMonth(
								COALESCE(
									termination_date,
									multiIf(
										status_code IN ('101', '105', '106', '107', '113', '114', '115', '116', '117', '701', '702', '801', '802'),
										extract_date,
										NULL
									)
								)
							) as termination_month,
							count() as count
						FROM egrul.companies_local FINAL
						WHERE (region_code = ? OR ? = '')
						  AND ((okved_main_code = ? OR has(okved_additional, ?)) OR ? = '')
						  AND (
							  (termination_date IS NOT NULL AND termination_date >= ? AND termination_date <= ?)
							  OR (termination_date IS NULL AND status_code IN ('101', '105', '106', '107', '113', '114', '115', '116', '117', '701', '702', '801', '802') AND extract_date >= ? AND extract_date <= ?)
						  )
						GROUP BY termination_month
					)
				SELECT
					r.registration_month as month,
					r.count as registrations,
					coalesce(t.count, 0) as terminations,
					r.count - coalesce(t.count, 0) as net_growth
				FROM registrations r
				LEFT JOIN terminations t ON r.registration_month = t.termination_month
				ORDER BY r.registration_month
			`
			args = []interface{}{regionCode, regionCode, okvedCode, okvedCode, okvedCode, from, to, regionCode, regionCode, okvedCode, okvedCode, okvedCode, from, to, from, to}
		} else if entityType != nil && *entityType == model.EntityTypeEntrepreneur {
			// Только ИП
			query = `
				WITH
					registrations AS (
						SELECT
							toStartOfMonth(registration_date) as registration_month,
							count() as count
						FROM egrul.entrepreneurs_local FINAL
						WHERE registration_date IS NOT NULL
						  AND (region_code = ? OR ? = '')
						  AND ((okved_main_code = ? OR has(okved_additional, ?)) OR ? = '')
						  AND registration_date >= ?
						  AND registration_date <= ?
						GROUP BY registration_month
					),
					terminations AS (
						SELECT
							toStartOfMonth(termination_date) as termination_month,
							count() as count
						FROM egrul.entrepreneurs_local FINAL
						WHERE termination_date IS NOT NULL
						  AND (region_code = ? OR ? = '')
						  AND ((okved_main_code = ? OR has(okved_additional, ?)) OR ? = '')
						  AND termination_date >= ?
						  AND termination_date <= ?
						GROUP BY termination_month
					)
				SELECT
					r.registration_month as month,
					r.count as registrations,
					coalesce(t.count, 0) as terminations,
					r.count - coalesce(t.count, 0) as net_growth
				FROM registrations r
				LEFT JOIN terminations t ON r.registration_month = t.termination_month
				ORDER BY r.registration_month
			`
			args = []interface{}{regionCode, regionCode, okvedCode, okvedCode, okvedCode, from, to, regionCode, regionCode, okvedCode, okvedCode, okvedCode, from, to}
		} else {
			// Все типы (ЮЛ + ИП)
			query = `
				WITH
					registrations AS (
						SELECT
							toStartOfMonth(registration_date) as registration_month,
							count() as count
						FROM egrul.companies_local FINAL
						WHERE registration_date IS NOT NULL
						  AND (region_code = ? OR ? = '')
						  AND ((okved_main_code = ? OR has(okved_additional, ?)) OR ? = '')
						  AND registration_date >= ?
						  AND registration_date <= ?
						GROUP BY registration_month
						UNION ALL
						SELECT
							toStartOfMonth(registration_date) as registration_month,
							count() as count
						FROM egrul.entrepreneurs_local FINAL
						WHERE registration_date IS NOT NULL
						  AND (region_code = ? OR ? = '')
						  AND ((okved_main_code = ? OR has(okved_additional, ?)) OR ? = '')
						  AND registration_date >= ?
						  AND registration_date <= ?
						GROUP BY registration_month
					),
					registrations_agg AS (
						SELECT
							registration_month,
							sum(count) as count
						FROM registrations
						GROUP BY registration_month
					),
					terminations AS (
						SELECT
							toStartOfMonth(
								COALESCE(
									termination_date,
									multiIf(
										status_code IN ('101', '105', '106', '107', '113', '114', '115', '116', '117', '701', '702', '801', '802'),
										extract_date,
										NULL
									)
								)
							) as termination_month,
							count() as count
						FROM egrul.companies_local FINAL
						WHERE (region_code = ? OR ? = '')
						  AND ((okved_main_code = ? OR has(okved_additional, ?)) OR ? = '')
						  AND (
							  (termination_date IS NOT NULL AND termination_date >= ? AND termination_date <= ?)
							  OR (termination_date IS NULL AND status_code IN ('101', '105', '106', '107', '113', '114', '115', '116', '117', '701', '702', '801', '802') AND extract_date >= ? AND extract_date <= ?)
						  )
						GROUP BY termination_month
						UNION ALL
						SELECT
							toStartOfMonth(termination_date) as termination_month,
							count() as count
						FROM egrul.entrepreneurs_local FINAL
						WHERE termination_date IS NOT NULL
						  AND (region_code = ? OR ? = '')
						  AND ((okved_main_code = ? OR has(okved_additional, ?)) OR ? = '')
						  AND termination_date >= ?
						  AND termination_date <= ?
						GROUP BY termination_month
					),
					terminations_agg AS (
						SELECT
							termination_month,
							sum(count) as count
						FROM terminations
						GROUP BY termination_month
					)
				SELECT
					r.registration_month as month,
					r.count as registrations,
					coalesce(t.count, 0) as terminations,
					r.count - coalesce(t.count, 0) as net_growth
				FROM registrations_agg r
				LEFT JOIN terminations_agg t ON r.registration_month = t.termination_month
				ORDER BY r.registration_month
			`
			args = []interface{}{
				regionCode, regionCode, okvedCode, okvedCode, okvedCode, from, to,
				regionCode, regionCode, okvedCode, okvedCode, okvedCode, from, to,
				regionCode, regionCode, okvedCode, okvedCode, okvedCode, from, to, from, to,
				regionCode, regionCode, okvedCode, okvedCode, okvedCode, from, to,
			}
		}
	} else {
		// Используем MV (быстро, но без фильтра по региону)
		if entityType != nil {
			// Фильтруем по конкретному типу
			entityTypeStr := "company"
			switch *entityType {
			case model.EntityTypeCompany:
				entityTypeStr = "company"
			case model.EntityTypeEntrepreneur:
				entityTypeStr = "entrepreneur"
			}

			query = `
				WITH
					registrations AS (
						SELECT
							registration_month,
							countMerge(count) as count
						FROM egrul.stats_registrations_by_month
						WHERE entity_type = ?
						  AND registration_month >= toStartOfMonth(?)
						  AND registration_month <= toStartOfMonth(?)
						GROUP BY registration_month
					),
					terminations AS (
						SELECT
							termination_month,
							countMerge(count) as count
						FROM egrul.stats_terminations_by_month
						WHERE entity_type = ?
						  AND termination_month >= toStartOfMonth(?)
						  AND termination_month <= toStartOfMonth(?)
						GROUP BY termination_month
					)
				SELECT
					r.registration_month as month,
					r.count as registrations,
					coalesce(t.count, 0) as terminations,
					r.count - coalesce(t.count, 0) as net_growth
				FROM registrations r
				LEFT JOIN terminations t ON r.registration_month = t.termination_month
				ORDER BY r.registration_month
			`
			args = []interface{}{entityTypeStr, from, to, entityTypeStr, from, to}
		} else {
			// Агрегируем данные по всем типам (companies + entrepreneurs)
			// Используем UNION ALL чтобы избежать вложенной агрегации
			query = `
				WITH
					registrations_raw AS (
						SELECT
							registration_month,
							countMerge(count) as count
						FROM egrul.stats_registrations_by_month
						WHERE entity_type = 'company'
						  AND registration_month >= toStartOfMonth(?)
						  AND registration_month <= toStartOfMonth(?)
						GROUP BY registration_month
						UNION ALL
						SELECT
							registration_month,
							countMerge(count) as count
						FROM egrul.stats_registrations_by_month
						WHERE entity_type = 'entrepreneur'
						  AND registration_month >= toStartOfMonth(?)
						  AND registration_month <= toStartOfMonth(?)
						GROUP BY registration_month
					),
					registrations AS (
						SELECT
							registration_month,
							sum(count) as count
						FROM registrations_raw
						GROUP BY registration_month
					),
					terminations_raw AS (
						SELECT
							termination_month,
							countMerge(count) as count
						FROM egrul.stats_terminations_by_month
						WHERE entity_type = 'company'
						  AND termination_month >= toStartOfMonth(?)
						  AND termination_month <= toStartOfMonth(?)
						GROUP BY termination_month
						UNION ALL
						SELECT
							termination_month,
							countMerge(count) as count
						FROM egrul.stats_terminations_by_month
						WHERE entity_type = 'entrepreneur'
						  AND termination_month >= toStartOfMonth(?)
						  AND termination_month <= toStartOfMonth(?)
						GROUP BY termination_month
					),
					terminations AS (
						SELECT
							termination_month,
							sum(count) as count
						FROM terminations_raw
						GROUP BY termination_month
					)
				SELECT
					r.registration_month as month,
					r.count as registrations,
					coalesce(t.count, 0) as terminations,
					r.count - coalesce(t.count, 0) as net_growth
				FROM registrations r
				LEFT JOIN terminations t ON r.registration_month = t.termination_month
				ORDER BY r.registration_month
			`
			args = []interface{}{from, to, from, to, from, to, from, to}
		}
	}

	rows, err := r.client.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get registrations by month: %w", err)
	}
	defer rows.Close()

	var result []*model.TimeSeriesPoint
	for rows.Next() {
		var month time.Time
		var registrations, terminations uint64
		var netGrowth int64

		if err := rows.Scan(&month, &registrations, &terminations, &netGrowth); err != nil {
			return nil, fmt.Errorf("scan time series point: %w", err)
		}

		result = append(result, &model.TimeSeriesPoint{
			Month:              month.Format("2006-01-02"),
			RegistrationsCount: int(registrations),
			TerminationsCount:  int(terminations),
			NetGrowth:          int(netGrowth),
		})
	}

	return result, nil
}

// GetRegionHeatmap возвращает статистику для ВСЕХ регионов (для тепловой карты)
func (r *StatisticsRepository) GetRegionHeatmap(ctx context.Context) ([]*model.RegionStatistics, error) {
	// Объединяем статистику компаний и ИП используя MV
	query := `
		WITH
			companies_stats AS (
				SELECT
					region_code,
					any(region) as region_name,
					countMerge(count) as companies_count,
					countMergeIf(count, status = 'active') as active_companies,
					countMergeIf(count, status = 'liquidated') as liquidated_companies
				FROM egrul.stats_companies_by_region
				GROUP BY region_code
			),
			entrepreneurs_stats AS (
				SELECT
					region_code,
					countMerge(count) as entrepreneurs_count
				FROM egrul.stats_entrepreneurs_by_region
				WHERE status = 'active'
				GROUP BY region_code
			)
		SELECT
			c.region_code,
			c.region_name,
			c.companies_count,
			coalesce(e.entrepreneurs_count, 0) as entrepreneurs_count,
			c.active_companies,
			c.liquidated_companies
		FROM companies_stats c
		LEFT JOIN entrepreneurs_stats e ON c.region_code = e.region_code
		WHERE c.region_code IS NOT NULL AND c.region_code != ''
		ORDER BY c.region_code
	`

	rows, err := r.client.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get region heatmap: %w", err)
	}
	defer rows.Close()

	var result []*model.RegionStatistics
	for rows.Next() {
		var regionCode, regionName string
		var companiesCount, entrepreneursCount, activeCount, liquidatedCount uint64

		if err := rows.Scan(&regionCode, &regionName, &companiesCount, &entrepreneursCount, &activeCount, &liquidatedCount); err != nil {
			return nil, fmt.Errorf("scan region statistics: %w", err)
		}

		result = append(result, &model.RegionStatistics{
			RegionCode:         regionCode,
			RegionName:         regionName,
			CompaniesCount:     int(companiesCount),
			EntrepreneursCount: int(entrepreneursCount),
			ActiveCount:        int(activeCount),
			LiquidatedCount:    int(liquidatedCount),
		})
	}

	return result, nil
}

