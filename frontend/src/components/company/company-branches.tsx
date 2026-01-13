"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Building2, MapPin, FileText, Users, ChevronDown, ChevronUp } from "lucide-react";
import { api, type LegalEntity, type Branch } from "@/lib/api";
import { decodeHtmlEntities } from "@/lib/html-utils";
import { useEffect, useState } from "react";

interface CompanyBranchesProps {
  company: LegalEntity;
}

// Функция для получения информации о статусе филиала
function getBranchStatusInfo(_branch: Branch) {
  // Пока все филиалы считаем действующими, так как в API нет поля status
  return {
    label: 'Действующий',
    variant: 'default' as const,
    color: 'text-green-600'
  };
}

// Функция для получения типа филиала
function getBranchTypeInfo(branchType: string) {
  switch (branchType) {
    case 'BRANCH':
      return {
        label: 'Филиал',
        icon: Building2,
        color: 'text-blue-600'
      };
    case 'REPRESENTATIVE':
      return {
        label: 'Представительство',
        icon: Users,
        color: 'text-purple-600'
      };
    default:
      return {
        label: 'Подразделение',
        icon: Building2,
        color: 'text-gray-600'
      };
  }
}

export function CompanyBranches({ company }: CompanyBranchesProps) {
  const [branches, setBranches] = useState<Branch[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAllBranches, setShowAllBranches] = useState(false);

  useEffect(() => {
    const fetchBranches = async () => {
      try {
        setLoading(true);
        const data = await api.getCompanyBranches(company.ogrn);
        setBranches(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Ошибка загрузки филиалов');
      } finally {
        setLoading(false);
      }
    };

    fetchBranches();
  }, [company.ogrn]);

  if (loading) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mb-4"></div>
          <p className="text-sm text-gray-500">Загрузка филиалов...</p>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Building2 className="h-12 w-12 text-red-400 mb-4" />
          <h3 className="text-lg font-semibold text-gray-900 mb-2">Ошибка загрузки</h3>
          <p className="text-sm text-gray-500 text-center max-w-md">{error}</p>
        </CardContent>
      </Card>
    );
  }

  if (branches.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Building2 className="h-12 w-12 text-gray-400 mb-4" />
          <h3 className="text-lg font-semibold text-gray-900 mb-2">Филиалы и представительства не найдены</h3>
          <p className="text-sm text-gray-500 text-center max-w-md">
            У данной компании нет информации о филиалах и представительствах в базе данных, либо компания осуществляет деятельность только через головной офис.
          </p>
        </CardContent>
      </Card>
    );
  }

  const activeBranches = branches; // Пока все филиалы считаем активными
  const branchesCount = branches.filter(b => b.type === 'BRANCH').length;
  const representativesCount = branches.filter(b => b.type === 'REPRESENTATIVE').length;

  // Логика для отображения ограниченного количества филиалов
  const visibleBranches = showAllBranches ? branches : branches.slice(0, 5);

  return (
    <div className="space-y-6">
      {/* Сводная информация */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Сводная информация о филиалах и представительствах
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">
                {activeBranches.length}
              </div>
              <div className="text-sm text-gray-500">Действующих</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">
                {branchesCount}
              </div>
              <div className="text-sm text-gray-500">Филиалов</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-purple-600">
                {representativesCount}
              </div>
              <div className="text-sm text-gray-500">Представительств</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-gray-600">
                {branches.length}
              </div>
              <div className="text-sm text-gray-500">Всего</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Список филиалов */}
      <div className="grid gap-4">
        {visibleBranches.map((branch) => {
          const statusInfo = getBranchStatusInfo(branch);
          const typeInfo = getBranchTypeInfo(branch.type);
          const TypeIcon = typeInfo.icon;
          
          return (
            <Card key={branch.id}>
              <CardHeader className="pb-3">
                <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-2">
                      <TypeIcon className={`h-5 w-5 ${typeInfo.color}`} />
                      <h3 className="font-semibold text-lg break-words">{decodeHtmlEntities(branch.name || 'Без названия')}</h3>
                      <Badge variant={statusInfo.variant} className="shrink-0">
                        {statusInfo.label}
                      </Badge>
                    </div>
                    <div className="flex flex-wrap items-center gap-4 text-sm text-gray-600">
                      <div className="flex items-center gap-1">
                        <span className={`font-medium ${typeInfo.color}`}>
                          {typeInfo.label}
                        </span>
                      </div>
                      {branch.kpp && (
                        <div className="flex items-center gap-1">
                          <FileText className="h-4 w-4" />
                          <span className="font-mono">КПП: {branch.kpp}</span>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </CardHeader>
              
              <CardContent className="pt-0">
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                  {/* Адрес */}
                  {branch.address && (
                    <div>
                      <p className="text-sm text-gray-500 mb-2">Адрес</p>
                      <div className="flex items-start gap-2">
                        <MapPin className="h-4 w-4 text-gray-400 mt-0.5 flex-shrink-0" />
                        <div>
                          <p className="font-medium break-words">
                            {decodeHtmlEntities(branch.address.fullAddress || 'Адрес не указан')}
                          </p>
                          {branch.address.region && branch.address.city && (
                            <p className="text-sm text-gray-500 mt-1">
                              {decodeHtmlEntities(branch.address.region)}, {decodeHtmlEntities(branch.address.city)}
                            </p>
                          )}
                        </div>
                      </div>
                    </div>
                  )}
                  
                  {/* Дополнительная информация */}
                  <div>
                    <p className="text-sm text-gray-500 mb-2">Дополнительная информация</p>
                    <div className="space-y-2">
                      <div className="flex items-center gap-2">
                        <FileText className="h-4 w-4 text-gray-400" />
                        <span className="text-sm">
                          Тип: {typeInfo.label}
                        </span>
                      </div>
                      {branch.kpp && (
                        <div className="flex items-center gap-2">
                          <FileText className="h-4 w-4 text-gray-400" />
                          <span className="text-sm font-mono">
                            КПП: {branch.kpp}
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}

        {/* Кнопка для показа всех филиалов */}
        {branches.length > 5 && (
          <div className="flex justify-center pt-4">
            <Button
              variant="outline"
              onClick={() => setShowAllBranches(!showAllBranches)}
              className="flex items-center gap-2"
            >
              {showAllBranches ? (
                <>
                  <ChevronUp className="h-4 w-4" />
                  Скрыть
                </>
              ) : (
                <>
                  <ChevronDown className="h-4 w-4" />
                  Показать все ({branches.length - 5} еще)
                </>
              )}
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}