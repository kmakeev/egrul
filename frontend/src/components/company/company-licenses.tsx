"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { FileText, Calendar, Building2, CheckCircle, XCircle, Clock } from "lucide-react";
import { formatDate } from "@/lib/format-utils";
import { api, type LegalEntity, type License } from "@/lib/api";
import { useEffect, useState } from "react";

interface CompanyLicensesProps {
  company: LegalEntity;
}

// Функция для получения реального статуса лицензии с учетом дат
function getRealLicenseStatus(license: License): 'active' | 'expired' | 'expiring_soon' | 'suspended' | 'revoked' {
  const now = new Date();
  const endDate = license.endDate ? new Date(license.endDate) : null;
  
  // Если лицензия уже помечена как аннулированная или приостановленная
  if (license.status === 'revoked') return 'revoked';
  if (license.status === 'suspended') return 'suspended';
  if (license.status === 'expired') return 'expired';
  
  // Проверяем даты для активных лицензий
  if (license.status === 'active') {
    if (endDate && endDate < now) {
      return 'expired';
    }
    if (endDate && (endDate.getTime() - now.getTime()) < 30 * 24 * 60 * 60 * 1000) {
      return 'expiring_soon';
    }
    return 'active';
  }
  
  // По умолчанию считаем истекшей, если статус неизвестен
  return 'expired';
}

// Функция для получения статуса лицензии
function getLicenseStatusInfo(license: License) {
  const realStatus = getRealLicenseStatus(license);
  
  switch (realStatus) {
    case 'active':
      return {
        label: 'Действующая',
        variant: 'default' as const,
        icon: CheckCircle
      };
    case 'expiring_soon':
      return {
        label: 'Истекает скоро',
        variant: 'secondary' as const,
        icon: Clock
      };
    case 'expired':
      return {
        label: 'Истекла',
        variant: 'destructive' as const,
        icon: XCircle
      };
    case 'suspended':
      return {
        label: 'Приостановлена',
        variant: 'secondary' as const,
        icon: Clock
      };
    case 'revoked':
      return {
        label: 'Аннулирована',
        variant: 'destructive' as const,
        icon: XCircle
      };
    default:
      return {
        label: 'Неизвестно',
        variant: 'outline' as const,
        icon: Clock
      };
  }
}

export function CompanyLicenses({ company }: CompanyLicensesProps) {
  const [licenses, setLicenses] = useState<License[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchLicenses = async () => {
      try {
        setLoading(true);
        const data = await api.getCompanyLicenses(company.ogrn);
        setLicenses(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Ошибка загрузки лицензий');
      } finally {
        setLoading(false);
      }
    };

    fetchLicenses();
  }, [company.ogrn]);

  if (loading) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mb-4"></div>
          <p className="text-sm text-gray-500">Загрузка лицензий...</p>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <XCircle className="h-12 w-12 text-red-400 mb-4" />
          <h3 className="text-lg font-semibold text-gray-900 mb-2">Ошибка загрузки</h3>
          <p className="text-sm text-gray-500 text-center max-w-md">{error}</p>
        </CardContent>
      </Card>
    );
  }

  if (licenses.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <FileText className="h-12 w-12 text-gray-400 mb-4" />
          <h3 className="text-lg font-semibold text-gray-900 mb-2">Лицензии не найдены</h3>
          <p className="text-sm text-gray-500 text-center max-w-md">
            У данной компании нет информации о лицензиях в базе данных, либо лицензии не требуются для осуществляемых видов деятельности.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Сводная информация */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Сводная информация о лицензиях
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">
                {licenses.filter(l => getRealLicenseStatus(l) === 'active').length}
              </div>
              <div className="text-sm text-gray-500">Действующих</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-yellow-600">
                {licenses.filter(l => getRealLicenseStatus(l) === 'expiring_soon').length}
              </div>
              <div className="text-sm text-gray-500">Истекают скоро</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-red-600">
                {licenses.filter(l => {
                  const status = getRealLicenseStatus(l);
                  return status === 'expired' || status === 'revoked' || status === 'suspended';
                }).length}
              </div>
              <div className="text-sm text-gray-500">Недействительных</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">
                {licenses.length}
              </div>
              <div className="text-sm text-gray-500">Всего</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Список лицензий */}
      <div className="grid gap-4">
        {licenses.map((license) => {
          const statusInfo = getLicenseStatusInfo(license);
          const StatusIcon = statusInfo.icon;
          
          return (
            <Card key={license.id}>
              <CardHeader className="pb-3">
                <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-2">
                      <h3 className="font-semibold text-lg break-words">{license.activity || 'Не указан вид деятельности'}</h3>
                      <Badge variant={statusInfo.variant} className="flex items-center gap-1 shrink-0">
                        <StatusIcon className="h-3 w-3" />
                        {statusInfo.label}
                      </Badge>
                    </div>
                    <div className="flex flex-wrap items-center gap-4 text-sm text-gray-600">
                      <div className="flex items-center gap-1">
                        <FileText className="h-4 w-4" />
                        <span className="font-mono">
                          {license.series ? `${license.series} ` : ''}
                          {license.number}
                        </span>
                      </div>
                      {license.authority && (
                        <div className="flex items-center gap-1">
                          <Building2 className="h-4 w-4" />
                          <span className="break-words">{license.authority}</span>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </CardHeader>
              
              <CardContent className="pt-0">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  {license.startDate && (
                    <div>
                      <p className="text-sm text-gray-500 mb-1">Дата выдачи</p>
                      <div className="flex items-center gap-2">
                        <Calendar className="h-4 w-4 text-gray-400" />
                        <p className="font-medium">{formatDate(license.startDate)}</p>
                      </div>
                    </div>
                  )}
                  
                  {license.endDate && (
                    <div>
                      <p className="text-sm text-gray-500 mb-1">Действительна до</p>
                      <div className="flex items-center gap-2">
                        <Calendar className="h-4 w-4 text-gray-400" />
                        <p className="font-medium">{formatDate(license.endDate)}</p>
                      </div>
                    </div>
                  )}
                </div>
                
                {/* Дополнительная информация */}
                <div className="mt-4 pt-4 border-t border-gray-100">
                  <div className="text-xs text-gray-500">
                    {license.authority && <p>Лицензирующий орган: {license.authority}</p>}
                    {license.endDate && (
                      <p className="mt-1">
                        Срок действия: {
                          (() => {
                            const now = new Date();
                            const endDate = new Date(license.endDate);
                            const diffTime = endDate.getTime() - now.getTime();
                            const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
                            
                            if (diffDays < 0) {
                              return `истекла ${Math.abs(diffDays)} дн. назад`;
                            } else if (diffDays === 0) {
                              return 'истекает сегодня';
                            } else if (diffDays === 1) {
                              return 'истекает завтра';
                            } else if (diffDays <= 30) {
                              return `истекает через ${diffDays} дн.`;
                            } else {
                              return `действительна еще ${diffDays} дн.`;
                            }
                          })()
                        }
                      </p>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
}