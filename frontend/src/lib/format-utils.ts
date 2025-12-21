/**
 * Утилиты для форматирования данных
 */

/**
 * Форматирует сумму в валюте с учетом локализации
 * @param amount - сумма для форматирования
 * @param currency - код валюты (может быть русским названием или ISO кодом)
 * @param locale - локаль для форматирования (по умолчанию ru-RU)
 * @returns отформатированная строка с валютой
 */
export function formatCurrency(
  amount: number, 
  currency: string = "RUB", 
  locale: string = "ru-RU"
): string {
  // Преобразуем русские названия валют в ISO коды
  const currencyMap: Record<string, string> = {
    "рубль": "RUB",
    "рублей": "RUB", 
    "руб": "RUB",
    "доллар": "USD",
    "долларов": "USD",
    "евро": "EUR",
    "RUB": "RUB",
    "USD": "USD", 
    "EUR": "EUR"
  };
  
  const normalizedCurrency = currencyMap[currency.toLowerCase()] || "RUB";
  
  try {
    return new Intl.NumberFormat(locale, {
      style: "currency",
      currency: normalizedCurrency,
    }).format(amount);
  } catch {
    // Fallback если валюта не поддерживается
    console.warn(`Unsupported currency: ${currency}, falling back to RUB`);
    return new Intl.NumberFormat(locale, {
      style: "currency",
      currency: "RUB",
    }).format(amount);
  }
}

/**
 * Форматирует процентное значение
 * @param value - значение для форматирования
 * @param decimals - количество знаков после запятой (по умолчанию 2)
 * @returns отформатированная строка с процентами
 */
export function formatPercentage(value: number, decimals: number = 2): string {
  return `${value.toFixed(decimals)}%`;
}

/**
 * Форматирует дату в русском формате
 * @param dateString - строка с датой
 * @param options - опции форматирования
 * @returns отформатированная дата
 */
export function formatDate(
  dateString: string, 
  options: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "long", 
    day: "numeric"
  }
): string {
  try {
    return new Date(dateString).toLocaleDateString("ru-RU", options);
  } catch {
    console.warn(`Invalid date: ${dateString}`);
    return dateString;
  }
}

/**
 * Форматирует большие числа с сокращениями (К, М, Б)
 * @param num - число для форматирования
 * @param decimals - количество знаков после запятой
 * @returns отформатированная строка
 */
export function formatLargeNumber(num: number, decimals: number = 1): string {
  if (num === 0) return "0";
  
  const k = 1000;
  const sizes = ["", "К", "М", "Б", "Т"];
  const i = Math.floor(Math.log(Math.abs(num)) / Math.log(k));
  
  if (i === 0) return num.toString();
  
  return (num / Math.pow(k, i)).toFixed(decimals) + sizes[i];
}

/**
 * Форматирует номер телефона в российском формате
 * @param phone - номер телефона
 * @returns отформатированный номер
 */
export function formatPhone(phone: string): string {
  const cleaned = phone.replace(/\D/g, "");
  
  if (cleaned.length === 11 && cleaned.startsWith("7")) {
    return `+7 (${cleaned.slice(1, 4)}) ${cleaned.slice(4, 7)}-${cleaned.slice(7, 9)}-${cleaned.slice(9)}`;
  }
  
  if (cleaned.length === 10) {
    return `+7 (${cleaned.slice(0, 3)}) ${cleaned.slice(3, 6)}-${cleaned.slice(6, 8)}-${cleaned.slice(8)}`;
  }
  
  return phone;
}

/**
 * Форматирует ИНН с разделителями для лучшей читаемости
 * @param inn - ИНН
 * @returns отформатированный ИНН
 */
export function formatINN(inn: string): string {
  if (inn.length === 10) {
    // ИНН юридического лица
    return `${inn.slice(0, 2)} ${inn.slice(2, 4)} ${inn.slice(4, 10)}`;
  }
  
  if (inn.length === 12) {
    // ИНН физического лица
    return `${inn.slice(0, 2)} ${inn.slice(2, 4)} ${inn.slice(4, 10)} ${inn.slice(10)}`;
  }
  
  return inn;
}

/**
 * Форматирует ОГРН с разделителями
 * @param ogrn - ОГРН
 * @returns отформатированный ОГРН
 */
export function formatOGRN(ogrn: string): string {
  if (ogrn.length === 13) {
    // ОГРН юридического лица
    return `${ogrn.slice(0, 1)} ${ogrn.slice(1, 3)} ${ogrn.slice(3, 5)} ${ogrn.slice(5, 10)} ${ogrn.slice(10)}`;
  }
  
  if (ogrn.length === 15) {
    // ОГРНИП
    return `${ogrn.slice(0, 3)} ${ogrn.slice(3, 5)} ${ogrn.slice(5, 7)} ${ogrn.slice(7, 12)} ${ogrn.slice(12)}`;
  }
  
  return ogrn;
}