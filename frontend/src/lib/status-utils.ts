/**
 * Утилиты для работы со статусами компаний и предпринимателей
 */

export interface StatusInfo {
  color: string;
  label: string;
  category: 'active' | 'liquidated' | 'liquidating' | 'suspended' | 'reorganizing' | 'unknown';
}

/**
 * Получает информацию о статусе компании
 * @param status - статус компании из API
 * @returns объект с цветом, меткой и категорией статуса
 */
export function getStatusInfo(status: string): StatusInfo {
  if (!status) {
    return {
      color: "bg-gray-100 text-gray-800 border-gray-200",
      label: "Статус не указан",
      category: 'unknown'
    };
  }
  
  // Отладочная информация в dev режиме
  if (process.env.NODE_ENV === "development") {
    console.log("Processing status:", status);
  }
  
  const normalizedStatus = status.toLowerCase().trim();
  
  // Действующие статусы
  if (isActiveStatus(normalizedStatus)) {
    return {
      color: "bg-green-100 text-green-800 border-green-200",
      label: status,
      category: 'active'
    };
  }
  
  // Ликвидированные статусы
  if (isLiquidatedStatus(normalizedStatus)) {
    return {
      color: "bg-red-100 text-red-800 border-red-200",
      label: status,
      category: 'liquidated'
    };
  }
  
  // В процессе ликвидации
  if (isLiquidatingStatus(normalizedStatus)) {
    return {
      color: "bg-yellow-100 text-yellow-800 border-yellow-200",
      label: status,
      category: 'liquidating'
    };
  }
  
  // Приостановленные статусы
  if (isSuspendedStatus(normalizedStatus)) {
    return {
      color: "bg-orange-100 text-orange-800 border-orange-200",
      label: status,
      category: 'suspended'
    };
  }
  
  // Реорганизация
  if (isReorganizingStatus(normalizedStatus)) {
    return {
      color: "bg-blue-100 text-blue-800 border-blue-200",
      label: status,
      category: 'reorganizing'
    };
  }
  
  // Неизвестный статус
  if (process.env.NODE_ENV === "development") {
    console.warn("Unknown company status:", status);
  }
  
  return {
    color: "bg-gray-100 text-gray-800 border-gray-200",
    label: status || "Неизвестный статус",
    category: 'unknown'
  };
}

/**
 * Получает только CSS классы для цвета статуса (для обратной совместимости)
 */
export function getStatusColor(status: string): string {
  return getStatusInfo(status).color;
}

// Вспомогательные функции для определения категорий статусов

function isActiveStatus(status: string): boolean {
  const activeStatuses = [
    "действующая",
    "active",
    "действует",
    "зарегистрирована",
    "registered",
    "действующее",
    "активная",
    "работает",
    "функционирует"
  ];
  return activeStatuses.includes(status);
}

function isLiquidatedStatus(status: string): boolean {
  const liquidatedStatuses = [
    "ликвидирована",
    "liquidated",
    "ликвидировано",
    "исключена",
    "excluded",
    "прекращена",
    "terminated",
    "ликвидированная",
    "исключено",
    "прекращено",
    "закрыта",
    "закрыто"
  ];
  return liquidatedStatuses.includes(status);
}

function isLiquidatingStatus(status: string): boolean {
  const liquidatingStatuses = [
    "в процессе ликвидации",
    "liquidating",
    "в стадии ликвидации",
    "ликвидируется",
    "процедура ликвидации",
    "ликвидационная процедура"
  ];
  return liquidatingStatuses.includes(status);
}

function isSuspendedStatus(status: string): boolean {
  const suspendedStatuses = [
    "приостановлена",
    "suspended",
    "заморожена",
    "временно приостановлена",
    "приостановлено",
    "заморожено",
    "на паузе"
  ];
  return suspendedStatuses.includes(status);
}

function isReorganizingStatus(status: string): boolean {
  const reorganizingStatuses = [
    "в процессе реорганизации",
    "reorganizing",
    "реорганизуется",
    "реорганизация",
    "преобразование",
    "слияние",
    "присоединение",
    "разделение",
    "выделение"
  ];
  return reorganizingStatuses.includes(status);
}

/**
 * Проверяет, является ли статус активным (компания работает)
 */
export function isCompanyActive(status: string): boolean {
  return getStatusInfo(status).category === 'active';
}

/**
 * Проверяет, является ли статус ликвидированным (компания закрыта)
 */
export function isCompanyLiquidated(status: string): boolean {
  const info = getStatusInfo(status);
  return info.category === 'liquidated' || info.category === 'liquidating';
}

/**
 * Получает человекочитаемое описание статуса
 */
export function getStatusDescription(status: string): string {
  const info = getStatusInfo(status);
  
  switch (info.category) {
    case 'active':
      return "Компания активно ведет деятельность";
    case 'liquidated':
      return "Компания ликвидирована и исключена из реестра";
    case 'liquidating':
      return "Компания находится в процессе ликвидации";
    case 'suspended':
      return "Деятельность компании приостановлена";
    case 'reorganizing':
      return "Компания проходит процедуру реорганизации";
    case 'unknown':
    default:
      return "Статус компании неизвестен или не определен";
  }
}