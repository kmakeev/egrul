"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { MapPin, Calendar, Building, Banknote, Mail, History } from "lucide-react";
import { formatCurrency, formatDate } from "@/lib/format-utils";
import { unifiedStatusOptions } from "@/lib/statuses";
import type { LegalEntity } from "@/lib/api";
import { CompanyMap } from './company-map';

interface CompanyInfoProps {
  company: LegalEntity;
}

export function CompanyInfo({ company }: CompanyInfoProps) {
  // Функция для получения текста статуса (аналогично логике в бейдже)
  const getStatusText = () => {
    // Если есть дата прекращения деятельности, значит компания закрыта
    if (company.terminationDate) {
      return "Прекратила деятельность";
    }

    // Если есть код статуса, используем его
    if (company.statusCode) {
      const statusOption = unifiedStatusOptions.find(opt => opt.code === company.statusCode);
      if (statusOption) {
        // Для ЮЛ показываем только первую часть до запятой (статус ЮЛ)
        return statusOption.label.split(',')[0].trim();
      }
    }

    // По умолчанию считаем действующей
    return "Действующая";
  };

  return (
    <div className="grid gap-6 md:grid-cols-2">
      {/* Адрес */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MapPin className="h-5 w-5" />
            Место нахождения и адрес юридического лица
          </CardTitle>
        </CardHeader>
        <CardContent>
          {company.address ? (
            <div className="space-y-3">
              {/* Полный адрес */}
              {company.address.fullAddress && (
                <div>
                  <p className="text-sm font-medium">
                    {company.address.fullAddress}
                  </p>
                </div>
              )}
              
              {/* Структурированные поля адреса */}
              <div className="grid grid-cols-2 gap-4 text-sm">
                {company.address.postalCode && (
                  <div>
                    <span className="text-gray-500">Индекс:</span>
                    <span className="ml-2 font-medium">{company.address.postalCode}</span>
                  </div>
                )}
                
                {company.address.regionCode && (
                  <div>
                    <span className="text-gray-500">Код региона:</span>
                    <span className="ml-2 font-medium">{company.address.regionCode}</span>
                  </div>
                )}
                
                {company.address.region && (
                  <div>
                    <span className="text-gray-500">Регион:</span>
                    <span className="ml-2 font-medium">{company.address.region}</span>
                  </div>
                )}
                
                {company.address.district && (
                  <div>
                    <span className="text-gray-500">Район:</span>
                    <span className="ml-2 font-medium">{company.address.district}</span>
                  </div>
                )}
                
                {company.address.city && (
                  <div>
                    <span className="text-gray-500">Город:</span>
                    <span className="ml-2 font-medium">{company.address.city}</span>
                  </div>
                )}
                
                {company.address.locality && (
                  <div>
                    <span className="text-gray-500">Населенный пункт:</span>
                    <span className="ml-2 font-medium">{company.address.locality}</span>
                  </div>
                )}
                
                {company.address.street && (
                  <div>
                    <span className="text-gray-500">Улица:</span>
                    <span className="ml-2 font-medium">{company.address.street}</span>
                  </div>
                )}
                
                {company.address.house && (
                  <div>
                    <span className="text-gray-500">Дом:</span>
                    <span className="ml-2 font-medium">{company.address.house}</span>
                  </div>
                )}
                
                {company.address.building && (
                  <div>
                    <span className="text-gray-500">Корпус:</span>
                    <span className="ml-2 font-medium">{company.address.building}</span>
                  </div>
                )}
                
                {company.address.flat && (
                  <div>
                    <span className="text-gray-500">Квартира/Офис:</span>
                    <span className="ml-2 font-medium">{company.address.flat}</span>
                  </div>
                )}
                
                {company.address.fiasId && (
                  <div>
                    <span className="text-gray-500">ФИАС ID:</span>
                    <span className="ml-2 font-mono text-xs">{company.address.fiasId}</span>
                  </div>
                )}
                
                {company.address.kladrCode && (
                  <div>
                    <span className="text-gray-500">Код КЛАДР:</span>
                    <span className="ml-2 font-mono text-xs">{company.address.kladrCode}</span>
                  </div>
                )}
              </div>
              
            {/* Карта */}
            <CompanyMap 
              address={company.address}
              companyName={company.fullName || company.shortName || 'Компания'}
            />
            </div>
          ) : (
            <p className="text-sm text-gray-500">Адрес не указан</p>
          )}
        </CardContent>
      </Card>

      {/* Регистрационные данные */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Calendar className="h-5 w-5" />
            Регистрационные данные
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {company.registrationDate && (
            <div>
              <p className="text-sm text-gray-500">Дата регистрации</p>
              <p className="font-medium">{formatDate(company.registrationDate)}</p>
            </div>
          )}
          
          {company.registrationAuthority && (
            <div>
              <p className="text-sm text-gray-500">Регистрирующий орган</p>
              <p className="font-medium">{company.registrationAuthority}</p>
            </div>
          )}
          
          <div>
            <p className="text-sm text-gray-500">Статус</p>
            <p className="font-medium">{getStatusText()}</p>
          </div>

          {/* Регистрация до 01.07.2002 */}
          {company.oldRegistration && (company.oldRegistration.regNumber || company.oldRegistration.regDate || company.oldRegistration.authority) && (
            <div className="mt-4 pt-4 border-t">
              <div className="flex items-center gap-2 mb-2">
                <History className="h-4 w-4 text-gray-500" />
                <p className="text-sm font-medium text-gray-700">Регистрация до 01.07.2002</p>
              </div>
              <div className="space-y-2 text-sm">
                {company.oldRegistration.regNumber && (
                  <div>
                    <span className="text-gray-500">Рег. номер:</span>
                    <span className="ml-2 font-medium">{company.oldRegistration.regNumber}</span>
                  </div>
                )}
                {company.oldRegistration.regDate && (
                  <div>
                    <span className="text-gray-500">Дата регистрации:</span>
                    <span className="ml-2 font-medium">{formatDate(company.oldRegistration.regDate)}</span>
                  </div>
                )}
                {company.oldRegistration.authority && (
                  <div>
                    <span className="text-gray-500">Орган регистрации:</span>
                    <span className="ml-2 font-medium">{company.oldRegistration.authority}</span>
                  </div>
                )}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Email */}
      {company.email && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Mail className="h-5 w-5" />
              Контактная информация
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div>
              <p className="text-sm text-gray-500">Адрес электронной почты</p>
              <a href={`mailto:${company.email}`} className="font-medium text-blue-600 hover:underline">
                {company.email}
              </a>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Уставный капитал */}
      {company.capital && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Banknote className="h-5 w-5" />
              Уставный капитал
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {formatCurrency(company.capital.amount, company.capital.currency)}
            </p>
          </CardContent>
        </Card>
      )}

      {/* Основной ОКВЭД */}
      {company.mainActivity && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Building className="h-5 w-5" />
              Основной вид деятельности
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <p className="font-mono text-sm font-semibold">
                {company.mainActivity.code}
              </p>
              <p className="text-sm">{company.mainActivity.name}</p>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}