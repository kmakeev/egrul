/**
 * Декодирует HTML-сущности в строке
 * @param str - строка с HTML-сущностями
 * @returns декодированная строка
 */
export function decodeHtmlEntities(str: string): string {
  if (typeof str !== 'string') return str;
  
  // Создаем временный элемент для декодирования HTML-сущностей
  if (typeof document !== 'undefined') {
    const textarea = document.createElement('textarea');
    textarea.innerHTML = str;
    return textarea.value;
  }
  
  // Fallback для серверного рендеринга - декодируем основные сущности вручную
  return str
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&amp;/g, '&'); // &amp; должно быть последним
}