"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Network, ExternalLink, Building, Users, Factory, Globe } from "lucide-react";
import Link from "next/link";
import { useCompanyRelationsQuery } from "@/lib/api/company-hooks";
import { decodeHtmlEntities } from "@/lib/html-utils";
import { formatDate } from "@/lib/format-utils";
import { unifiedStatusOptions } from "@/lib/statuses";
import type { LegalEntity } from "@/lib/api";

interface CompanyRelationsProps {
  company: LegalEntity;
}

// Функция для определения статуса компании (аналогично CompanyStatusBadge)
function getCompanyStatusInfo(company: {
  status?: string;
  statusCode?: string;
  terminationDate?: string;
}) {
  // Если есть дата прекращения деятельности, значит компания закрыта
  if (company.terminationDate) {
    return { text: "Прекратила деятельность", variant: "destructive" as const };
  }

  // Если есть код статуса, используем его
  if (company.statusCode) {
    const statusText = getShortStatusText(company.statusCode);
    return {
      text: statusText,
      variant: getVariantByCode(company.statusCode)
    };
  }

  // По умолчанию считаем действующей
  return { text: "Действующая", variant: "default" as const };
}

// Функция для получения короткого текста статуса
function getShortStatusText(code: string): string {
  switch (code) {
    case "101":
      return "Ликвидируется";
    case "105":
    case "106":
    case "107":
      return "Исключается из реестра";
    case "111":
      return "Уменьшение капитала";
    case "112":
      return "Изменение адреса";
    case "113":
      return "Банкротство возбуждено";
    case "114":
      return "Наблюдение";
    case "115":
      return "Финансовое оздоровление";
    case "116":
      return "Внешнее управление";
    case "117":
      return "Конкурсное производство";
    case "121":
      return "Реорганизация (преобразование)";
    case "122":
      return "Реорганизация (слияние)";
    case "123":
      return "Реорганизация (разделение)";
    case "124":
      return "Реорганизация (присоединение)";
    case "129":
    case "139":
      return "Реорганизация (смешанная)";
    case "131":
      return "Реорганизация (выделение)";
    case "132":
      return "Реорганизация (присоединение к нему)";
    case "134":
      return "Реорганизация (присоединение + выделение)";
    case "136":
      return "Реорганизация (выделение + присоединение)";
    case "701":
    case "702":
      return "Регистрация недействительна";
    case "801":
      return "Запись ошибочна";
    default:
      // Для неизвестных кодов пытаемся найти в unifiedStatusOptions
      const statusOption = unifiedStatusOptions.find(opt => opt.code === code);
      if (statusOption) {
        // Берем первую часть до запятой или весь текст, если запятой нет
        const parts = statusOption.label.split(',');
        return parts[0].trim();
      }
      return "Действующая";
  }
}

// Определяем вариант бейджа по коду статуса (аналогично CompanyStatusBadge)
function getVariantByCode(code: string): "default" | "secondary" | "destructive" | "outline" {
  // Коды ликвидации
  if (code === "101") {
    return "destructive";
  }
  
  // Коды исключения из реестра (недействующие)
  if (code === "105" || code === "106" || code === "107") {
    return "destructive";
  }
  
  // Коды банкротства (113-117)
  if (code === "113" || code === "114" || code === "115" || code === "116" || code === "117") {
    return "destructive";
  }
  
  // Коды реорганизации (121-139)
  if (code.startsWith("12") || code.startsWith("13")) {
    return "secondary";
  }
  
  // Коды недействительности регистрации (701, 702, 801, 802)
  if (code.startsWith("70") || code.startsWith("80")) {
    return "destructive";
  }
  
  // Остальные коды (например, 111 - уменьшение капитала, 112 - изменение места нахождения)
  return "outline";
}

// Функция для получения иконки типа связи
function getRelationshipIcon(type: string) {
  switch (type) {
    case "FOUNDER_COMPANY":
      return <Factory className="h-4 w-4" />;
    case "SUBSIDIARY_COMPANY":
      return <Building className="h-4 w-4" />;
    case "COMMON_FOUNDERS":
      return <Users className="h-4 w-4" />;
    case "COMMON_DIRECTORS":
      return <Users className="h-4 w-4" />;
    case "FOUNDER_TO_DIRECTOR":
      return <Users className="h-4 w-4" />;
    case "DIRECTOR_TO_FOUNDER":
      return <Users className="h-4 w-4" />;
    case "RELATED_BY_PERSON":
      return <Globe className="h-4 w-4" />;
    default:
      return <Building className="h-4 w-4" />;
  }
}

// Функция для получения текста типа связи
function getRelationshipText(type: string) {
  switch (type) {
    case "FOUNDER_COMPANY":
      return "Учредитель";
    case "SUBSIDIARY_COMPANY":
      return "Дочерняя";
    case "COMMON_FOUNDERS":
      return "Общие учредители";
    case "COMMON_DIRECTORS":
      return "Общие руководители";
    case "FOUNDER_TO_DIRECTOR":
      return "Перекрестные связи";
    case "DIRECTOR_TO_FOUNDER":
      return "Перекрестные связи";
    case "RELATED_BY_PERSON":
      return "Связанная";
    default:
      return "Связанная";
  }
}

// Функция для получения цвета бейджа типа связи
function getRelationshipBadgeVariant(type: string) {
  switch (type) {
    case "FOUNDER_COMPANY":
      return "default" as const;
    case "SUBSIDIARY_COMPANY":
      return "secondary" as const;
    case "COMMON_FOUNDERS":
      return "outline" as const;
    case "COMMON_DIRECTORS":
      return "outline" as const;
    case "FOUNDER_TO_DIRECTOR":
      return "destructive" as const;
    case "DIRECTOR_TO_FOUNDER":
      return "destructive" as const;
    case "RELATED_BY_PERSON":
      return "destructive" as const;
    default:
      return "secondary" as const;
  }
}

export function CompanyRelations({ company }: CompanyRelationsProps) {
  // Настройки лимитов для связанных компаний
  const RELATIONS_LIMIT = 50; // Максимальное количество связанных компаний
  
  const { data: relationsData, isLoading: relationsLoading, error } = useCompanyRelationsQuery(company.ogrn, RELATIONS_LIMIT);
  
  const relatedCompanies = relationsData?.company?.relatedCompanies || [];
  const isLimitReached = relatedCompanies.length >= RELATIONS_LIMIT;

  // Сначала группируем все связи по ОГРН компании, затем по типу связи
  const companiesWithMultipleRelations = relatedCompanies.reduce((acc, relation) => {
    const ogrn = relation.company.ogrn;
    if (!acc[ogrn]) {
      acc[ogrn] = {
        company: relation.company,
        relations: []
      };
    }
    acc[ogrn].relations.push(relation);
    return acc;
  }, {} as Record<string, { company: any; relations: typeof relatedCompanies }>);

  // Затем группируем по типу связи для отображения секций
  const groupedCompanies = relatedCompanies.reduce((acc, relation) => {
    const type = relation.relationshipType;
    if (!acc[type]) {
      acc[type] = [];
    }
    acc[type].push(relation);
    return acc;
  }, {} as Record<string, typeof relatedCompanies>);

  return (
    <div className="space-y-6">
      {/* Граф связей */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Network className="h-5 w-5" />
            Граф связей
          </CardTitle>
        </CardHeader>
        <CardContent>
          {/* TODO: Реализовать визуализацию графа связей */}
          <div className="h-64 bg-gray-100 rounded-lg flex items-center justify-center">
            <div className="text-center">
              <Network className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-500 mb-2">Визуализация графа связей</p>
              <p className="text-sm text-gray-400">(в разработке)</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Список связанных компаний */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <span className="flex items-center gap-2">
              <Building className="h-5 w-5" />
              Связанные компании
              {relationsLoading && (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
              )}
            </span>
            <div className="flex items-center gap-2">
              {relatedCompanies.length > 0 && (
                <>
                  <Badge variant="secondary">
                    {Object.keys(companiesWithMultipleRelations).length} компаний
                  </Badge>
                  <Badge variant="outline">
                    {relatedCompanies.length} связей
                  </Badge>
                </>
              )}
              {isLimitReached && (
                <Badge variant="outline" className="text-xs">
                  Показаны первые {RELATIONS_LIMIT}
                </Badge>
              )}
            </div>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {relationsLoading ? (
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <div key={i} className="animate-pulse border rounded-lg p-4">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="h-6 w-20 bg-muted rounded"></div>
                    <div className="h-6 w-16 bg-muted rounded"></div>
                  </div>
                  <div className="h-6 bg-muted rounded w-3/4 mb-1"></div>
                  <div className="h-4 bg-muted rounded w-1/2"></div>
                </div>
              ))}
            </div>
          ) : error ? (
            <div className="text-center py-8">
              <Building className="h-12 w-12 text-red-400 mx-auto mb-4" />
              <p className="text-red-500 mb-2">Ошибка загрузки связанных компаний</p>
              <p className="text-sm text-gray-400">{error.message}</p>
            </div>
          ) : relatedCompanies.length > 0 ? (
            <div className="space-y-6">
              {/* Группируем все связи по ОГРН компании */}
              {(() => {
                // Сначала группируем все связи по ОГРН
                const companiesByOgrn = relatedCompanies.reduce((acc, relation) => {
                  const ogrn = relation.company.ogrn;
                  if (!acc[ogrn]) {
                    acc[ogrn] = {
                      company: relation.company,
                      relations: []
                    };
                  }
                  acc[ogrn].relations.push(relation);
                  return acc;
                }, {} as Record<string, { company: any; relations: typeof relatedCompanies }>);

                // Показываем все уникальные компании
                return (
                  <div>
                    <div className="flex items-center gap-2 mb-3">
                      <Building className="h-4 w-4" />
                      <h3 className="font-semibold text-lg">Все связанные компании</h3>
                      <Badge variant="secondary">
                        {Object.keys(companiesByOgrn).length}
                      </Badge>
                    </div>
                    
                    <div className="space-y-3 ml-6">
                      {Object.entries(companiesByOgrn).map(([ogrn, { company: relatedCompany, relations }]) => {
                        const statusInfo = getCompanyStatusInfo(relatedCompany);
                        const hasMultipleRelations = relations.length > 1;

                        return (
                          <div
                            key={ogrn}
                            className="border rounded-lg p-4 hover:border-gray-300 transition-colors"
                          >
                            <div className="flex items-start justify-between">
                              <div className="flex-1">
                                <div className="flex items-center gap-3 mb-2">
                                  <Badge variant={statusInfo.variant} className="text-xs">
                                    {statusInfo.text}
                                  </Badge>
                                  
                                  {/* Если несколько типов связей, показываем раскрывающийся список */}
                                  {hasMultipleRelations ? (
                                    <details className="text-sm text-gray-500">
                                      <summary className="cursor-pointer hover:text-gray-700">
                                        Множественные связи ({relations.length})
                                      </summary>
                                      <div className="mt-2 space-y-1 pl-4 border-l-2 border-gray-200">
                                        {relations.map((relation, idx) => (
                                          <div key={idx} className="text-xs">
                                            • {relation.description}
                                          </div>
                                        ))}
                                      </div>
                                    </details>
                                  ) : (
                                    <span className="text-sm text-gray-500">
                                      {relations[0].description}
                                    </span>
                                  )}
                                </div>
                                
                                <h4 className="font-semibold text-lg mb-1">
                                  {decodeHtmlEntities(relatedCompany.fullName)}
                                </h4>
                                
                                {relatedCompany.shortName && relatedCompany.shortName !== relatedCompany.fullName && (
                                  <p className="text-gray-600 mb-2">
                                    {decodeHtmlEntities(relatedCompany.shortName)}
                                  </p>
                                )}
                                
                                <div className="grid grid-cols-2 gap-4 text-sm text-gray-600">
                                  <div>
                                    <span className="font-mono">ОГРН: {relatedCompany.ogrn}</span>
                                  </div>
                                  <div>
                                    <span className="font-mono">ИНН: {relatedCompany.inn}</span>
                                  </div>
                                  {relatedCompany.registrationDate && (
                                    <div>
                                      <span>Регистрация: {formatDate(relatedCompany.registrationDate)}</span>
                                    </div>
                                  )}
                                  {relatedCompany.address?.city && (
                                    <div>
                                      <span>Город: {relatedCompany.address.city}</span>
                                    </div>
                                  )}
                                </div>
                              </div>
                              
                              <Link href={`/company/${relatedCompany.ogrn}`}>
                                <Button variant="outline" size="sm" className="flex items-center gap-2">
                                  <ExternalLink className="h-4 w-4" />
                                  Открыть
                                </Button>
                              </Link>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                );
              })()}
            </div>
          ) : (
            <div className="text-center py-8">
              <Building className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-500 mb-2">Связанные компании не найдены</p>
              <p className="text-sm text-gray-400">
                Связанными считаются компании с общими учредителями, 
                компании-учредители и дочерние компании
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}