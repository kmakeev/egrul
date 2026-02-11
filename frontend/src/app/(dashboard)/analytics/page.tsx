"use client";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";

// Hooks
import { useAnalyticsFilters } from "@/components/analytics/use-analytics-filters";
import { useDashboardStatistics } from "@/lib/api/dashboard-hooks";

// Компоненты дашборда
import { AnalyticsFilters } from "@/components/analytics/analytics-filters";
import { KPICards } from "@/components/analytics/kpi-cards";
import { RegionMapChart } from "@/components/analytics/region-map-chart";
import { RegistrationsTimelineChart } from "@/components/analytics/registrations-timeline-chart";

/**
 * Страница аналитического дашборда
 * Отображает KPI метрики, карту регионов и временные ряды
 */
export default function AnalyticsPage() {
  // Управление фильтрами с сохранением в localStorage
  const { filters, setFilters, resetFilters } = useAnalyticsFilters();

  // Формируем переменные для GraphQL запроса
  const dateFrom = filters.dateFrom
    ? filters.dateFrom.toISOString().split("T")[0]
    : undefined;
  const dateTo = filters.dateTo
    ? filters.dateTo.toISOString().split("T")[0]
    : undefined;

  const statsFilter = {
    regionCode: filters.regionCode,
    okved: filters.okved,
  };

  // Преобразуем entityType для GraphQL (company -> COMPANY, entrepreneur -> ENTREPRENEUR)
  const entityTypeGQL = filters.entityType
    ? filters.entityType === "company"
      ? "COMPANY"
      : "ENTREPRENEUR"
    : undefined;

  // Загружаем данные дашборда
  const { data, isLoading, error } = useDashboardStatistics({
    filter: statsFilter,
    dateFrom,
    dateTo,
    entityType: entityTypeGQL,
  });

  // Обработчик клика по региону на карте
  const handleRegionClick = (regionCode: string) => {
    setFilters({ ...filters, regionCode });
  };

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* Заголовок */}
      <div>
        <h1 className="text-3xl font-bold mb-2">Аналитический дашборд</h1>
        <p className="text-muted-foreground">
          Интерактивная визуализация данных ЕГРЮЛ/ЕГРИП с фильтрацией по
          регионам, отраслям и периодам
        </p>
      </div>

      {/* Фильтры */}
      <AnalyticsFilters
        filters={filters}
        onChange={setFilters}
        onReset={resetFilters}
      />

      {/* Ошибка загрузки */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Ошибка загрузки данных:{" "}
            {error instanceof Error ? error.message : "Неизвестная ошибка"}
          </AlertDescription>
        </Alert>
      )}

      {/* KPI карточки */}
      <KPICards
        statistics={data?.statistics}
        dateFrom={filters.dateFrom}
        dateTo={filters.dateTo}
        entityType={filters.entityType}
        filter={statsFilter}
        isLoading={isLoading}
      />

      {/* Карта регионов */}
      <RegionMapChart
        data={data?.dashboardStatistics.regionHeatmap}
        entityType={filters.entityType}
        isLoading={isLoading}
        onRegionClick={handleRegionClick}
        selectedRegionCode={filters.regionCode}
      />

      {/* График динамики регистраций и ликвидаций */}
      <RegistrationsTimelineChart
        data={data?.dashboardStatistics.registrationsByMonth}
        entityType={filters.entityType}
        isLoading={isLoading}
      />

      {/* Информация о данных */}
      {!isLoading && data && (
        <div className="text-xs text-muted-foreground text-center pt-4 border-t">
          Данные обновлены: {new Date().toLocaleString("ru-RU")} • Кэширование:
          5 минут
        </div>
      )}
    </div>
  );
}
