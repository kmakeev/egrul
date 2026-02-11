"use client";

import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { RegionSelect } from "@/components/search/region-select";
import { OkvedSelect } from "@/components/search/okved-select";
import { MonthRangePicker } from "./month-range-picker";
import { X } from "lucide-react";
import type { AnalyticsFilters as Filters, EntityType } from "./use-analytics-filters";

interface AnalyticsFiltersProps {
  filters: Filters;
  onChange: (filters: Filters) => void;
  onReset: () => void;
}

/**
 * Компонент панели фильтров для аналитического дашборда
 */
export function AnalyticsFilters({
  filters,
  onChange,
  onReset,
}: AnalyticsFiltersProps) {
  const hasActiveFilters =
    filters.entityType ||
    filters.regionCode ||
    filters.okved ||
    filters.dateFrom ||
    filters.dateTo;

  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex flex-col gap-4">
          {/* Заголовок */}
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium text-muted-foreground">
              Фильтры
            </h3>
            {hasActiveFilters && (
              <Button
                variant="ghost"
                size="sm"
                onClick={onReset}
                className="h-8 px-2 text-xs"
              >
                <X className="mr-1 h-3 w-3" />
                Сбросить
              </Button>
            )}
          </div>

          {/* Фильтры */}
          <div className="flex flex-wrap gap-4 items-center">
            {/* Тип организации */}
            <div className="flex-1 min-w-[180px]">
              <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
                Тип организации
              </label>
              <Select
                value={filters.entityType || "all"}
                onValueChange={(value) =>
                  onChange({
                    ...filters,
                    entityType: value === "all" ? undefined : (value as EntityType),
                  })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Все" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Все</SelectItem>
                  <SelectItem value="company">Юридические лица</SelectItem>
                  <SelectItem value="entrepreneur">Индивидуальные предприниматели</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Регион */}
            <div className="flex-1 min-w-[200px]">
              <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
                Регион
              </label>
              <RegionSelect
                value={filters.regionCode}
                onChange={(regionCode) =>
                  onChange({ ...filters, regionCode: regionCode || undefined })
                }
              />
            </div>

            {/* ОКВЭД */}
            <div className="flex-1 min-w-[200px]">
              <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
                Вид деятельности (ОКВЭД)
              </label>
              <OkvedSelect
                value={filters.okved}
                onChange={(okved) =>
                  onChange({ ...filters, okved: okved || undefined })
                }
              />
            </div>

            {/* Период */}
            <div className="flex-1 min-w-[300px]">
              <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
                Период (по месяцам)
              </label>
              <MonthRangePicker
                dateFrom={filters.dateFrom}
                dateTo={filters.dateTo}
                onChange={(dateFrom, dateTo) =>
                  onChange({ ...filters, dateFrom, dateTo })
                }
              />
            </div>
          </div>

          {/* Активные фильтры (бейджи) */}
          {hasActiveFilters && (
            <div className="flex flex-wrap gap-2 pt-2 border-t">
              <span className="text-xs text-muted-foreground">
                Активные фильтры:
              </span>
              {filters.entityType && (
                <div className="inline-flex items-center gap-1 px-2 py-1 bg-primary/10 text-primary text-xs rounded-md">
                  <span>
                    {filters.entityType === "company"
                      ? "Юридические лица"
                      : "Индивидуальные предприниматели"}
                  </span>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-auto w-auto p-0 hover:bg-transparent"
                    onClick={() =>
                      onChange({ ...filters, entityType: undefined })
                    }
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </div>
              )}
              {filters.regionCode && (
                <div className="inline-flex items-center gap-1 px-2 py-1 bg-primary/10 text-primary text-xs rounded-md">
                  <span>Регион: {filters.regionCode}</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-auto w-auto p-0 hover:bg-transparent"
                    onClick={() =>
                      onChange({ ...filters, regionCode: undefined })
                    }
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </div>
              )}
              {filters.okved && (
                <div className="inline-flex items-center gap-1 px-2 py-1 bg-primary/10 text-primary text-xs rounded-md">
                  <span>ОКВЭД: {filters.okved}</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-auto w-auto p-0 hover:bg-transparent"
                    onClick={() =>
                      onChange({ ...filters, okved: undefined })
                    }
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </div>
              )}
              {(filters.dateFrom || filters.dateTo) && (
                <div className="inline-flex items-center gap-1 px-2 py-1 bg-primary/10 text-primary text-xs rounded-md">
                  <span>Период выбран</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-auto w-auto p-0 hover:bg-transparent"
                    onClick={() =>
                      onChange({
                        ...filters,
                        dateFrom: undefined,
                        dateTo: undefined,
                      })
                    }
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </div>
              )}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
