"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { User, Users, Building, Globe, Landmark, Banknote } from "lucide-react";
import { formatCurrency, formatPercentage } from "@/lib/format-utils";
import { decodeHtmlEntities } from "@/lib/html-utils";
import type { LegalEntity, Founder } from "@/lib/api";

interface CompanyManagementProps {
  company: LegalEntity;
}

// Функция для получения иконки типа учредителя
function getFounderTypeIcon(type: Founder["type"]) {
  switch (type) {
    case "PERSON":
      return <User className="h-4 w-4" />;
    case "RUSSIAN_COMPANY":
      return <Building className="h-4 w-4" />;
    case "FOREIGN_COMPANY":
      return <Globe className="h-4 w-4" />;
    case "PUBLIC_ENTITY":
      return <Landmark className="h-4 w-4" />;
    case "FUND":
      return <Building className="h-4 w-4" />;
    default:
      return <User className="h-4 w-4" />;
  }
}

// Функция для получения текста типа учредителя
function getFounderTypeText(type: Founder["type"]) {
  switch (type) {
    case "PERSON":
      return "ФЛ";
    case "RUSSIAN_COMPANY":
      return "ЮЛ (РФ)";
    case "FOREIGN_COMPANY":
      return "ЮЛ (ИН)";
    case "PUBLIC_ENTITY":
      return "ПОО";
    case "FUND":
      return "Фонд";
    default:
      return "ФЛ";
  }
}

// Функция для получения цвета бейджа типа учредителя
function getFounderTypeBadgeVariant(type: Founder["type"]) {
  switch (type) {
    case "PERSON":
      return "secondary" as const;
    case "RUSSIAN_COMPANY":
      return "default" as const;
    case "FOREIGN_COMPANY":
      return "outline" as const;
    case "PUBLIC_ENTITY":
      return "destructive" as const;
    case "FUND":
      return "secondary" as const;
    default:
      return "secondary" as const;
  }
}

// Функция для формирования полного имени учредителя
function getFounderFullName(founder: Founder): string {
  if (founder.type === "PERSON" && founder.lastName && founder.firstName) {
    const parts = [founder.lastName, founder.firstName];
    if (founder.middleName) {
      parts.push(founder.middleName);
    }
    return parts.join(" ");
  }
  return decodeHtmlEntities(founder.name);
}

export function CompanyManagement({ company }: CompanyManagementProps) {
  // Используем учредителей из основного запроса компании
  const founders = company.founders || [];

  // Отладочная информация
  console.log("CompanyManagement Debug:", {
    ogrn: company.ogrn,
    foundersFromCompany: company.founders,
    finalFounders: founders,
    foundersCount: founders.length
  });

  return (
    <div className="space-y-6">
      {/* Руководитель */}
      {(company.director || company.head) && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <User className="h-5 w-5" />
              Руководитель
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="border rounded-lg p-4">
                {/* ФИО */}
                <h3 className="font-semibold text-lg mb-3">
                  {`${(company.director || company.head)?.lastName || ""} ${(company.director || company.head)?.firstName || ""} ${(company.director || company.head)?.middleName || ""}`.trim()}
                </h3>
                
                {/* Реквизиты */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
                  {/* ФИО в одном столбце */}
                  {((company.director || company.head)?.lastName || (company.director || company.head)?.firstName || (company.director || company.head)?.middleName) && (
                    <div className="flex flex-col gap-1">
                      {(company.director || company.head)?.lastName && (
                        <div className="flex items-start gap-2">
                          <span className="text-muted-foreground whitespace-nowrap">Фамилия:</span>
                          <span className="font-medium">{(company.director || company.head)!.lastName}</span>
                        </div>
                      )}
                      {(company.director || company.head)?.firstName && (
                        <div className="flex items-start gap-2">
                          <span className="text-muted-foreground whitespace-nowrap">Имя:</span>
                          <span className="font-medium">{(company.director || company.head)!.firstName}</span>
                        </div>
                      )}
                      {(company.director || company.head)?.middleName && (
                        <div className="flex items-start gap-2">
                          <span className="text-muted-foreground whitespace-nowrap">Отчество:</span>
                          <span className="font-medium">{(company.director || company.head)!.middleName}</span>
                        </div>
                      )}
                    </div>
                  )}
                  
                  {/* Остальные поля */}
                  <div className="flex flex-col gap-1">
                    {(company.director || company.head)?.position && (
                      <div className="flex items-start gap-2">
                        <span className="text-muted-foreground whitespace-nowrap">Должность:</span>
                        <span className="font-medium break-words">{decodeHtmlEntities((company.director || company.head)!.position!)}</span>
                      </div>
                    )}
                    
                    {(company.director || company.head)?.inn && (
                      <div className="flex items-start gap-2">
                        <span className="text-muted-foreground whitespace-nowrap">ИНН:</span>
                        <span className="font-mono font-medium">{(company.director || company.head)!.inn}</span>
                      </div>
                    )}
                    
                    {(company.director || company.head)?.positionCode && (
                      <div className="flex items-start gap-2">
                        <span className="text-muted-foreground whitespace-nowrap">Код должности:</span>
                        <span className="font-mono text-xs">{(company.director || company.head)!.positionCode}</span>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Учредители */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              Учредители
            </div>
            {founders.length > 0 && (
              <Badge variant="secondary" className="text-sm">
                {founders.length} {founders.length === 1 ? 'учредитель' : founders.length < 5 ? 'учредителя' : 'учредителей'}
              </Badge>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {founders.length > 0 ? (
            <div className="space-y-4">
              {founders.map((founder, index) => (
                <div key={index} className="border rounded-lg p-4 transition-colors">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      {/* Тип учредителя и страна */}
                      <div className="flex items-center gap-2 mb-3 flex-wrap">
                        <Badge variant={getFounderTypeBadgeVariant(founder.type)} className="flex items-center gap-1">
                          {getFounderTypeIcon(founder.type)}
                          {getFounderTypeText(founder.type)}
                        </Badge>
                        {founder.country && founder.country.trim() !== "" && (
                          <Badge variant="outline" className="text-xs">
                            <Globe className="h-3 w-3 mr-1" />
                            {founder.country}
                          </Badge>
                        )}
                        {founder.citizenship && (
                          <Badge variant="outline" className="text-xs">
                            {founder.citizenship}
                          </Badge>
                        )}
                      </div>
                      
                      {/* Имя учредителя */}
                      <h3 className="font-semibold text-lg mb-3 break-words">
                        {getFounderFullName(founder)}
                      </h3>
                      
                      {/* Реквизиты */}
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
                        {/* ФИО в одном столбце для физических лиц */}
                        {founder.type === "PERSON" && (founder.lastName || founder.firstName || founder.middleName) && (
                          <div className="flex flex-col gap-1">
                            {founder.lastName && (
                              <div className="flex items-start gap-2">
                                <span className="text-muted-foreground whitespace-nowrap">Фамилия:</span>
                                <span className="font-medium break-words">{founder.lastName}</span>
                              </div>
                            )}
                            {founder.firstName && (
                              <div className="flex items-start gap-2">
                                <span className="text-muted-foreground whitespace-nowrap">Имя:</span>
                                <span className="font-medium break-words">{founder.firstName}</span>
                              </div>
                            )}
                            {founder.middleName && (
                              <div className="flex items-start gap-2">
                                <span className="text-muted-foreground whitespace-nowrap">Отчество:</span>
                                <span className="font-medium break-words">{founder.middleName}</span>
                              </div>
                            )}
                          </div>
                        )}
                        
                        {/* Остальные реквизиты */}
                        <div className="flex flex-col gap-1">
                          {founder.inn && founder.inn.trim() !== "" && (
                            <div className="flex items-start gap-2">
                              <span className="text-muted-foreground whitespace-nowrap">ИНН:</span>
                              <span className="font-mono font-medium break-all">{founder.inn}</span>
                            </div>
                          )}
                          
                          {founder.ogrn && founder.ogrn.trim() !== "" && (
                            <div className="flex items-start gap-2">
                              <span className="text-muted-foreground whitespace-nowrap">ОГРН:</span>
                              <span className="font-mono font-medium break-all">{founder.ogrn}</span>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                    
                    {/* Доля владения */}
                    {(founder.sharePercent !== undefined && founder.sharePercent !== null && founder.sharePercent > 0) || 
                     (founder.shareNominalValue !== undefined && founder.shareNominalValue !== null && founder.shareNominalValue > 0) ? (
                      <div className="text-right flex-shrink-0">
                        {founder.sharePercent !== undefined && founder.sharePercent !== null && founder.sharePercent > 0 && (
                          <div className="text-lg font-bold text-blue-600 dark:text-blue-400 mb-1">
                            {formatPercentage(founder.sharePercent)}
                          </div>
                        )}
                        {founder.shareNominalValue !== undefined && founder.shareNominalValue !== null && founder.shareNominalValue > 0 && (
                          <div className="text-sm text-muted-foreground whitespace-nowrap">
                            {formatCurrency(founder.shareNominalValue, "RUB")}
                          </div>
                        )}
                      </div>
                    ) : null}
                  </div>
                </div>
              ))}
              
              {/* Итоговая информация */}
              {founders.some(f => f.sharePercent && f.sharePercent > 0) && (
                <div className="mt-4 pt-4 border-t">
                  <div className="flex justify-between items-center text-sm">
                    <span className="text-muted-foreground">Общая доля учредителей:</span>
                    <span className="font-semibold">
                      {formatPercentage(
                        founders.reduce((sum, f) => sum + (f.sharePercent || 0), 0)
                      )}
                    </span>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="text-center py-8">
              <Users className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-muted-foreground">Информация об учредителях отсутствует</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Доля компании в уставном капитале */}
      {company.companyShare && (company.companyShare.percent || company.companyShare.nominalValue) && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Banknote className="h-5 w-5" />
              Доля в уставном капитале, принадлежащая обществу
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="border rounded-lg p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground mb-1">
                    Доля, принадлежащая самому обществу
                  </p>
                  <p className="text-sm text-muted-foreground">
                    (не распределена между участниками)
                  </p>
                </div>
                <div className="text-right">
                  {company.companyShare.percent !== undefined && company.companyShare.percent !== null && company.companyShare.percent > 0 && (
                    <div className="text-lg font-bold text-orange-600 dark:text-orange-400 mb-1">
                      {formatPercentage(company.companyShare.percent)}
                    </div>
                  )}
                  {company.companyShare.nominalValue !== undefined && company.companyShare.nominalValue !== null && company.companyShare.nominalValue > 0 && (
                    <div className="text-sm text-muted-foreground">
                      {formatCurrency(company.companyShare.nominalValue, "RUB")}
                    </div>
                  )}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}