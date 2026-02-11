"use client";

import { useMemo } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Building2, TrendingUp, Users, TrendingDown } from "lucide-react";
import { cn } from "@/lib/utils";
import { useRegistrationsByMonth } from "@/lib/api/dashboard-hooks";
import type { Statistics, StatsFilter } from "@/lib/api/dashboard-hooks";
import type { EntityType } from "./use-analytics-filters";

interface KPICardsProps {
  statistics?: Statistics;
  dateFrom?: Date;
  dateTo?: Date;
  entityType?: EntityType;
  filter?: StatsFilter;
  isLoading?: boolean;
}

interface KPICard {
  title: string;
  value?: number;
  icon: React.ReactNode;
  colorClass: string;
  subtitle?: string;
}

/**
 * Компонент KPI карточек для дашборда
 * Отображает ключевые метрики в 2 ряда: компании и ИП
 */
export function KPICards({
  statistics,
  dateFrom,
  dateTo,
  entityType,
  filter,
  isLoading
}: KPICardsProps) {
  // Форматируем даты для GraphQL запросов
  const dateFromStr = dateFrom?.toISOString().split("T")[0];
  const dateToStr = dateTo?.toISOString().split("T")[0];

  // Получаем данные по компаниям (отдельный запрос)
  const { data: companyData, isLoading: isLoadingCompanies } = useRegistrationsByMonth(
    dateFromStr,
    dateToStr,
    "COMPANY",
    filter,
    { enabled: !isLoading }
  );

  // Получаем данные по ИП (отдельный запрос)
  const { data: entrepreneurData, isLoading: isLoadingEntrepreneurs } = useRegistrationsByMonth(
    dateFromStr,
    dateToStr,
    "ENTREPRENEUR",
    filter,
    { enabled: !isLoading }
  );

  // Вычисляем регистрации за выбранный период
  const companyRegistrations = useMemo(() => {
    if (!companyData || companyData.length === 0) return 0;
    return companyData.reduce((sum, point) => sum + point.registrationsCount, 0);
  }, [companyData]);

  const entrepreneurRegistrations = useMemo(() => {
    if (!entrepreneurData || entrepreneurData.length === 0) return 0;
    return entrepreneurData.reduce((sum, point) => sum + point.registrationsCount, 0);
  }, [entrepreneurData]);

  // Вычисляем ликвидации за выбранный период
  const companyTerminations = useMemo(() => {
    if (!companyData || companyData.length === 0) return 0;
    return companyData.reduce((sum, point) => sum + point.terminationsCount, 0);
  }, [companyData]);

  const entrepreneurTerminations = useMemo(() => {
    if (!entrepreneurData || entrepreneurData.length === 0) return 0;
    return entrepreneurData.reduce((sum, point) => sum + point.terminationsCount, 0);
  }, [entrepreneurData]);

  // Формируем текст для подсказки периода
  const periodText = useMemo(() => {
    if (dateFrom && dateTo) {
      return `с ${dateFrom.toLocaleDateString("ru-RU", { month: "short", year: "numeric" })} по ${dateTo.toLocaleDateString("ru-RU", { month: "short", year: "numeric" })}`;
    }
    if (dateFrom) {
      return `с ${dateFrom.toLocaleDateString("ru-RU", { month: "short", year: "numeric" })}`;
    }
    if (dateTo) {
      return `до ${dateTo.toLocaleDateString("ru-RU", { month: "short", year: "numeric" })}`;
    }
    return "за все время";
  }, [dateFrom, dateTo]);

  // Карточки для компаний (первый ряд) - 5 карточек
  const companyCards: KPICard[] = [
    {
      title: "Всего компаний",
      value: statistics?.totalCompanies,
      icon: <Building2 className="h-4 w-4" />,
      colorClass: "text-blue-500",
    },
    {
      title: "Активные компании",
      value: statistics?.activeCompanies,
      icon: <TrendingUp className="h-4 w-4" />,
      colorClass: "text-green-500",
    },
    {
      title: "Ликвидировано (ЮЛ)",
      value: statistics?.liquidatedCompanies,
      subtitle: "всего в базе",
      icon: <TrendingDown className="h-4 w-4" />,
      colorClass: "text-red-500",
    },
    {
      title: "Зарегистрировано (ЮЛ)",
      value: companyRegistrations,
      subtitle: periodText,
      icon: <TrendingUp className="h-4 w-4" />,
      colorClass: "text-emerald-500",
    },
    {
      title: "Ликвидировано за период (ЮЛ)",
      value: companyTerminations,
      subtitle: periodText,
      icon: <TrendingDown className="h-4 w-4" />,
      colorClass: "text-red-600",
    },
  ];

  // Карточки для ИП (второй ряд) - 5 карточек
  const entrepreneurCards: KPICard[] = [
    {
      title: "Всего ИП",
      value: statistics?.totalEntrepreneurs,
      icon: <Users className="h-4 w-4" />,
      colorClass: "text-purple-500",
    },
    {
      title: "Активные ИП",
      value: statistics?.activeEntrepreneurs,
      icon: <TrendingUp className="h-4 w-4" />,
      colorClass: "text-violet-500",
    },
    {
      title: "Ликвидировано (ИП)",
      value: statistics?.liquidatedEntrepreneurs,
      subtitle: "всего в базе",
      icon: <TrendingDown className="h-4 w-4" />,
      colorClass: "text-orange-500",
    },
    {
      title: "Зарегистрировано (ИП)",
      value: entrepreneurRegistrations,
      subtitle: periodText,
      icon: <TrendingUp className="h-4 w-4" />,
      colorClass: "text-teal-500",
    },
    {
      title: "Ликвидировано за период (ИП)",
      value: entrepreneurTerminations,
      subtitle: periodText,
      icon: <TrendingDown className="h-4 w-4" />,
      colorClass: "text-rose-500",
    },
  ];

  // Определяем какой ряд неактивен
  const isCompanyInactive = entityType === "entrepreneur";
  const isEntrepreneurInactive = entityType === "company";

  // Показываем skeleton если загружаются основные данные или данные регистраций
  const showLoading = isLoading || isLoadingCompanies || isLoadingEntrepreneurs;

  if (showLoading) {
    return (
      <div className="space-y-4">
        {/* Skeleton для компаний */}
        <div className="grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-5">
          {companyCards.map((_, index) => (
            <Card key={index}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-4 w-4 rounded" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-8 w-24" />
              </CardContent>
            </Card>
          ))}
        </div>

        {/* Skeleton для ИП */}
        <div className="grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-5">
          {entrepreneurCards.map((_, index) => (
            <Card key={index}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-4 w-4 rounded" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-8 w-24" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Карточки для компаний */}
      <div className="grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-5">
        {companyCards.map((card, index) => (
          <Card
            key={index}
            className={cn(
              "hover:shadow-md transition-all",
              isCompanyInactive && "opacity-50 grayscale cursor-not-allowed"
            )}
          >
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {card.title}
              </CardTitle>
              <div className={card.colorClass}>{card.icon}</div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {card.value !== undefined
                  ? card.value.toLocaleString("ru-RU")
                  : "—"}
              </div>
              {(card.subtitle || (card.value !== undefined && statistics)) && (
                <p className="text-xs text-muted-foreground mt-1">
                  {card.subtitle || getPercentageText(card, statistics!)}
                </p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Карточки для ИП */}
      <div className="grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-5">
        {entrepreneurCards.map((card, index) => (
          <Card
            key={index}
            className={cn(
              "hover:shadow-md transition-all",
              isEntrepreneurInactive && "opacity-50 grayscale cursor-not-allowed"
            )}
          >
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {card.title}
              </CardTitle>
              <div className={card.colorClass}>{card.icon}</div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {card.value !== undefined
                  ? card.value.toLocaleString("ru-RU")
                  : "—"}
              </div>
              {(card.subtitle || (card.value !== undefined && statistics)) && (
                <p className="text-xs text-muted-foreground mt-1">
                  {card.subtitle || getPercentageText(card, statistics!)}
                </p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}

/**
 * Получить текст процента для карточки
 */
function getPercentageText(card: KPICard, statistics: Statistics): string {
  const { title, value } = card;

  if (value === undefined) return "";

  // Процент активных компаний
  if (title === "Активные компании" && statistics.totalCompanies > 0) {
    const percentage = ((value / statistics.totalCompanies) * 100).toFixed(1);
    return `${percentage}% от всех компаний`;
  }

  // Процент активных ИП
  if (title === "Активные ИП" && statistics.totalEntrepreneurs > 0) {
    const percentage = (
      (value / statistics.totalEntrepreneurs) *
      100
    ).toFixed(1);
    return `${percentage}% от всех ИП`;
  }

  return "";
}
