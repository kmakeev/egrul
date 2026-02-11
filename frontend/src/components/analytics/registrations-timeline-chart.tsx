"use client";

import { useMemo } from "react";
import ReactECharts from "echarts-for-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { format } from "date-fns";
import { ru } from "date-fns/locale";
import echarts, {
  darkTheme,
  darkAxisConfig,
  chartColors,
} from "@/lib/echarts-config";
import type { TimeSeriesPoint } from "@/lib/api/dashboard-hooks";
import type { EntityType } from "./use-analytics-filters";

interface RegistrationsTimelineChartProps {
  data?: TimeSeriesPoint[];
  entityType?: EntityType;
  isLoading?: boolean;
}

/**
 * Компонент графика динамики регистраций и ликвидаций
 * Отображает временной ряд с тремя линиями
 */
export function RegistrationsTimelineChart({
  data,
  entityType,
  isLoading,
}: RegistrationsTimelineChartProps) {
  const chartOption = useMemo(() => {
    if (!data || data.length === 0) {
      return {};
    }

    // Форматируем даты для оси X
    const months = data.map((d) =>
      format(new Date(d.month), "MMM yyyy", { locale: ru })
    );

    // Извлекаем данные для каждой линии
    const registrations = data.map((d) => d.registrationsCount);
    const terminations = data.map((d) => d.terminationsCount);
    const netGrowth = data.map((d) => d.netGrowth);

    return {
      // Применяем темную тему
      ...darkTheme,

      tooltip: {
        trigger: "axis",
        backgroundColor: "rgba(15, 23, 42, 0.95)",
        borderColor: "#334155",
        borderWidth: 1,
        textStyle: {
          color: "#f1f5f9",
        },
        axisPointer: {
          type: "cross",
          crossStyle: {
            color: "#94a3b8",
          },
        },
        formatter: (params: any) => {
          // Маппинг названий серий на цвета (соответствуют legend)
          const colorMap: Record<string, string> = {
            "Регистрации": chartColors.success,
            "Ликвидации": chartColors.danger,
            "Прирост": chartColors.primary,
          };

          let result = `<div style="padding: 4px;"><strong>${params[0].axisValue}</strong><br/>`;
          params.forEach((param: any) => {
            const color = colorMap[param.seriesName] || param.color;
            result += `
              <div style="display: flex; align-items: center; gap: 8px; margin-top: 4px;">
                <span style="display: inline-block; width: 10px; height: 10px; border-radius: 50%; background-color: ${color};"></span>
                <span>${param.seriesName}:</span>
                <strong>${param.value.toLocaleString("ru-RU")}</strong>
              </div>
            `;
          });
          result += "</div>";
          return result;
        },
      },

      legend: {
        data: ["Регистрации", "Ликвидации", "Прирост"],
        top: "5%",
        textStyle: {
          color: "#e2e8f0",
        },
      },

      grid: {
        left: "3%",
        right: "4%",
        bottom: "10%",
        top: "15%",
        containLabel: true,
      },

      xAxis: {
        type: "category",
        data: months,
        ...darkAxisConfig,
        axisLabel: {
          ...darkAxisConfig.axisLabel,
          rotate: 45,
          fontSize: 11,
        },
      },

      yAxis: {
        type: "value",
        ...darkAxisConfig,
        splitLine: {
          lineStyle: {
            color: "#334155",
            type: "dashed",
          },
        },
        axisLabel: {
          ...darkAxisConfig.axisLabel,
          formatter: (value: number) => value.toLocaleString("ru-RU"),
        },
      },

      series: [
        {
          name: "Регистрации",
          type: "line",
          data: registrations,
          smooth: true,
          showSymbol: true,
          symbol: "circle",
          symbolSize: 6,
          lineStyle: {
            width: 2,
            color: chartColors.success,
          },
          itemStyle: {
            color: chartColors.success,
          },
          areaStyle: {
            color: {
              type: "linear",
              x: 0,
              y: 0,
              x2: 0,
              y2: 1,
              colorStops: [
                { offset: 0, color: "rgba(16, 185, 129, 0.3)" },
                { offset: 1, color: "rgba(16, 185, 129, 0.05)" },
              ],
            },
          },
        },
        {
          name: "Ликвидации",
          type: "line",
          data: terminations,
          smooth: true,
          showSymbol: true,
          symbol: "circle",
          symbolSize: 6,
          lineStyle: {
            width: 2,
            color: chartColors.danger,
          },
          itemStyle: {
            color: chartColors.danger,
          },
          areaStyle: {
            color: {
              type: "linear",
              x: 0,
              y: 0,
              x2: 0,
              y2: 1,
              colorStops: [
                { offset: 0, color: "rgba(239, 68, 68, 0.3)" },
                { offset: 1, color: "rgba(239, 68, 68, 0.05)" },
              ],
            },
          },
        },
        {
          name: "Прирост",
          type: "line",
          data: netGrowth,
          smooth: true,
          showSymbol: true,
          symbol: "diamond",
          symbolSize: 8,
          lineStyle: {
            width: 3,
            color: chartColors.primary,
          },
          itemStyle: {
            color: chartColors.primary,
            borderColor: "#1e40af",
            borderWidth: 2,
          },
        },
      ],

      dataZoom: [
        {
          type: "inside",
          start: 0,
          end: 100,
        },
        {
          start: 0,
          end: 100,
          height: 20,
          bottom: "3%",
          borderColor: "#334155",
          textStyle: {
            color: "#94a3b8",
          },
          dataBackground: {
            lineStyle: {
              color: "#475569",
            },
            areaStyle: {
              color: "#1e293b",
            },
          },
          selectedDataBackground: {
            lineStyle: {
              color: "#3b82f6",
            },
            areaStyle: {
              color: "#1e3a8a",
            },
          },
        },
      ],
    };
  }, [data, entityType]);

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Динамика регистраций и ликвидаций</CardTitle>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-[400px] w-full" />
        </CardContent>
      </Card>
    );
  }

  if (!data || data.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Динамика регистраций и ликвидаций</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-[400px] flex items-center justify-center text-muted-foreground">
            Нет данных для отображения
          </div>
        </CardContent>
      </Card>
    );
  }

  // Определяем заголовок в зависимости от типа организации
  const chartTitle =
    entityType === "company"
      ? "Динамика регистраций и ликвидаций ЮЛ"
      : entityType === "entrepreneur"
      ? "Динамика регистраций и ликвидаций ИП"
      : "Динамика регистраций и ликвидаций";

  const chartSubtitle =
    entityType === "company"
      ? "Временной ряд регистраций, ликвидаций и чистого прироста юридических лиц"
      : entityType === "entrepreneur"
      ? "Временной ряд регистраций, ликвидаций и чистого прироста индивидуальных предпринимателей"
      : "Временной ряд регистраций, ликвидаций и чистого прироста компаний";

  return (
    <Card>
      <CardHeader>
        <CardTitle>{chartTitle}</CardTitle>
        <p className="text-sm text-muted-foreground">{chartSubtitle}</p>
      </CardHeader>
      <CardContent>
        <ReactECharts
          echarts={echarts}
          option={chartOption}
          style={{ height: "400px", width: "100%" }}
          opts={{ renderer: "canvas" }}
          theme="dark"
        />
      </CardContent>
    </Card>
  );
}
