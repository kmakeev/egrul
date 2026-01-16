package service

import (
	"context"
	"testing"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// stringPtr helper для создания указателя на строку
func stringPtr(s string) *string {
	return &s
}

// MockCompanyRepository мок для CompanyRepository
type MockCompanyRepository struct {
	mock.Mock
}

func (m *MockCompanyRepository) GetByOGRN(ctx context.Context, ogrn string) (*model.Company, error) {
	args := m.Called(ctx, ogrn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Company), args.Error(1)
}

func (m *MockCompanyRepository) GetByINN(ctx context.Context, inn string) (*model.Company, error) {
	args := m.Called(ctx, inn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Company), args.Error(1)
}

func (m *MockCompanyRepository) List(ctx context.Context, filter *model.CompanyFilter, pagination *model.Pagination, sort *model.CompanySort) ([]*model.Company, int, error) {
	args := m.Called(ctx, filter, pagination, sort)
	return args.Get(0).([]*model.Company), args.Int(1), args.Error(2)
}

func (m *MockCompanyRepository) Search(ctx context.Context, query string, limit, offset int) ([]*model.Company, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Company), args.Error(1)
}

// MockFounderRepository мок для FounderRepository
type MockFounderRepository struct {
	mock.Mock
}

func (m *MockFounderRepository) GetByCompanyOGRN(ctx context.Context, ogrn string, limit, offset int) ([]*model.Founder, error) {
	args := m.Called(ctx, ogrn, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Founder), args.Error(1)
}

func (m *MockFounderRepository) GetRelatedCompanies(ctx context.Context, inn string, limit, offset int) ([]string, error) {
	args := m.Called(ctx, inn, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFounderRepository) GetCompaniesWithCommonFounders(ctx context.Context, ogrn string, limit, offset int) ([]string, error) {
	args := m.Called(ctx, ogrn, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFounderRepository) GetFounderCompanies(ctx context.Context, ogrn string, limit, offset int) ([]string, error) {
	args := m.Called(ctx, ogrn, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFounderRepository) GetCommonFoundersDetails(ctx context.Context, ogrn1, ogrn2 string) ([]*model.Founder, error) {
	args := m.Called(ctx, ogrn1, ogrn2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Founder), args.Error(1)
}

func (m *MockFounderRepository) GetCompaniesWithCommonDirectors(ctx context.Context, ogrn string, limit, offset int) ([]string, error) {
	args := m.Called(ctx, ogrn, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFounderRepository) GetCommonDirectorsDetails(ctx context.Context, ogrn1, ogrn2 string) ([]*model.Person, error) {
	args := m.Called(ctx, ogrn1, ogrn2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Person), args.Error(1)
}

func (m *MockFounderRepository) GetCompaniesWhereFounderIsDirector(ctx context.Context, ogrn string, limit, offset int) ([]string, error) {
	args := m.Called(ctx, ogrn, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFounderRepository) GetCompaniesWhereDirectorIsFounder(ctx context.Context, ogrn string, limit, offset int) ([]string, error) {
	args := m.Called(ctx, ogrn, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFounderRepository) GetCrossPersonDetails(ctx context.Context, ogrn1, ogrn2 string, crossType string) ([]*model.Person, error) {
	args := m.Called(ctx, ogrn1, ogrn2, crossType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Person), args.Error(1)
}

func (m *MockFounderRepository) GetCompaniesWithCommonAddress(ctx context.Context, ogrn string, limit, offset int) ([]string, error) {
	args := m.Called(ctx, ogrn, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFounderRepository) GetCommonAddressDetails(ctx context.Context, ogrn1, ogrn2 string) (*model.Address, error) {
	args := m.Called(ctx, ogrn1, ogrn2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Address), args.Error(1)
}

// MockLicenseRepository мок для LicenseRepository
type MockLicenseRepository struct {
	mock.Mock
}

func (m *MockLicenseRepository) GetByEntityOGRN(ctx context.Context, ogrn string) ([]*model.License, error) {
	args := m.Called(ctx, ogrn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.License), args.Error(1)
}

// MockBranchRepository мок для BranchRepository
type MockBranchRepository struct {
	mock.Mock
}

func (m *MockBranchRepository) GetByCompanyOGRN(ctx context.Context, ogrn string) ([]*model.Branch, error) {
	args := m.Called(ctx, ogrn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Branch), args.Error(1)
}

// MockHistoryRepository мок для HistoryRepository
type MockHistoryRepository struct {
	mock.Mock
}

func (m *MockHistoryRepository) GetByEntityID(ctx context.Context, entityType, entityID string, limit, offset int) ([]*model.HistoryRecord, error) {
	args := m.Called(ctx, entityType, entityID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.HistoryRecord), args.Error(1)
}

func (m *MockHistoryRepository) CountByEntityID(ctx context.Context, entityType, entityID string) (int, error) {
	args := m.Called(ctx, entityType, entityID)
	return args.Int(0), args.Error(1)
}

func (m *MockHistoryRepository) InsertOrUpdate(ctx context.Context, record *model.HistoryRecord, entityType, entityID string, extractDate string, sourceFile string, fileHash string) error {
	args := m.Called(ctx, record, entityType, entityID, extractDate, sourceFile, fileHash)
	return args.Error(0)
}

// Тесты для CompanyService

func TestCompanyService_GetByOGRN_Success(t *testing.T) {
	// Arrange
	mockCompanyRepo := new(MockCompanyRepository)
	logger := zap.NewNop()

	expected := &model.Company{
		Ogrn:     "1234567890123",
		Inn:      "7707083893",
		FullName: "ООО ТЕСТ",
		Status:   model.EntityStatusActive,
	}

	mockCompanyRepo.On("GetByOGRN", mock.Anything, "1234567890123").Return(expected, nil)

	service := NewCompanyService(mockCompanyRepo, nil, nil, nil, nil, logger)

	// Act
	result, err := service.GetByOGRN(context.Background(), "1234567890123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expected.Ogrn, result.Ogrn)
	assert.Equal(t, expected.Inn, result.Inn)
	assert.Equal(t, expected.FullName, result.FullName)
	mockCompanyRepo.AssertExpectations(t)
}

func TestCompanyService_GetByOGRN_NotFound(t *testing.T) {
	// Arrange
	mockCompanyRepo := new(MockCompanyRepository)
	logger := zap.NewNop()

	mockCompanyRepo.On("GetByOGRN", mock.Anything, "9999999999999").Return(nil, nil)

	service := NewCompanyService(mockCompanyRepo, nil, nil, nil, nil, logger)

	// Act
	result, err := service.GetByOGRN(context.Background(), "9999999999999")

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, result)
	mockCompanyRepo.AssertExpectations(t)
}

func TestCompanyService_GetByINN_Success(t *testing.T) {
	// Arrange
	mockCompanyRepo := new(MockCompanyRepository)
	logger := zap.NewNop()

	expected := &model.Company{
		Ogrn:     "1234567890123",
		Inn:      "7707083893",
		FullName: "ООО ТЕСТ",
		Status:   model.EntityStatusActive,
	}

	mockCompanyRepo.On("GetByINN", mock.Anything, "7707083893").Return(expected, nil)

	service := NewCompanyService(mockCompanyRepo, nil, nil, nil, nil, logger)

	// Act
	result, err := service.GetByINN(context.Background(), "7707083893")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expected.Inn, result.Inn)
	assert.Equal(t, expected.FullName, result.FullName)
	mockCompanyRepo.AssertExpectations(t)
}

func TestCompanyService_Search_ValidatesLimit(t *testing.T) {
	// Arrange
	mockCompanyRepo := new(MockCompanyRepository)
	logger := zap.NewNop()

	companies := []*model.Company{
		{Ogrn: "1234567890123", Inn: "7707083893", FullName: "ООО ТЕСТ 1"},
		{Ogrn: "1234567890124", Inn: "7707083894", FullName: "ООО ТЕСТ 2"},
	}

	// При передаче limit=0 должен быть установлен default=20
	mockCompanyRepo.On("Search", mock.Anything, "тест", 20, 0).Return(companies, nil)

	service := NewCompanyService(mockCompanyRepo, nil, nil, nil, nil, logger)

	// Act
	result, err := service.Search(context.Background(), "тест", 0, 0)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result))
	mockCompanyRepo.AssertExpectations(t)
}

func TestCompanyService_Search_LimitsMaxValue(t *testing.T) {
	// Arrange
	mockCompanyRepo := new(MockCompanyRepository)
	logger := zap.NewNop()

	companies := []*model.Company{
		{Ogrn: "1234567890123", Inn: "7707083893", FullName: "ООО ТЕСТ"},
	}

	// При передаче limit=500 должен быть ограничен до 100
	mockCompanyRepo.On("Search", mock.Anything, "тест", 100, 0).Return(companies, nil)

	service := NewCompanyService(mockCompanyRepo, nil, nil, nil, nil, logger)

	// Act
	result, err := service.Search(context.Background(), "тест", 500, 0)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockCompanyRepo.AssertExpectations(t)
}

func TestCompanyService_GetFounders_Success(t *testing.T) {
	// Arrange
	mockFounderRepo := new(MockFounderRepository)
	logger := zap.NewNop()

	founders := []*model.Founder{
		{Ogrn: stringPtr("1234567890123"), Type: model.FounderTypePerson, Name: "Test Founder"},
	}

	mockFounderRepo.On("GetByCompanyOGRN", mock.Anything, "1234567890123", 100, 0).Return(founders, nil)

	service := NewCompanyService(nil, mockFounderRepo, nil, nil, nil, logger)

	// Act
	result, err := service.GetFounders(context.Background(), "1234567890123", 0, 0)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result))
	mockFounderRepo.AssertExpectations(t)
}

func TestCompanyService_GetLicenses_Success(t *testing.T) {
	// Arrange
	mockLicenseRepo := new(MockLicenseRepository)
	logger := zap.NewNop()

	licenses := []*model.License{
		{ID: "1", Series: stringPtr("77"), Number: "123456"},
	}

	mockLicenseRepo.On("GetByEntityOGRN", mock.Anything, "1234567890123").Return(licenses, nil)

	service := NewCompanyService(nil, nil, mockLicenseRepo, nil, nil, logger)

	// Act
	result, err := service.GetLicenses(context.Background(), "1234567890123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result))
	mockLicenseRepo.AssertExpectations(t)
}

func TestCompanyService_GetBranches_Success(t *testing.T) {
	// Arrange
	mockBranchRepo := new(MockBranchRepository)
	logger := zap.NewNop()

	branches := []*model.Branch{
		{ID: "1", Name: stringPtr("Филиал №1"), Type: model.BranchTypeBranch},
	}

	mockBranchRepo.On("GetByCompanyOGRN", mock.Anything, "1234567890123").Return(branches, nil)

	service := NewCompanyService(nil, nil, nil, mockBranchRepo, nil, logger)

	// Act
	result, err := service.GetBranches(context.Background(), "1234567890123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result))
	mockBranchRepo.AssertExpectations(t)
}

func TestCompanyService_GetHistory_Success(t *testing.T) {
	// Arrange
	mockHistoryRepo := new(MockHistoryRepository)
	logger := zap.NewNop()

	history := []*model.HistoryRecord{
		{ID: "1", Grn: "GRN123"},
	}

	mockHistoryRepo.On("GetByEntityID", mock.Anything, "company", "1234567890123", 50, 0).Return(history, nil)

	service := NewCompanyService(nil, nil, nil, nil, mockHistoryRepo, logger)

	// Act
	result, err := service.GetHistory(context.Background(), "1234567890123", 0, 0)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result))
	mockHistoryRepo.AssertExpectations(t)
}

func TestCompanyService_GetHistoryCount_Success(t *testing.T) {
	// Arrange
	mockHistoryRepo := new(MockHistoryRepository)
	logger := zap.NewNop()

	mockHistoryRepo.On("CountByEntityID", mock.Anything, "company", "1234567890123").Return(42, nil)

	service := NewCompanyService(nil, nil, nil, nil, mockHistoryRepo, logger)

	// Act
	result, err := service.GetHistoryCount(context.Background(), "1234567890123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 42, result)
	mockHistoryRepo.AssertExpectations(t)
}
