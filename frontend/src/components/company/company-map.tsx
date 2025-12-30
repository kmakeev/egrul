import React, { useEffect, useRef, useState, useCallback, useMemo } from 'react';
import { MapPin, ExternalLink } from 'lucide-react';
import { mapsConfig, getMapsConfigError } from '@/lib/maps-config';
import type { Address } from '@/lib/api';

interface CompanyMapProps {
  address: Address;
  companyName?: string;
}

// Типы для Yandex Maps API
interface YMapsAPI {
  ready: (callback: () => void) => void;
  Map: new (container: string | HTMLElement, state: MapState, options?: MapOptions) => YMap;
  Placemark: new (geometry: [number, number], properties?: PlacemarkProperties, options?: PlacemarkOptions) => YPlacemark;
  geocode: (request: string, options?: GeocodeOptions) => Promise<GeocodeResult>;
}

interface MapState {
  center: [number, number];
  zoom: number;
}

interface MapOptions {
  controls?: string[];
}

interface YMap {
  geoObjects: {
    add: (placemark: YPlacemark) => void;
  };
  destroy: () => void;
}

interface YPlacemark {
  // Placemark interface - можно расширить при необходимости
  [key: string]: unknown;
}

interface PlacemarkProperties {
  balloonContent?: string;
  hintContent?: string;
}

interface PlacemarkOptions {
  preset?: string;
}

interface GeocodeOptions {
  results?: number;
}

interface GeocodeResult {
  geoObjects: {
    get: (index: number) => GeoObject | undefined;
  };
}

interface GeoObject {
  geometry: {
    getCoordinates: () => [number, number];
  };
}

declare global {
  interface Window {
    ymaps: YMapsAPI;
  }
}

export const CompanyMap: React.FC<CompanyMapProps> = ({ address, companyName = 'Компания' }) => {
  const mapRef = useRef<HTMLDivElement>(null);
  const [mapInstance, setMapInstance] = useState<YMap | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [coordinates, setCoordinates] = useState<[number, number] | null>(null);

  // Мемоизируем адрес для стабильности зависимостей
  const memoizedAddress = useMemo(() => ({
    fullAddress: address.fullAddress,
    kladrCode: address.kladrCode,
    fiasId: address.fiasId,
    region: address.region,
    city: address.city,
    street: address.street,
    house: address.house,
    building: address.building,
    flat: address.flat,
    postalCode: address.postalCode,
    regionCode: address.regionCode,
    district: address.district,
    locality: address.locality
  }), [
    address.fullAddress,
    address.kladrCode,
    address.fiasId,
    address.region,
    address.city,
    address.street,
    address.house,
    address.building,
    address.flat,
    address.postalCode,
    address.regionCode,
    address.district,
    address.locality
  ]);

  // Функция для очистки адреса от дублированных префиксов
  const cleanAddress = useCallback((address: Address) => {
    const cleanPart = (part: string | undefined) => {
      if (!part) return '';
      
      // Убираем дублированные префиксы
      return part
        .replace(/УЛ\.\s*УЛ\./gi, 'УЛ.')
        .replace(/Д\.\s*Д\./gi, 'Д.')
        .replace(/КВ\.\s*КВ\./gi, 'КВ.')
        .replace(/СТР\.\s*СТР\./gi, 'СТР.')
        .replace(/КОРП\.\s*КОРП\./gi, 'КОРП.')
        .trim();
    };

    return {
      ...address,
      street: cleanPart(address.street),
      house: cleanPart(address.house),
      building: cleanPart(address.building),
      flat: cleanPart(address.flat),
      fullAddress: cleanPart(address.fullAddress)
    };
  }, []);

  // Получить координаты по КЛАДР коду
  const getCoordinatesByKladrCode = (kladrCode: string): [number, number] | null => {
    try {
      if (kladrCode.length < 2) return null;

      const regionCode = kladrCode.substring(0, 2);
      
      // Координаты центров регионов по кодам КЛАДР
      const kladrRegionCoordinates: { [key: string]: [number, number] } = {
        '01': [44.6098, 40.1006], // Адыгея
        '02': [54.7431, 55.9678], // Башкортостан
        '03': [51.8272, 107.5847], // Бурятия
        '04': [50.7114, 86.8928], // Алтай
        '05': [42.9849, 47.5047], // Дагестан
        '77': [55.7558, 37.6176], // Москва
        '78': [59.9311, 30.3609], // Санкт-Петербург
        '69': [56.8587, 35.9176], // Тверская область
        // Добавим еще несколько популярных
        '50': [55.7558, 37.6176], // Московская область
        '47': [60.0386, 30.3141], // Ленинградская область
        '52': [56.2965, 43.9361], // Нижегородская область
        '66': [56.8431, 60.6454], // Свердловская область
        '23': [45.0328, 38.9769], // Краснодарский край
        '61': [47.2357, 39.7015], // Ростовская область
      };

      return kladrRegionCoordinates[regionCode] || null;
    } catch (error) {
      console.error('KLADR code parsing error:', error);
      return null;
    }
  };

  // Получить координаты по названию региона
  const getCoordinatesByRegion = (region: string): [number, number] | null => {
    const regionKey = region.toUpperCase();
    
    const regionCoordinates: { [key: string]: [number, number] } = {
      'МОСКВА': [55.7558, 37.6176],
      'МОСКОВСКАЯ': [55.7558, 37.6176],
      'САНКТ-ПЕТЕРБУРГ': [59.9311, 30.3609],
      'ЛЕНИНГРАДСКАЯ': [59.9311, 30.3609],
      'ТВЕРСКАЯ': [56.8587, 35.9176],
      'НИЖЕГОРОДСКАЯ': [56.2965, 43.9361],
      'СВЕРДЛОВСКАЯ': [56.8431, 60.6454],
      'КРАСНОДАРСКИЙ': [45.0328, 38.9769],
      'РОСТОВСКАЯ': [47.2357, 39.7015],
      'НОВОСИБИРСКАЯ': [55.0084, 82.9357]
    };

    for (const [key, value] of Object.entries(regionCoordinates)) {
      if (regionKey.includes(key)) {
        return value;
      }
    }

    return null;
  };

  // Показать примерное местоположение по КЛАДР коду или региону
  const showApproximateLocationByKladr = useCallback(async () => {
    try {
      console.log('Showing approximate location by KLADR or region');
      
      let coords: [number, number] | null = null;

      // Сначала пробуем определить по КЛАДР коду
      if (memoizedAddress.kladrCode) {
        coords = getCoordinatesByKladrCode(memoizedAddress.kladrCode);
        if (coords) {
          console.log('KLADR code coordinates found:', coords);
        }
      }

      // Если КЛАДР не помог, используем регион
      if (!coords && memoizedAddress.region) {
        coords = getCoordinatesByRegion(memoizedAddress.region);
        if (coords) {
          console.log('Region coordinates found:', coords);
        }
      }

      // Если ничего не нашли, используем центр России
      if (!coords) {
        coords = [55.7558, 37.6176]; // Москва
        console.log('Using default coordinates (Moscow)');
      }

      setCoordinates(coords);
      setError(null);
      setIsLoading(false);
    } catch (err) {
      console.error('Error showing approximate location:', err);
      setError('Не удалось определить местоположение');
      setIsLoading(false);
    }
  }, [memoizedAddress.kladrCode, memoizedAddress.region]);

  // Fallback геокодирование по городу/региону
  const tryFallbackGeocoding = useCallback(async () => {
    try {
      const fallbackParts = [memoizedAddress.region, memoizedAddress.city].filter(Boolean);
      if (fallbackParts.length === 0) {
        setError('Адрес не найден на карте');
        setIsLoading(false);
        return;
      }

      const fallbackQuery = fallbackParts.join(', ');
      console.log('Fallback geocoding query:', fallbackQuery);

      if (window.ymaps) {
        const geocoder = window.ymaps.geocode(fallbackQuery);
        const result = await geocoder;
        const firstGeoObject = result.geoObjects.get(0);

        if (firstGeoObject) {
          const coords = firstGeoObject.geometry.getCoordinates();
          setCoordinates([coords[0], coords[1]]);
          setError(null);
          console.log('Fallback geocoding successful:', coords);
        } else {
          setError('Адрес не найден на карте');
        }
      } else {
        setError('Адрес не найден на карте');
      }
    } catch (err: unknown) {
      console.error('Ошибка fallback геокодирования:', err);
      if (err instanceof Error && err.message === 'scriptError') {
        console.log('All geocoding failed due to API restrictions, showing approximate location');
        await showApproximateLocationByKladr();
      } else {
        setError('Адрес не найден на карте');
        setIsLoading(false);
      }
    }
  }, [memoizedAddress.region, memoizedAddress.city, showApproximateLocationByKladr]);

  // Геокодирование по текстовому адресу (fallback для КЛАДР)
  const geocodeByTextAddress = useCallback(async (textAddress: string) => {
    try {
      console.log('Geocoding by text address:', textAddress);
      
      const geocoder = window.ymaps.geocode(textAddress);
      const result = await geocoder;
      const firstGeoObject = result.geoObjects.get(0);

      if (firstGeoObject) {
        const coords = firstGeoObject.geometry.getCoordinates();
        setCoordinates([coords[0], coords[1]]);
        setError(null);
        console.log('Text address geocoding successful:', coords);
        setIsLoading(false);
      } else {
        console.log('Text address not found, trying city-level search');
        await tryFallbackGeocoding();
      }
    } catch (err: unknown) {
      console.error('Text address geocoding error:', err);
      if (err instanceof Error && err.message === 'scriptError') {
        console.log('Text address geocoding failed due to API restrictions');
        await showApproximateLocationByKladr();
      } else {
        await tryFallbackGeocoding();
      }
    }
  }, [tryFallbackGeocoding, showApproximateLocationByKladr]);

  // Функция для получения координат по адресу
  const geocodeAddress = useCallback(async () => {
    if (!window.ymaps) {
      setError('Яндекс.Карты не загружены');
      setIsLoading(false);
      return;
    }

    try {
      const cleanedAddress = cleanAddress(memoizedAddress);
      
      let geocodeQuery = '';
      
      // Приоритет геокодирования:
      // 1. КЛАДР код (самый точный)
      // 2. Полный адрес
      // 3. Собранный из компонентов адрес
      
      if (memoizedAddress.kladrCode) {
        // Используем КЛАДР код для точного поиска
        geocodeQuery = memoizedAddress.kladrCode;
        console.log('Using KLADR code for geocoding:', geocodeQuery);
      } else if (cleanedAddress.fullAddress) {
        geocodeQuery = cleanedAddress.fullAddress;
        console.log('Using full address for geocoding:', geocodeQuery);
      } else {
        const parts = [
          cleanedAddress.region,
          cleanedAddress.city,
          cleanedAddress.street,
          cleanedAddress.house
        ].filter(Boolean);
        geocodeQuery = parts.join(', ');
        console.log('Using constructed address for geocoding:', geocodeQuery);
      }

      if (!geocodeQuery) {
        setError('Недостаточно данных для определения местоположения');
        setIsLoading(false);
        return;
      }

      try {
        // Для КЛАДР кода используем специальные параметры
        const geocodeOptions = memoizedAddress.kladrCode ? {
          results: 1,
          kind: 'house', // Ищем конкретный дом
          // Для КЛАДР кода указываем, что это код
          provider: 'yandex#map'
        } : {
          results: 1
        };

        const geocoder = window.ymaps.geocode(geocodeQuery, geocodeOptions);
        const timeoutPromise = new Promise((_, reject) => {
          setTimeout(() => reject(new Error('Timeout')), 15000);
        });

        const result = await Promise.race([geocoder, timeoutPromise]) as GeocodeResult;
        const firstGeoObject = result.geoObjects.get(0);

        if (firstGeoObject) {
          const coords = firstGeoObject.geometry.getCoordinates();
          setCoordinates([coords[0], coords[1]]);
          setError(null);
          console.log('Geocoding successful:', coords);
          setIsLoading(false);
        } else {
          // Если КЛАДР код не сработал, попробуем текстовый адрес
          if (memoizedAddress.kladrCode && cleanedAddress.fullAddress) {
            console.log('KLADR code failed, trying full address');
            await geocodeByTextAddress(cleanedAddress.fullAddress);
          } else {
            console.log('Exact address not found, trying city-level search');
            await tryFallbackGeocoding();
          }
        }
      } catch (jsErr: unknown) {
        console.error('Geocoding error:', jsErr);
        if (jsErr instanceof Error && jsErr.message === 'Timeout') {
          setError('Превышено время ожидания ответа от сервиса карт');
          setIsLoading(false);
        } else if (jsErr instanceof Error && jsErr.message === 'scriptError') {
          // API ключ все еще ограничен, показываем примерные координаты
          console.log('API key still restricted, showing approximate location');
          await showApproximateLocationByKladr();
        } else {
          // Если КЛАДР код не сработал, попробуем текстовый адрес
          if (memoizedAddress.kladrCode && cleanedAddress.fullAddress) {
            console.log('KLADR geocoding failed, trying text address');
            await geocodeByTextAddress(cleanedAddress.fullAddress);
          } else {
            await tryFallbackGeocoding();
          }
        }
      }
    } catch (err: unknown) {
      console.error('Ошибка геокодирования:', err);
      setError('Ошибка при поиске адреса');
      setIsLoading(false);
    }
  }, [memoizedAddress, cleanAddress, geocodeByTextAddress, tryFallbackGeocoding, showApproximateLocationByKladr]);

  // Инициализация карты
  const initMap = useCallback(() => {
    if (!mapRef.current || !coordinates || !window.ymaps) return;

    const map = new window.ymaps.Map(mapRef.current, {
      center: coordinates,
      zoom: mapsConfig.defaults.zoom
    }, {
      controls: mapsConfig.defaults.controls
    });

    // Добавляем метку
    const placemark = new window.ymaps.Placemark(coordinates, {
      balloonContent: `
        <div style="padding: 8px;">
          <strong>${companyName}</strong><br/>
          <small>${memoizedAddress.fullAddress || 'Адрес не указан'}</small>
        </div>
      `,
      hintContent: companyName
    }, {
      preset: mapsConfig.defaults.placemarkPreset
    });

    map.geoObjects.add(placemark);
    setMapInstance(map);
  }, [coordinates, companyName, memoizedAddress.fullAddress]);

  // Загрузка Яндекс.Карт
  useEffect(() => {
    const loadYandexMaps = () => {
      const configError = getMapsConfigError();
      if (configError) {
        setError(configError);
        setIsLoading(false);
        return;
      }

      // Отладочная информация об адресе
      console.log('=== Address Debug Info ===');
      console.log('Full address:', memoizedAddress.fullAddress);
      console.log('KLADR code:', memoizedAddress.kladrCode);
      console.log('FIAS ID:', memoizedAddress.fiasId);
      console.log('Region:', memoizedAddress.region);
      console.log('City:', memoizedAddress.city);
      console.log('Street:', memoizedAddress.street);
      console.log('House:', memoizedAddress.house);
      console.log('========================');

      console.log('Loading Yandex Maps with API key:', process.env.NEXT_PUBLIC_YANDEX_MAPS_API_KEY?.substring(0, 8) + '...');

      if (window.ymaps) {
        console.log('Yandex Maps already loaded');
        window.ymaps.ready(() => {
          geocodeAddress();
        });
        return;
      }

      const script = document.createElement('script');
      script.src = mapsConfig.yandex.getScriptUrl();
      script.async = true;
      
      script.onload = () => {
        console.log('Yandex Maps script loaded successfully');
        window.ymaps.ready(() => {
          console.log('Yandex Maps API ready');
          geocodeAddress();
        });
      };
      
      script.onerror = (e) => {
        console.error('Failed to load Yandex Maps script:', e);
        setError('Не удалось загрузить Яндекс.Карты. Проверьте API ключ и его настройки.');
        setIsLoading(false);
      };
      
      document.head.appendChild(script);
    };

    loadYandexMaps();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [geocodeAddress]);

  // Инициализация карты после получения координат
  useEffect(() => {
    if (coordinates && !mapInstance) {
      initMap();
    }
  }, [coordinates, mapInstance, initMap]);

  // Функция для открытия в Яндекс.Картах
  const openInYandexMaps = () => {
    if (coordinates) {
      const url = mapsConfig.externalUrls.yandexMaps(coordinates[0], coordinates[1]);
      window.open(url, '_blank');
    }
  };

  if (isLoading) {
    return (
      <div className="h-64 bg-gray-100 rounded-md flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-2"></div>
          <p className="text-sm text-gray-500">Загрузка карты...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="h-64 bg-gray-100 rounded-md flex items-center justify-center">
        <div className="text-center">
          <MapPin className="h-8 w-8 text-gray-400 mx-auto mb-2" />
          <p className="text-sm text-gray-500">{error}</p>
          {address.fullAddress && (
            <button
              onClick={() => {
                const fullAddress = address.fullAddress;
                if (fullAddress) {
                  window.open(`https://yandex.ru/maps/?text=${encodeURIComponent(fullAddress)}`, '_blank');
                }
              }}
              className="mt-2 text-xs text-blue-600 hover:text-blue-800 flex items-center gap-1 mx-auto"
            >
              <ExternalLink className="h-3 w-3" />
              Открыть в Яндекс.Картах
            </button>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="relative">
      <div ref={mapRef} className="h-64 w-full rounded-md" />
      {coordinates && (
        <button
          onClick={openInYandexMaps}
          className="absolute top-2 right-2 bg-white shadow-md rounded-md p-2 hover:bg-gray-50 transition-colors"
          title="Открыть в Яндекс.Картах"
        >
          <ExternalLink className="h-4 w-4 text-gray-600" />
        </button>
      )}
    </div>
  );
};