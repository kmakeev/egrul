/**
 * Конфигурация для интеграции с картографическими сервисами
 */

export const mapsConfig = {
  yandex: {
    apiKey: process.env.NEXT_PUBLIC_YANDEX_MAPS_API_KEY,
    scriptUrl: 'https://api-maps.yandex.ru/2.1/',
    language: 'ru_RU',
    // Используем версию 2.1 с поддержкой современного геокодирования
    apiVersion: '2.1',
    
    // Проверка корректности настройки API ключа
    isConfigured(): boolean {
      return !!(this.apiKey && this.apiKey !== 'your_yandex_maps_api_key_here');
    },
    
    // Получение URL для загрузки скрипта
    getScriptUrl(): string {
      if (!this.isConfigured()) {
        throw new Error('Яндекс.Карты API ключ не настроен');
      }
      return `${this.scriptUrl}?apikey=${this.apiKey}&lang=${this.language}`;
    }
  },
  
  // Настройки по умолчанию для карт
  defaults: {
    zoom: 16,
    controls: ['zoomControl', 'fullscreenControl'],
    placemarkPreset: 'islands#redDotIcon'
  },
  
  // URL для открытия в внешних сервисах
  externalUrls: {
    yandexMaps: (lat: number, lon: number) => 
      `https://yandex.ru/maps/?ll=${lon},${lat}&z=16&l=map&pt=${lon},${lat},pm2rdm`,
    
    googleMaps: (lat: number, lon: number) => 
      `https://www.google.com/maps?q=${lat},${lon}`,
    
    openStreetMap: (lat: number, lon: number) => 
      `https://www.openstreetmap.org/?mlat=${lat}&mlon=${lon}&zoom=16`
  }
};

/**
 * Проверяет доступность картографических сервисов
 */
export function checkMapsAvailability() {
  const availability = {
    yandex: mapsConfig.yandex.isConfigured(),
    hasAnyProvider: false
  };
  
  availability.hasAnyProvider = availability.yandex;
  
  return availability;
}

/**
 * Получает сообщение об ошибке конфигурации
 */
export function getMapsConfigError(): string | null {
  if (!mapsConfig.yandex.isConfigured()) {
    return 'API ключ Яндекс.Карт не настроен. Добавьте NEXT_PUBLIC_YANDEX_MAPS_API_KEY в .env файл';
  }
  
  return null;
}