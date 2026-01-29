package detector

import (
	"encoding/json"
	"fmt"

	"github.com/egrul/change-detection-service/internal/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Comparator отвечает за сравнение старых и новых данных сущностей
type Comparator struct {
	logger     *zap.Logger
	classifier *Classifier
}

// NewComparator создает новый экземпляр Comparator
func NewComparator(logger *zap.Logger, classifier *Classifier) *Comparator {
	return &Comparator{
		logger:     logger,
		classifier: classifier,
	}
}

// CompareCompany сравнивает старую и новую версии компании
func (c *Comparator) CompareCompany(old, new *model.Company) ([]*model.ChangeEvent, error) {
	if old == nil && new == nil {
		return nil, fmt.Errorf("both old and new companies are nil")
	}
	if old == nil {
		// Новая компания, не считается изменением
		return []*model.ChangeEvent{}, nil
	}
	if new == nil {
		return nil, fmt.Errorf("new company is nil")
	}

	var changes []*model.ChangeEvent

	// Сравнение статуса
	if statusChange := c.compareStatus(old.Status, new.Status, old.OGRN, old.FullName, old.RegionCode, old.INN); statusChange != nil {
		changes = append(changes, statusChange)
	}

	// Сравнение руководителя
	if directorChange := c.compareDirector(old, new); directorChange != nil {
		changes = append(changes, directorChange)
	}

	// Сравнение учредителей
	founderChanges := c.compareFounders(old, new)
	changes = append(changes, founderChanges...)

	// Сравнение адреса
	if addressChange := c.compareAddress(old, new); addressChange != nil {
		changes = append(changes, addressChange)
	}

	// Сравнение уставного капитала
	if capitalChange := c.compareCapital(old, new); capitalChange != nil {
		changes = append(changes, capitalChange)
	}

	// Сравнение видов деятельности (ОКВЭД)
	activityChanges := c.compareActivities(old, new)
	changes = append(changes, activityChanges...)

	// Сравнение лицензий
	if licenseChange := c.compareLicensesCount(old, new); licenseChange != nil {
		changes = append(changes, licenseChange)
	}

	// Сравнение филиалов
	if branchChange := c.compareBranchesCount(old, new); branchChange != nil {
		changes = append(changes, branchChange)
	}

	c.logger.Debug("compared companies",
		zap.String("ogrn", old.OGRN),
		zap.Int("changes_count", len(changes)),
	)

	return changes, nil
}

// CompareEntrepreneur сравнивает старую и новую версии ИП
func (c *Comparator) CompareEntrepreneur(old, new *model.Entrepreneur) ([]*model.ChangeEvent, error) {
	if old == nil && new == nil {
		return nil, fmt.Errorf("both old and new entrepreneurs are nil")
	}
	if old == nil {
		// Новый ИП, не считается изменением
		return []*model.ChangeEvent{}, nil
	}
	if new == nil {
		return nil, fmt.Errorf("new entrepreneur is nil")
	}

	var changes []*model.ChangeEvent

	// Сравнение статуса
	if statusChange := c.compareIPStatus(old.Status, new.Status, old.OGRNIP, old.FullName, old.RegionCode, old.INN); statusChange != nil {
		changes = append(changes, statusChange)
	}

	// Сравнение адреса
	if addressChange := c.compareIPAddress(old, new); addressChange != nil {
		changes = append(changes, addressChange)
	}

	// Сравнение видов деятельности (ОКВЭД)
	activityChanges := c.compareIPActivities(old, new)
	changes = append(changes, activityChanges...)

	// Сравнение лицензий
	if licenseChange := c.compareIPLicensesCount(old, new); licenseChange != nil {
		changes = append(changes, licenseChange)
	}

	c.logger.Debug("compared entrepreneurs",
		zap.String("ogrnip", old.OGRNIP),
		zap.Int("changes_count", len(changes)),
	)

	return changes, nil
}

// compareStatus сравнивает статусы компании
func (c *Comparator) compareStatus(oldStatus, newStatus, ogrn, name, regionCode, inn string) *model.ChangeEvent {
	if oldStatus == newStatus {
		return nil
	}

	oldJSON, _ := json.Marshal(oldStatus)
	newJSON, _ := json.Marshal(newStatus)

	change := &model.ChangeEvent{
		ChangeID:      uuid.New().String(),
		EntityType:    "company",
		EntityID:      ogrn,
		EntityName:    name,
		ChangeType:    model.ChangeTypeStatus,
		FieldName:     "status",
		OldValue:      string(oldJSON),
		NewValue:      string(newJSON),
		IsSignificant: c.classifier.IsStatusChangeSignificant(oldStatus, newStatus),
		Description:   fmt.Sprintf("Статус изменен: %s → %s", oldStatus, newStatus),
		RegionCode:    regionCode,
		INN:           inn,
	}

	return change
}

// compareIPStatus сравнивает статусы ИП
func (c *Comparator) compareIPStatus(oldStatus, newStatus, ogrnip, name, regionCode, inn string) *model.ChangeEvent {
	if oldStatus == newStatus {
		return nil
	}

	oldJSON, _ := json.Marshal(oldStatus)
	newJSON, _ := json.Marshal(newStatus)

	change := &model.ChangeEvent{
		ChangeID:      uuid.New().String(),
		EntityType:    "entrepreneur",
		EntityID:      ogrnip,
		EntityName:    name,
		ChangeType:    model.ChangeTypeIPStatus,
		FieldName:     "status",
		OldValue:      string(oldJSON),
		NewValue:      string(newJSON),
		IsSignificant: c.classifier.IsStatusChangeSignificant(oldStatus, newStatus),
		Description:   fmt.Sprintf("Статус ИП изменен: %s → %s", oldStatus, newStatus),
		RegionCode:    regionCode,
		INN:           inn,
	}

	return change
}

// compareDirector сравнивает руководителей
func (c *Comparator) compareDirector(old, new *model.Company) *model.ChangeEvent {
	// Изменение произошло если изменилось ФИО или ИНН
	if old.DirectorFullName == new.DirectorFullName && old.DirectorINN == new.DirectorINN {
		return nil
	}

	oldDirector := map[string]string{
		"full_name": old.DirectorFullName,
		"inn":       old.DirectorINN,
		"position":  old.DirectorPosition,
	}
	newDirector := map[string]string{
		"full_name": new.DirectorFullName,
		"inn":       new.DirectorINN,
		"position":  new.DirectorPosition,
	}

	oldJSON, _ := json.Marshal(oldDirector)
	newJSON, _ := json.Marshal(newDirector)

	change := &model.ChangeEvent{
		ChangeID:      uuid.New().String(),
		EntityType:    "company",
		EntityID:      old.OGRN,
		EntityName:    old.FullName,
		ChangeType:    model.ChangeTypeDirector,
		FieldName:     "director",
		OldValue:      string(oldJSON),
		NewValue:      string(newJSON),
		IsSignificant: true, // Смена руководителя всегда значима
		Description:   fmt.Sprintf("Руководитель изменен: %s → %s", old.DirectorFullName, new.DirectorFullName),
		RegionCode:    old.RegionCode,
		INN:           old.INN,
	}

	return change
}

// compareFounders сравнивает учредителей
func (c *Comparator) compareFounders(old, new *model.Company) []*model.ChangeEvent {
	var changes []*model.ChangeEvent

	// Создаем мапы для быстрого поиска
	oldMap := make(map[string]*model.Founder)
	newMap := make(map[string]*model.Founder)

	for i := range old.Founders {
		key := old.Founders[i].INN
		if key == "" {
			key = old.Founders[i].OGRN
		}
		oldMap[key] = &old.Founders[i]
	}

	for i := range new.Founders {
		key := new.Founders[i].INN
		if key == "" {
			key = new.Founders[i].OGRN
		}
		newMap[key] = &new.Founders[i]
	}

	// Проверяем удаленных учредителей
	for key, founder := range oldMap {
		if _, exists := newMap[key]; !exists {
			founderJSON, _ := json.Marshal(founder)
			change := &model.ChangeEvent{
				ChangeID:      uuid.New().String(),
				EntityType:    "company",
				EntityID:      old.OGRN,
				EntityName:    old.FullName,
				ChangeType:    model.ChangeTypeFounderRemoved,
				FieldName:     "founder_removed",
				OldValue:      string(founderJSON),
				NewValue:      "null",
				IsSignificant: true,
				Description:   fmt.Sprintf("Учредитель удален: %s", founder.FullName),
				RegionCode:    old.RegionCode,
				INN:           old.INN,
			}
			changes = append(changes, change)
		}
	}

	// Проверяем новых учредителей
	for key, founder := range newMap {
		if _, exists := oldMap[key]; !exists {
			founderJSON, _ := json.Marshal(founder)
			change := &model.ChangeEvent{
				ChangeID:      uuid.New().String(),
				EntityType:    "company",
				EntityID:      old.OGRN,
				EntityName:    old.FullName,
				ChangeType:    model.ChangeTypeFounderAdded,
				FieldName:     "founder_added",
				OldValue:      "null",
				NewValue:      string(founderJSON),
				IsSignificant: true,
				Description:   fmt.Sprintf("Учредитель добавлен: %s", founder.FullName),
				RegionCode:    old.RegionCode,
				INN:           old.INN,
			}
			changes = append(changes, change)
		}
	}

	// Проверяем изменение долей
	for key := range oldMap {
		if oldFounder, ok := oldMap[key]; ok {
			if newFounder, ok := newMap[key]; ok {
				if oldFounder.SharePercent != newFounder.SharePercent {
					shareChange := map[string]interface{}{
						"founder":    oldFounder.FullName,
						"old_share":  oldFounder.SharePercent,
						"new_share":  newFounder.SharePercent,
					}
					oldJSON, _ := json.Marshal(map[string]float64{"share_percent": oldFounder.SharePercent})
					newJSON, _ := json.Marshal(map[string]float64{"share_percent": newFounder.SharePercent})

					change := &model.ChangeEvent{
						ChangeID:      uuid.New().String(),
						EntityType:    "company",
						EntityID:      old.OGRN,
						EntityName:    old.FullName,
						ChangeType:    model.ChangeTypeFounderShare,
						FieldName:     "founder_share",
						OldValue:      string(oldJSON),
						NewValue:      string(newJSON),
						IsSignificant: c.classifier.IsShareChangeSignificant(oldFounder.SharePercent, newFounder.SharePercent),
						Description:   fmt.Sprintf("Доля учредителя %s изменена: %.2f%% → %.2f%%", shareChange["founder"], oldFounder.SharePercent, newFounder.SharePercent),
						RegionCode:    old.RegionCode,
						INN:           old.INN,
					}
					changes = append(changes, change)
				}
			}
		}
	}

	return changes
}

// compareAddress сравнивает адреса
func (c *Comparator) compareAddress(old, new *model.Company) *model.ChangeEvent {
	if old.AddressFull == new.AddressFull {
		return nil
	}

	oldAddr := map[string]string{
		"full":        old.AddressFull,
		"postal_code": old.AddressPostalCode,
		"region":      old.AddressRegion,
		"city":        old.AddressCity,
		"street":      old.AddressStreet,
		"house":       old.AddressHouse,
	}
	newAddr := map[string]string{
		"full":        new.AddressFull,
		"postal_code": new.AddressPostalCode,
		"region":      new.AddressRegion,
		"city":        new.AddressCity,
		"street":      new.AddressStreet,
		"house":       new.AddressHouse,
	}

	oldJSON, _ := json.Marshal(oldAddr)
	newJSON, _ := json.Marshal(newAddr)

	change := &model.ChangeEvent{
		ChangeID:      uuid.New().String(),
		EntityType:    "company",
		EntityID:      old.OGRN,
		EntityName:    old.FullName,
		ChangeType:    model.ChangeTypeAddress,
		FieldName:     "address",
		OldValue:      string(oldJSON),
		NewValue:      string(newJSON),
		IsSignificant: c.classifier.IsAddressChangeSignificant(old.AddressFull, new.AddressFull),
		Description:   fmt.Sprintf("Адрес изменен: %s → %s", old.AddressFull, new.AddressFull),
		RegionCode:    old.RegionCode,
		INN:           old.INN,
	}

	return change
}

// compareIPAddress сравнивает адреса ИП
func (c *Comparator) compareIPAddress(old, new *model.Entrepreneur) *model.ChangeEvent {
	if old.AddressFull == new.AddressFull {
		return nil
	}

	oldAddr := map[string]string{
		"full":        old.AddressFull,
		"postal_code": old.AddressPostalCode,
		"region":      old.AddressRegion,
		"city":        old.AddressCity,
		"street":      old.AddressStreet,
		"house":       old.AddressHouse,
	}
	newAddr := map[string]string{
		"full":        new.AddressFull,
		"postal_code": new.AddressPostalCode,
		"region":      new.AddressRegion,
		"city":        new.AddressCity,
		"street":      new.AddressStreet,
		"house":       new.AddressHouse,
	}

	oldJSON, _ := json.Marshal(oldAddr)
	newJSON, _ := json.Marshal(newAddr)

	change := &model.ChangeEvent{
		ChangeID:      uuid.New().String(),
		EntityType:    "entrepreneur",
		EntityID:      old.OGRNIP,
		EntityName:    old.FullName,
		ChangeType:    model.ChangeTypeIPAddress,
		FieldName:     "address",
		OldValue:      string(oldJSON),
		NewValue:      string(newJSON),
		IsSignificant: c.classifier.IsAddressChangeSignificant(old.AddressFull, new.AddressFull),
		Description:   fmt.Sprintf("Адрес ИП изменен: %s → %s", old.AddressFull, new.AddressFull),
		RegionCode:    old.RegionCode,
		INN:           old.INN,
	}

	return change
}

// compareCapital сравнивает уставный капитал
func (c *Comparator) compareCapital(old, new *model.Company) *model.ChangeEvent {
	if old.AuthorizedCapital == new.AuthorizedCapital {
		return nil
	}

	oldJSON, _ := json.Marshal(map[string]interface{}{"amount": old.AuthorizedCapital, "currency": old.CapitalCurrency})
	newJSON, _ := json.Marshal(map[string]interface{}{"amount": new.AuthorizedCapital, "currency": new.CapitalCurrency})

	change := &model.ChangeEvent{
		ChangeID:      uuid.New().String(),
		EntityType:    "company",
		EntityID:      old.OGRN,
		EntityName:    old.FullName,
		ChangeType:    model.ChangeTypeCapital,
		FieldName:     "authorized_capital",
		OldValue:      string(oldJSON),
		NewValue:      string(newJSON),
		IsSignificant: c.classifier.IsCapitalChangeSignificant(old.AuthorizedCapital, new.AuthorizedCapital),
		Description:   fmt.Sprintf("Уставный капитал изменен: %.2f → %.2f", old.AuthorizedCapital, new.AuthorizedCapital),
		RegionCode:    old.RegionCode,
		INN:           old.INN,
	}

	return change
}

// compareActivities сравнивает виды деятельности компании
func (c *Comparator) compareActivities(old, new *model.Company) []*model.ChangeEvent {
	var changes []*model.ChangeEvent

	// Создаем мапы для быстрого поиска
	oldMap := make(map[string]bool)
	newMap := make(map[string]bool)

	for _, okved := range old.AdditionalOKVED {
		oldMap[okved] = true
	}
	for _, okved := range new.AdditionalOKVED {
		newMap[okved] = true
	}

	// Проверяем удаленные ОКВЭД
	for okved := range oldMap {
		if !newMap[okved] {
			change := &model.ChangeEvent{
				ChangeID:      uuid.New().String(),
				EntityType:    "company",
				EntityID:      old.OGRN,
				EntityName:    old.FullName,
				ChangeType:    model.ChangeTypeActivityRemoved,
				FieldName:     "activity_removed",
				OldValue:      okved,
				NewValue:      "",
				IsSignificant: false,
				Description:   fmt.Sprintf("Вид деятельности удален: %s", okved),
				RegionCode:    old.RegionCode,
				INN:           old.INN,
			}
			changes = append(changes, change)
		}
	}

	// Проверяем новые ОКВЭД
	for okved := range newMap {
		if !oldMap[okved] {
			change := &model.ChangeEvent{
				ChangeID:      uuid.New().String(),
				EntityType:    "company",
				EntityID:      old.OGRN,
				EntityName:    old.FullName,
				ChangeType:    model.ChangeTypeActivityAdded,
				FieldName:     "activity_added",
				OldValue:      "",
				NewValue:      okved,
				IsSignificant: false,
				Description:   fmt.Sprintf("Вид деятельности добавлен: %s", okved),
				RegionCode:    old.RegionCode,
				INN:           old.INN,
			}
			changes = append(changes, change)
		}
	}

	return changes
}

// compareIPActivities сравнивает виды деятельности ИП
func (c *Comparator) compareIPActivities(old, new *model.Entrepreneur) []*model.ChangeEvent {
	var changes []*model.ChangeEvent

	// Создаем мапы для быстрого поиска
	oldMap := make(map[string]bool)
	newMap := make(map[string]bool)

	for _, okved := range old.AdditionalOKVED {
		oldMap[okved] = true
	}
	for _, okved := range new.AdditionalOKVED {
		newMap[okved] = true
	}

	// Проверяем удаленные ОКВЭД
	for okved := range oldMap {
		if !newMap[okved] {
			change := &model.ChangeEvent{
				ChangeID:      uuid.New().String(),
				EntityType:    "entrepreneur",
				EntityID:      old.OGRNIP,
				EntityName:    old.FullName,
				ChangeType:    model.ChangeTypeIPActivity,
				FieldName:     "activity_removed",
				OldValue:      okved,
				NewValue:      "",
				IsSignificant: false,
				Description:   fmt.Sprintf("Вид деятельности ИП удален: %s", okved),
				RegionCode:    old.RegionCode,
				INN:           old.INN,
			}
			changes = append(changes, change)
		}
	}

	// Проверяем новые ОКВЭД
	for okved := range newMap {
		if !oldMap[okved] {
			change := &model.ChangeEvent{
				ChangeID:      uuid.New().String(),
				EntityType:    "entrepreneur",
				EntityID:      old.OGRNIP,
				EntityName:    old.FullName,
				ChangeType:    model.ChangeTypeIPActivity,
				FieldName:     "activity_added",
				OldValue:      "",
				NewValue:      okved,
				IsSignificant: false,
				Description:   fmt.Sprintf("Вид деятельности ИП добавлен: %s", okved),
				RegionCode:    old.RegionCode,
				INN:           old.INN,
			}
			changes = append(changes, change)
		}
	}

	return changes
}

// compareLicensesCount сравнивает количество лицензий
func (c *Comparator) compareLicensesCount(old, new *model.Company) *model.ChangeEvent {
	if old.LicensesCount == new.LicensesCount {
		return nil
	}

	oldJSON, _ := json.Marshal(old.LicensesCount)
	newJSON, _ := json.Marshal(new.LicensesCount)

	changeType := model.ChangeTypeLicenseAdded
	if new.LicensesCount < old.LicensesCount {
		changeType = model.ChangeTypeLicenseRevoked
	}

	change := &model.ChangeEvent{
		ChangeID:      uuid.New().String(),
		EntityType:    "company",
		EntityID:      old.OGRN,
		EntityName:    old.FullName,
		ChangeType:    changeType,
		FieldName:     "licenses_count",
		OldValue:      string(oldJSON),
		NewValue:      string(newJSON),
		IsSignificant: true,
		Description:   fmt.Sprintf("Количество лицензий изменено: %d → %d", old.LicensesCount, new.LicensesCount),
		RegionCode:    old.RegionCode,
		INN:           old.INN,
	}

	return change
}

// compareIPLicensesCount сравнивает количество лицензий ИП
func (c *Comparator) compareIPLicensesCount(old, new *model.Entrepreneur) *model.ChangeEvent {
	if old.LicensesCount == new.LicensesCount {
		return nil
	}

	oldJSON, _ := json.Marshal(old.LicensesCount)
	newJSON, _ := json.Marshal(new.LicensesCount)

	change := &model.ChangeEvent{
		ChangeID:      uuid.New().String(),
		EntityType:    "entrepreneur",
		EntityID:      old.OGRNIP,
		EntityName:    old.FullName,
		ChangeType:    model.ChangeTypeLicenseAdded, // Используем общий тип для ИП
		FieldName:     "licenses_count",
		OldValue:      string(oldJSON),
		NewValue:      string(newJSON),
		IsSignificant: true,
		Description:   fmt.Sprintf("Количество лицензий ИП изменено: %d → %d", old.LicensesCount, new.LicensesCount),
		RegionCode:    old.RegionCode,
		INN:           old.INN,
	}

	return change
}

// compareBranchesCount сравнивает количество филиалов
func (c *Comparator) compareBranchesCount(old, new *model.Company) *model.ChangeEvent {
	if old.BranchesCount == new.BranchesCount {
		return nil
	}

	oldJSON, _ := json.Marshal(old.BranchesCount)
	newJSON, _ := json.Marshal(new.BranchesCount)

	changeType := model.ChangeTypeBranchAdded
	if new.BranchesCount < old.BranchesCount {
		changeType = model.ChangeTypeBranchClosed
	}

	change := &model.ChangeEvent{
		ChangeID:      uuid.New().String(),
		EntityType:    "company",
		EntityID:      old.OGRN,
		EntityName:    old.FullName,
		ChangeType:    changeType,
		FieldName:     "branches_count",
		OldValue:      string(oldJSON),
		NewValue:      string(newJSON),
		IsSignificant: false,
		Description:   fmt.Sprintf("Количество филиалов изменено: %d → %d", old.BranchesCount, new.BranchesCount),
		RegionCode:    old.RegionCode,
		INN:           old.INN,
	}

	return change
}
