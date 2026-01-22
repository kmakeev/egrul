package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/yourusername/egrul/services/sync-service/internal/config"
	"github.com/yourusername/egrul/services/sync-service/internal/mapper"
	"go.uber.org/zap"
)

type Reader struct {
	conn   clickhouse.Conn
	logger *zap.Logger
}

func NewReader(cfg config.ClickHouseConfig, logger *zap.Logger) (*Reader, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 300,
		},
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	logger.Info("Connected to ClickHouse",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database))

	return &Reader{
		conn:   conn,
		logger: logger,
	}, nil
}

func (r *Reader) Close() error {
	return r.conn.Close()
}

// ReadCompanies читает компании из ClickHouse
func (r *Reader) ReadCompanies(ctx context.Context, batchSize int, offset int) ([]mapper.CompanyRow, error) {
	query := `
		SELECT
			ogrn, inn, kpp, full_name, short_name, brand_name,
			status, region_code, region, city, full_address, email,
			okved_main_code, okved_main_name, okved_additional, okved_additional_names,
			head_last_name, head_first_name, head_middle_name, head_inn,
			opf_code, opf_name,
			registration_date, termination_date,
			updated_at
		FROM egrul.companies FINAL
		ORDER BY ogrn
		LIMIT ? OFFSET ?
	`

	rows, err := r.conn.Query(ctx, query, batchSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query companies: %w", err)
	}
	defer rows.Close()

	var companies []mapper.CompanyRow
	for rows.Next() {
		var company mapper.CompanyRow
		if err := rows.Scan(
			&company.OGRN, &company.INN, &company.KPP,
			&company.FullName, &company.ShortName, &company.BrandName,
			&company.Status, &company.RegionCode, &company.Region, &company.City,
			&company.FullAddress, &company.Email,
			&company.OKVEDMainCode, &company.OKVEDMainName,
			&company.OKVEDAdditional, &company.OKVEDAdditionalNames,
			&company.HeadLastName, &company.HeadFirstName, &company.HeadMiddleName,
			&company.HeadINN, &company.OPFCode, &company.OPFName,
			&company.RegistrationDate, &company.TerminationDate,
			&company.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan company: %w", err)
		}
		companies = append(companies, company)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating companies: %w", err)
	}

	return companies, nil
}

// ReadCompaniesUpdatedAfter читает компании измененные после указанной даты
func (r *Reader) ReadCompaniesUpdatedAfter(ctx context.Context, timestamp time.Time, batchSize int) ([]mapper.CompanyRow, error) {
	query := `
		SELECT
			ogrn, inn, kpp, full_name, short_name, brand_name,
			status, region_code, region, city, full_address, email,
			okved_main_code, okved_main_name, okved_additional, okved_additional_names,
			head_last_name, head_first_name, head_middle_name, head_inn,
			opf_code, opf_name,
			registration_date, termination_date,
			updated_at
		FROM egrul.companies FINAL
		WHERE updated_at > ?
		ORDER BY updated_at ASC
		LIMIT ?
	`

	rows, err := r.conn.Query(ctx, query, timestamp, batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to query updated companies: %w", err)
	}
	defer rows.Close()

	var companies []mapper.CompanyRow
	for rows.Next() {
		var company mapper.CompanyRow
		if err := rows.Scan(
			&company.OGRN, &company.INN, &company.KPP,
			&company.FullName, &company.ShortName, &company.BrandName,
			&company.Status, &company.RegionCode, &company.Region, &company.City,
			&company.FullAddress, &company.Email,
			&company.OKVEDMainCode, &company.OKVEDMainName,
			&company.OKVEDAdditional, &company.OKVEDAdditionalNames,
			&company.HeadLastName, &company.HeadFirstName, &company.HeadMiddleName,
			&company.HeadINN, &company.OPFCode, &company.OPFName,
			&company.RegistrationDate, &company.TerminationDate,
			&company.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan company: %w", err)
		}
		companies = append(companies, company)
	}

	return companies, rows.Err()
}

// ReadEntrepreneurs читает предпринимателей из ClickHouse
func (r *Reader) ReadEntrepreneurs(ctx context.Context, batchSize int, offset int) ([]mapper.EntrepreneurRow, error) {
	query := `
		SELECT
			ogrnip, inn, last_name, first_name, middle_name,
			gender, citizenship_type, status,
			region_code, region, city, full_address, email,
			okved_main_code, okved_main_name, okved_additional, okved_additional_names,
			registration_date, termination_date, is_bankrupt,
			updated_at
		FROM egrul.entrepreneurs FINAL
		ORDER BY ogrnip
		LIMIT ? OFFSET ?
	`

	rows, err := r.conn.Query(ctx, query, batchSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query entrepreneurs: %w", err)
	}
	defer rows.Close()

	var entrepreneurs []mapper.EntrepreneurRow
	for rows.Next() {
		var entr mapper.EntrepreneurRow
		if err := rows.Scan(
			&entr.OGRNIP, &entr.INN, &entr.LastName, &entr.FirstName, &entr.MiddleName,
			&entr.Gender, &entr.CitizenshipType, &entr.Status,
			&entr.RegionCode, &entr.Region, &entr.City, &entr.FullAddress, &entr.Email,
			&entr.OKVEDMainCode, &entr.OKVEDMainName, &entr.OKVEDAdditional, &entr.OKVEDAdditionalNames,
			&entr.RegistrationDate, &entr.TerminationDate, &entr.IsBankrupt,
			&entr.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan entrepreneur: %w", err)
		}
		entrepreneurs = append(entrepreneurs, entr)
	}

	return entrepreneurs, rows.Err()
}

// ReadEntrepreneursUpdatedAfter читает предпринимателей измененных после указанной даты
func (r *Reader) ReadEntrepreneursUpdatedAfter(ctx context.Context, timestamp time.Time, batchSize int) ([]mapper.EntrepreneurRow, error) {
	query := `
		SELECT
			ogrnip, inn, last_name, first_name, middle_name,
			gender, citizenship_type, status,
			region_code, region, city, full_address, email,
			okved_main_code, okved_main_name, okved_additional, okved_additional_names,
			registration_date, termination_date, is_bankrupt,
			updated_at
		FROM egrul.entrepreneurs FINAL
		WHERE updated_at > ?
		ORDER BY updated_at ASC
		LIMIT ?
	`

	rows, err := r.conn.Query(ctx, query, timestamp, batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to query updated entrepreneurs: %w", err)
	}
	defer rows.Close()

	var entrepreneurs []mapper.EntrepreneurRow
	for rows.Next() {
		var entr mapper.EntrepreneurRow
		if err := rows.Scan(
			&entr.OGRNIP, &entr.INN, &entr.LastName, &entr.FirstName, &entr.MiddleName,
			&entr.Gender, &entr.CitizenshipType, &entr.Status,
			&entr.RegionCode, &entr.Region, &entr.City, &entr.FullAddress, &entr.Email,
			&entr.OKVEDMainCode, &entr.OKVEDMainName, &entr.OKVEDAdditional, &entr.OKVEDAdditionalNames,
			&entr.RegistrationDate, &entr.TerminationDate, &entr.IsBankrupt,
			&entr.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan entrepreneur: %w", err)
		}
		entrepreneurs = append(entrepreneurs, entr)
	}

	return entrepreneurs, rows.Err()
}

// CountCompanies возвращает общее количество компаний
func (r *Reader) CountCompanies(ctx context.Context) (uint64, error) {
	var count uint64
	err := r.conn.QueryRow(ctx, "SELECT count() FROM egrul.companies FINAL").Scan(&count)
	return count, err
}

// CountEntrepreneurs возвращает общее количество предпринимателей
func (r *Reader) CountEntrepreneurs(ctx context.Context) (uint64, error) {
	var count uint64
	err := r.conn.QueryRow(ctx, "SELECT count() FROM egrul.entrepreneurs FINAL").Scan(&count)
	return count, err
}
