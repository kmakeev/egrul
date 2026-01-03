"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { MapPin, Building, User, FileText } from "lucide-react";
import { formatDate } from "@/lib/format-utils";
import { unifiedStatusOptions } from "@/lib/statuses";
import type { IndividualEntrepreneur } from "@/lib/api";
import { EntrepreneurMap } from './entrepreneur-map';

interface EntrepreneurInfoProps {
  entrepreneur: IndividualEntrepreneur;
}

export function EntrepreneurInfo({ entrepreneur }: EntrepreneurInfoProps) {
  // Функция для получения текста статуса
  const getStatusText = () => {
    // Если есть дата прекращения деятельности, значит ИП закрыт
    if (entrepreneur.terminationDate) {
      return "Прекратил деятельность";
    }

    // Если есть код статуса, используем его
    if (entrepreneur.statusCode) {
      const statusOption = unifiedStatusOptions.find(opt => opt.code === entrepreneur.statusCode);
      if (statusOption) {
        // Для ИП показываем вторую часть после запятой (статус ИП)
        const parts = statusOption.label.split(',');
        return parts.length > 1 ? parts[1].trim() : parts[0].trim();
      }
    }

    // По умолчанию считаем действующим
    return "Действующий";
  };

  return (
    <div className="grid gap-6 md:grid-cols-2">
      {/* Адрес */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MapPin className="h-5 w-5" />
            Адрес регистрирующего органа
          </CardTitle>
        </CardHeader>
        <CardContent>
          {entrepreneur.address?.fullAddress ? (
            <div className="space-y-3">
              {/* Полный адрес */}
              <div>
                <p className="text-sm">
                  {entrepreneur.address.fullAddress}
                </p>
              </div>
              
              {/* Код региона если есть */}
              {entrepreneur.address.regionCode && (
                <div className="text-sm">
                  <span className="text-gray-500">Код региона:</span>
                  <span className="ml-2 font-medium">{entrepreneur.address.regionCode}</span>
                </div>
              )}
              
              {/* Карта */}
              <EntrepreneurMap 
                address={entrepreneur.address}
                entrepreneurName={`${entrepreneur.lastName} ${entrepreneur.firstName} ${entrepreneur.middleName || ''}`.trim()}
              />
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">Адрес не указан</p>
          )}
        </CardContent>
      </Card>

      {/* Регистрационные данные */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Регистрационные данные
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 gap-4 text-sm">
            <div>
              <p className="font-medium text-muted-foreground mb-1">ОГРНИП</p>
              <p className="font-mono">{entrepreneur.ogrnip}</p>
            </div>
            <div>
              <p className="font-medium text-muted-foreground mb-1">ИНН</p>
              <p className="font-mono">{entrepreneur.inn}</p>
            </div>
            {entrepreneur.registrationDate && (
              <div>
                <p className="font-medium text-muted-foreground mb-1">Дата регистрации</p>
                <p>{formatDate(entrepreneur.registrationDate)}</p>
              </div>
            )}
            {entrepreneur.ogrnipDate && (
              <div>
                <p className="font-medium text-muted-foreground mb-1">Дата присвоения ОГРНИП</p>
                <p>{formatDate(entrepreneur.ogrnipDate)}</p>
              </div>
            )}
            <div>
              <p className="font-medium text-muted-foreground mb-1">Статус</p>
              <p>{getStatusText()}</p>
            </div>
            {entrepreneur.terminationDate && (
              <div>
                <p className="font-medium text-muted-foreground mb-1">Дата прекращения деятельности</p>
                <p>{formatDate(entrepreneur.terminationDate)}</p>
              </div>
            )}
            {entrepreneur.regAuthority?.name && (
              <div>
                <p className="font-medium text-muted-foreground mb-1">Регистрирующий орган</p>
                <p>{entrepreneur.regAuthority.name}</p>
              </div>
            )}
            {entrepreneur.taxAuthority?.name && (
              <div>
                <p className="font-medium text-muted-foreground mb-1">Налоговый орган</p>
                <p>{entrepreneur.taxAuthority.name}</p>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Персональная информация */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <User className="h-5 w-5" />
            Персональная информация
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 gap-4 text-sm">
            <div>
              <p className="font-medium text-muted-foreground mb-1">Фамилия</p>
              <p>{entrepreneur.lastName}</p>
            </div>
            <div>
              <p className="font-medium text-muted-foreground mb-1">Имя</p>
              <p>{entrepreneur.firstName}</p>
            </div>
            {entrepreneur.middleName && (
              <div>
                <p className="font-medium text-muted-foreground mb-1">Отчество</p>
                <p>{entrepreneur.middleName}</p>
              </div>
            )}
            {entrepreneur.citizenshipCountryName && (
              <div>
                <p className="font-medium text-muted-foreground mb-1">Гражданство</p>
                <p>{entrepreneur.citizenshipCountryName}</p>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Основной вид деятельности */}
      {entrepreneur.mainActivity && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Building className="h-5 w-5" />
              Основной вид деятельности
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div>
                <p className="font-medium text-muted-foreground mb-1">Код ОКВЭД</p>
                <p className="font-mono text-sm">{entrepreneur.mainActivity.code}</p>
              </div>
              <div>
                <p className="font-medium text-muted-foreground mb-1">Наименование</p>
                <p className="text-sm">{entrepreneur.mainActivity.name}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}