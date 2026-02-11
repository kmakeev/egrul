"use client";

import { useQuery, type UseQueryOptions } from "@tanstack/react-query";
import { defaultGraphQLClient } from "@/lib/api/graphql-client";

// ==================== Типы данных ====================

export interface StatsFilter {
  regionCode?: string;
  okved?: string;
  dateFrom?: string;
  dateTo?: string;
}

export interface TimeSeriesPoint {
  month: string;
  registrationsCount: number;
  terminationsCount: number;
  netGrowth: number;
}

export interface RegionStatistics {
  regionCode: string;
  regionName: string;
  companiesCount: number;
  entrepreneursCount: number;
  activeCount: number;
  liquidatedCount: number;
}

export interface Statistics {
  totalCompanies: number;
  totalEntrepreneurs: number;
  activeCompanies: number;
  activeEntrepreneurs: number;
  liquidatedCompanies: number;
  liquidatedEntrepreneurs: number;
  registeredToday: number;
  registeredThisMonth: number;
  registeredThisYear: number;
}

export interface DashboardStatistics {
  registrationsByMonth: TimeSeriesPoint[];
  regionHeatmap: RegionStatistics[];
}

interface GetDashboardStatisticsResponse {
  statistics: Statistics;
  dashboardStatistics: DashboardStatistics;
}

// ==================== Хуки ====================

export interface UseDashboardStatisticsOptions {
  filter?: StatsFilter;
  dateFrom?: string;
  dateTo?: string;
  entityType?: string; // "COMPANY" | "ENTREPRENEUR"
  enabled?: boolean;
}

/**
 * Hook для получения полной статистики дашборда
 * Включает KPI метрики, временной ряд и региональную статистику
 */
export function useDashboardStatistics(
  options: UseDashboardStatisticsOptions = {},
  queryOptions?: Omit<
    UseQueryOptions<GetDashboardStatisticsResponse, Error>,
    "queryKey" | "queryFn"
  >
) {
  const { filter, dateFrom, dateTo, entityType, enabled = true } = options;

  return useQuery<GetDashboardStatisticsResponse, Error>({
    queryKey: ["dashboard-statistics", filter, dateFrom, dateTo, entityType],
    queryFn: () =>
      defaultGraphQLClient.request<
        GetDashboardStatisticsResponse,
        {
          filter?: StatsFilter;
          dateFrom?: string;
          dateTo?: string;
          entityType?: string;
        }
      >(
        /* GraphQL */ `
          query GetDashboardStatistics($filter: StatsFilter, $dateFrom: Date, $dateTo: Date, $entityType: EntityType) {
            statistics(filter: $filter) {
              totalCompanies
              totalEntrepreneurs
              activeCompanies
              activeEntrepreneurs
              liquidatedCompanies
              liquidatedEntrepreneurs
              registeredToday
              registeredThisMonth
              registeredThisYear
            }

            dashboardStatistics(filter: $filter) {
              registrationsByMonth(dateFrom: $dateFrom, dateTo: $dateTo, entityType: $entityType) {
                month
                registrationsCount
                terminationsCount
                netGrowth
              }

              regionHeatmap {
                regionCode
                regionName
                companiesCount
                entrepreneursCount
                activeCount
                liquidatedCount
              }
            }
          }
        `,
        { filter, dateFrom, dateTo, entityType }
      ),
    enabled,
    staleTime: 5 * 60 * 1000, // 5 минут кэш
    placeholderData: (previousData) => previousData, // Сохраняем предыдущие данные при обновлении
    ...queryOptions,
  });
}

/**
 * Hook только для KPI метрик (более легкий запрос)
 */
export function useStatistics(
  filter?: StatsFilter,
  queryOptions?: Omit<UseQueryOptions<Statistics, Error>, "queryKey" | "queryFn">
) {
  return useQuery<Statistics, Error>({
    queryKey: ["statistics", filter],
    queryFn: async () => {
      const response = await defaultGraphQLClient.request<
        { statistics: Statistics },
        { filter?: StatsFilter }
      >(
        /* GraphQL */ `
          query GetStatistics($filter: StatsFilter) {
            statistics(filter: $filter) {
              totalCompanies
              totalEntrepreneurs
              activeCompanies
              activeEntrepreneurs
              liquidatedCompanies
              liquidatedEntrepreneurs
              registeredToday
              registeredThisMonth
              registeredThisYear
            }
          }
        `,
        { filter }
      );
      return response.statistics;
    },
    staleTime: 5 * 60 * 1000,
    ...queryOptions,
  });
}

/**
 * Hook для временного ряда регистраций
 */
export function useRegistrationsByMonth(
  dateFrom?: string,
  dateTo?: string,
  entityType?: string, // "COMPANY" | "ENTREPRENEUR"
  filter?: StatsFilter,
  queryOptions?: Omit<UseQueryOptions<TimeSeriesPoint[], Error>, "queryKey" | "queryFn">
) {
  return useQuery<TimeSeriesPoint[], Error>({
    queryKey: ["registrations-by-month", dateFrom, dateTo, entityType, filter],
    queryFn: async () => {
      const response = await defaultGraphQLClient.request<
        { dashboardStatistics: DashboardStatistics },
        { filter?: StatsFilter; dateFrom?: string; dateTo?: string; entityType?: string }
      >(
        /* GraphQL */ `
          query GetRegistrationsByMonth($filter: StatsFilter, $dateFrom: Date, $dateTo: Date, $entityType: EntityType) {
            dashboardStatistics(filter: $filter) {
              registrationsByMonth(dateFrom: $dateFrom, dateTo: $dateTo, entityType: $entityType) {
                month
                registrationsCount
                terminationsCount
                netGrowth
              }
            }
          }
        `,
        { filter, dateFrom, dateTo, entityType }
      );
      return response.dashboardStatistics.registrationsByMonth;
    },
    staleTime: 5 * 60 * 1000,
    placeholderData: (previousData) => previousData, // Сохраняем предыдущие данные при обновлении
    ...queryOptions,
  });
}

/**
 * Hook для региональной тепловой карты
 */
export function useRegionHeatmap(
  filter?: StatsFilter,
  queryOptions?: Omit<UseQueryOptions<RegionStatistics[], Error>, "queryKey" | "queryFn">
) {
  return useQuery<RegionStatistics[], Error>({
    queryKey: ["region-heatmap", filter],
    queryFn: async () => {
      const response = await defaultGraphQLClient.request<
        { dashboardStatistics: DashboardStatistics },
        { filter?: StatsFilter }
      >(
        /* GraphQL */ `
          query GetRegionHeatmap($filter: StatsFilter) {
            dashboardStatistics(filter: $filter) {
              regionHeatmap {
                regionCode
                regionName
                companiesCount
                entrepreneursCount
                activeCount
                liquidatedCount
              }
            }
          }
        `,
        { filter }
      );
      return response.dashboardStatistics.regionHeatmap;
    },
    staleTime: 5 * 60 * 1000,
    ...queryOptions,
  });
}
