"use client";

import { useState, useEffect } from "react";

/**
 * Тип организации для фильтра
 */
export type EntityType = "all" | "company" | "entrepreneur";

/**
 * Фильтры для аналитического дашборда
 */
export interface AnalyticsFilters {
  entityType?: EntityType;
  regionCode?: string;
  okved?: string;
  dateFrom?: Date;
  dateTo?: Date;
}

/**
 * Hook для управления фильтрами дашборда с сохранением в localStorage
 */
export function useAnalyticsFilters() {
  const [filters, setFilters] = useState<AnalyticsFilters>(() => {
    // Загрузка из localStorage при инициализации
    if (typeof window !== "undefined") {
      try {
        const saved = localStorage.getItem("analytics-filters");
        if (saved) {
          const parsed = JSON.parse(saved);
          return {
            ...parsed,
            // Конвертируем ISO строки обратно в Date объекты
            dateFrom: parsed.dateFrom ? new Date(parsed.dateFrom) : undefined,
            dateTo: parsed.dateTo ? new Date(parsed.dateTo) : undefined,
          };
        }
      } catch (error) {
        console.error("Failed to load analytics filters from localStorage:", error);
      }
    }
    return {};
  });

  useEffect(() => {
    // Сохранение в localStorage при изменении фильтров
    if (typeof window !== "undefined") {
      try {
        localStorage.setItem("analytics-filters", JSON.stringify(filters));
      } catch (error) {
        console.error("Failed to save analytics filters to localStorage:", error);
      }
    }
  }, [filters]);

  const resetFilters = () => {
    setFilters({});
  };

  const updateFilters = (updates: Partial<AnalyticsFilters>) => {
    setFilters((prev) => ({ ...prev, ...updates }));
  };

  return {
    filters,
    setFilters,
    updateFilters,
    resetFilters,
  };
}
