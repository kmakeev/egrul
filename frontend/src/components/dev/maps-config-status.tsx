import React from 'react';
import { AlertCircle, CheckCircle, Settings } from 'lucide-react';
import { checkMapsAvailability, getMapsConfigError } from '@/lib/maps-config';

/**
 * Компонент для отображения статуса конфигурации карт
 * Показывается только в режиме разработки
 */
export const MapsConfigStatus: React.FC = () => {
  // Показываем только в development режиме
  if (process.env.NODE_ENV !== 'development') {
    return null;
  }

  const availability = checkMapsAvailability();
  const configError = getMapsConfigError();

  return (
    <div className="fixed bottom-4 right-4 z-50 max-w-sm">
      <div className={`p-3 rounded-lg shadow-lg border ${
        availability.hasAnyProvider 
          ? 'bg-green-50 border-green-200' 
          : 'bg-yellow-50 border-yellow-200'
      }`}>
        <div className="flex items-start gap-2">
          {availability.hasAnyProvider ? (
            <CheckCircle className="h-5 w-5 text-green-600 mt-0.5" />
          ) : (
            <AlertCircle className="h-5 w-5 text-yellow-600 mt-0.5" />
          )}
          
          <div className="flex-1">
            <h4 className="text-sm font-medium text-gray-900 mb-1">
              Статус карт
            </h4>
            
            <div className="space-y-1 text-xs">
              <div className="flex items-center gap-2">
                <span className={`w-2 h-2 rounded-full ${
                  availability.yandex ? 'bg-green-500' : 'bg-red-500'
                }`} />
                <span>Яндекс.Карты: {availability.yandex ? 'OK' : 'Не настроено'}</span>
              </div>
            </div>
            
            {configError && (
              <div className="mt-2 p-2 bg-yellow-100 rounded text-xs text-yellow-800">
                <div className="flex items-start gap-1">
                  <Settings className="h-3 w-3 mt-0.5 flex-shrink-0" />
                  <span>{configError}</span>
                </div>
              </div>
            )}
            
            {!availability.hasAnyProvider && (
              <div className="mt-2 text-xs text-gray-600">
                Добавьте API ключ в <code>.env.local</code>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};