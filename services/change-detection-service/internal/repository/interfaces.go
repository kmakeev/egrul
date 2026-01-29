package repository

import (
	"context"

	"github.com/egrul/change-detection-service/internal/model"
)

// CompanyRepository определяет интерфейс для работы с данными компаний
type CompanyRepository interface {
	// GetByOGRN возвращает текущую версию компании по ОГРН
	GetByOGRN(ctx context.Context, ogrn string) (*model.Company, error)

	// GetByOGRNs возвращает несколько компаний по списку ОГРН (для батчинга)
	GetByOGRNs(ctx context.Context, ogrns []string) ([]*model.Company, error)

	// GetFounders возвращает учредителей компании
	GetFounders(ctx context.Context, ogrn string) ([]model.Founder, error)

	// GetActivities возвращает виды деятельности компании
	GetActivities(ctx context.Context, ogrn string) (mainOKVED string, additionalOKVED []string, err error)

	// GetLicensesCount возвращает количество лицензий компании
	GetLicensesCount(ctx context.Context, ogrn string) (int, error)

	// GetBranchesCount возвращает количество филиалов компании
	GetBranchesCount(ctx context.Context, ogrn string) (int, error)
}

// EntrepreneurRepository определяет интерфейс для работы с данными ИП
type EntrepreneurRepository interface {
	// GetByOGRNIP возвращает текущую версию ИП по ОГРНИП
	GetByOGRNIP(ctx context.Context, ogrnip string) (*model.Entrepreneur, error)

	// GetByOGRNIPs возвращает несколько ИП по списку ОГРНИП (для батчинга)
	GetByOGRNIPs(ctx context.Context, ogrnips []string) ([]*model.Entrepreneur, error)

	// GetActivities возвращает виды деятельности ИП
	GetActivities(ctx context.Context, ogrnip string) (mainOKVED string, additionalOKVED []string, err error)

	// GetLicensesCount возвращает количество лицензий ИП
	GetLicensesCount(ctx context.Context, ogrnip string) (int, error)
}

// ChangeRepository определяет интерфейс для сохранения событий изменений
type ChangeRepository interface {
	// SaveCompanyChange сохраняет событие изменения компании
	SaveCompanyChange(ctx context.Context, change *model.ChangeEvent) error

	// SaveCompanyChanges сохраняет несколько событий изменений компаний (батчинг)
	SaveCompanyChanges(ctx context.Context, changes []*model.ChangeEvent) error

	// SaveEntrepreneurChange сохраняет событие изменения ИП
	SaveEntrepreneurChange(ctx context.Context, change *model.ChangeEvent) error

	// SaveEntrepreneurChanges сохраняет несколько событий изменений ИП (батчинг)
	SaveEntrepreneurChanges(ctx context.Context, changes []*model.ChangeEvent) error

	// GetCompanyChanges возвращает историю изменений компании
	GetCompanyChanges(ctx context.Context, ogrn string, limit int) ([]*model.ChangeEvent, error)

	// GetEntrepreneurChanges возвращает историю изменений ИП
	GetEntrepreneurChanges(ctx context.Context, ogrnip string, limit int) ([]*model.ChangeEvent, error)

	// GetRecentChanges возвращает последние изменения за период (для тестирования)
	GetRecentChanges(ctx context.Context, entityType string, since int64) ([]*model.ChangeEvent, error)
}
