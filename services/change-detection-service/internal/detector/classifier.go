package detector

import (
	"math"
	"strings"

	"go.uber.org/zap"
)

// Classifier отвечает за классификацию изменений и определение их значимости
type Classifier struct {
	logger *zap.Logger
}

// NewClassifier создает новый экземпляр Classifier
func NewClassifier(logger *zap.Logger) *Classifier {
	return &Classifier{
		logger: logger,
	}
}

// IsStatusChangeSignificant определяет, является ли изменение статуса значимым
func (c *Classifier) IsStatusChangeSignificant(oldStatus, newStatus string) bool {
	// Значимые изменения статуса
	significantTransitions := map[string]map[string]bool{
		"ДЕЙСТВУЮЩАЯ": {
			"ЛИКВИДИРОВАНА":            true,
			"В ПРОЦЕССЕ ЛИКВИДАЦИИ":    true,
			"РЕОРГАНИЗУЕТСЯ":           true,
			"ПРЕКРАЩЕНА":               true,
		},
		"ДЕЙСТВУЮЩИЙ": {
			"ПРЕКРАТИЛ ДЕЯТЕЛЬНОСТЬ":   true,
		},
	}

	if transitions, ok := significantTransitions[oldStatus]; ok {
		if transitions[newStatus] {
			c.logger.Debug("status change is significant",
				zap.String("old_status", oldStatus),
				zap.String("new_status", newStatus),
			)
			return true
		}
	}

	return false
}

// IsShareChangeSignificant определяет, является ли изменение доли учредителя значимым
// Считается значимым если изменение более 5% или переход через порог контроля (25%, 50%)
func (c *Classifier) IsShareChangeSignificant(oldShare, newShare float64) bool {
	diff := math.Abs(newShare - oldShare)

	// Изменение более 5%
	if diff > 5.0 {
		return true
	}

	// Переход через пороги контроля
	thresholds := []float64{25.0, 50.0, 75.0}
	for _, threshold := range thresholds {
		if (oldShare < threshold && newShare >= threshold) ||
			(oldShare >= threshold && newShare < threshold) {
			c.logger.Debug("share change crosses control threshold",
				zap.Float64("old_share", oldShare),
				zap.Float64("new_share", newShare),
				zap.Float64("threshold", threshold),
			)
			return true
		}
	}

	return false
}

// IsAddressChangeSignificant определяет, является ли изменение адреса значимым
// Считается значимым если изменился регион или город
func (c *Classifier) IsAddressChangeSignificant(oldAddress, newAddress string) bool {
	// Простая проверка: если адреса полностью различаются
	if oldAddress == newAddress {
		return false
	}

	// Извлекаем регионы из адресов (упрощенно)
	oldRegion := extractRegion(oldAddress)
	newRegion := extractRegion(newAddress)

	if oldRegion != newRegion && oldRegion != "" && newRegion != "" {
		c.logger.Debug("address change involves region change",
			zap.String("old_region", oldRegion),
			zap.String("new_region", newRegion),
		)
		return true
	}

	// Извлекаем города из адресов
	oldCity := extractCity(oldAddress)
	newCity := extractCity(newAddress)

	if oldCity != newCity && oldCity != "" && newCity != "" {
		c.logger.Debug("address change involves city change",
			zap.String("old_city", oldCity),
			zap.String("new_city", newCity),
		)
		return true
	}

	// Изменение внутри города (улица, дом) не считается значимым
	return false
}

// IsCapitalChangeSignificant определяет, является ли изменение уставного капитала значимым
// Считается значимым если изменение более 50% или более 1 млн руб
func (c *Classifier) IsCapitalChangeSignificant(oldCapital, newCapital float64) bool {
	if oldCapital == 0 {
		// Новый капитал
		return newCapital > 1000000
	}

	diff := math.Abs(newCapital - oldCapital)

	// Изменение более 1 млн руб
	if diff > 1000000 {
		return true
	}

	// Изменение более 50%
	percentChange := (diff / oldCapital) * 100
	if percentChange > 50 {
		c.logger.Debug("capital change is significant",
			zap.Float64("old_capital", oldCapital),
			zap.Float64("new_capital", newCapital),
			zap.Float64("percent_change", percentChange),
		)
		return true
	}

	return false
}

// extractRegion извлекает название региона из адреса (упрощенная версия)
func extractRegion(address string) string {
	address = strings.ToLower(address)

	// Список регионов РФ (частичный, для примера)
	regions := []string{
		"москва", "санкт-петербург", "московская", "ленинградская",
		"новосибирская", "екатеринбург", "свердловская", "краснодарский",
		"ростовская", "нижегородская", "казань", "татарстан",
	}

	for _, region := range regions {
		if strings.Contains(address, region) {
			return region
		}
	}

	return ""
}

// extractCity извлекает название города из адреса (упрощенная версия)
func extractCity(address string) string {
	address = strings.ToLower(address)

	// Ищем паттерны "г. Название", "город Название"
	if idx := strings.Index(address, "г."); idx != -1 {
		cityPart := address[idx+2:]
		cityPart = strings.TrimSpace(cityPart)
		// Берем первое слово после "г."
		parts := strings.Split(cityPart, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if idx := strings.Index(address, "город"); idx != -1 {
		cityPart := address[idx+5:]
		cityPart = strings.TrimSpace(cityPart)
		parts := strings.Split(cityPart, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	return ""
}
