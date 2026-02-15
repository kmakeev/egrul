"use client";

import { useMemo, useRef, useEffect, useState } from "react";
import ReactECharts from "echarts-for-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import echarts, { darkTheme } from "@/lib/echarts-config";
import { useRussiaMap } from "@/hooks/use-russia-map";
import { getGeoJsonId } from "@/lib/region-code-to-geojson-id";
import type { RegionStatistics } from "@/lib/api/dashboard-hooks";
import type { EntityType } from "./use-analytics-filters";

interface RegionMapChartProps {
  data?: RegionStatistics[];
  entityType?: EntityType;
  isLoading?: boolean;
  onRegionClick?: (regionCode: string) => void;
  selectedRegionCode?: string; // Добавляем prop для отслеживания выбранного региона
}

export function RegionMapChart({
  data,
  entityType,
  isLoading,
  onRegionClick,
  selectedRegionCode,
}: RegionMapChartProps) {
  const {
    isLoading: isMapLoading,
    error: mapError,
    isRegistered,
  } = useRussiaMap();

  // Ref для доступа к ECharts instance
  const chartRef = useRef<any>(null);
  // Ref для отслеживания предыдущего выбранного региона
  const prevSelectedRegionRef = useRef<string | undefined>(undefined);
  // Счетчик для принудительной перерисовки при сбросе фильтра
  const [resetCounter, setResetCounter] = useState(0);

  // Отслеживаем сброс фильтра и увеличиваем счетчик для перерисовки
  useEffect(() => {
    // Проверяем: был ли регион выбран ранее, но сейчас сброшен
    const wasReset = !selectedRegionCode && prevSelectedRegionRef.current;

    if (wasReset) {
      // Увеличиваем счетчик - это изменит key и пересоздаст компонент
      setResetCounter((prev) => prev + 1);
    }

    // Сохраняем текущее значение для следующего сравнения
    prevSelectedRegionRef.current = selectedRegionCode;
  }, [selectedRegionCode]);

  const chartOption = useMemo(() => {
    if (!data || data.length === 0 || !isRegistered) {
      return {};
    }

    const getValue = (region: RegionStatistics) => {
      if (entityType === "company") return region.companiesCount;
      if (entityType === "entrepreneur") return region.entrepreneursCount;
      return region.companiesCount + region.entrepreneursCount;
    };

    const values = data.map(getValue).filter((v) => v > 0);
    const maxValue = Math.max(...values, 1);
    const minValue = Math.min(...values, 0);

    const mapData = data
      .map((region) => {
        // Получаем GeoJSON ID по коду региона (01 -> RU.AD, 77 -> RU.MS)
        const geoJsonId = getGeoJsonId(region.regionCode);

        if (!geoJsonId) {
          console.warn(`No GeoJSON ID found for region code: ${region.regionCode}`);
          return null;
        }

        return {
          name: geoJsonId, // ВАЖНО: name должен совпадать с id из GeoJSON
          regionName: region.regionName, // Полное название для tooltip
          value: getValue(region),
          regionCode: region.regionCode,
          companiesCount: region.companiesCount,
          entrepreneursCount: region.entrepreneursCount,
          activeCount: region.activeCount,
          liquidatedCount: region.liquidatedCount,
        };
      })
      .filter((item): item is NonNullable<typeof item> => item !== null);

    const option = {
      ...darkTheme,
      tooltip: {
        trigger: "item",
        backgroundColor: "rgba(15, 23, 42, 0.95)",
        borderColor: "#334155",
        borderWidth: 1,
        textStyle: { color: "#f1f5f9" },
        formatter: (params: any) => {
          const d = params.data || {};
          let content = `<div style="padding: 4px;">
            <div style="font-weight: 600; margin-bottom: 8px;">${d.regionName || params.name || "Нет данных"}</div>
            <div style="display: flex; flex-direction: column; gap: 4px;">`;

          if (entityType === "company") {
            content += `<div>Компаний: <strong>${(d.companiesCount || 0).toLocaleString("ru-RU")}</strong></div>`;
          } else if (entityType === "entrepreneur") {
            content += `<div>ИП: <strong>${(d.entrepreneursCount || 0).toLocaleString("ru-RU")}</strong></div>`;
          } else {
            content += `
              <div>Компаний: <strong>${(d.companiesCount || 0).toLocaleString("ru-RU")}</strong></div>
              <div>ИП: <strong>${(d.entrepreneursCount || 0).toLocaleString("ru-RU")}</strong></div>
              <div style="border-top: 1px solid #334155; margin-top: 4px; padding-top: 4px;">
                Всего: <strong>${((d.companiesCount || 0) + (d.entrepreneursCount || 0)).toLocaleString("ru-RU")}</strong>
              </div>`;
          }

          content += `
              <div>Активных: <strong>${(d.activeCount || 0).toLocaleString("ru-RU")}</strong></div>
              <div>Ликвидированных: <strong>${(d.liquidatedCount || 0).toLocaleString("ru-RU")}</strong></div>
            </div></div>`;
          return content;
        },
      },
      visualMap: {
        min: minValue,
        max: maxValue,
        text: ["Больше", "Меньше"],
        realtime: false,
        calculable: true,
        inRange: {
          color: ["#1e3a8a", "#1e40af", "#2563eb", "#3b82f6", "#60a5fa", "#93c5fd"],
        },
        textStyle: { color: "#e2e8f0" },
        orient: "horizontal",
        left: "center",
        bottom: "5%",
        formatter: (value: number) => value.toLocaleString("ru-RU"),
      },
      series: [{
        name: entityType === "company" ? "Компании" : entityType === "entrepreneur" ? "ИП" : "Организации",
        type: "map",
        map: "russia",
        roam: true,
        scaleLimit: { min: 0.5, max: 5 },
        selectedMode: 'single', // Позволяет выделять только один регион за раз
        // nameProperty не указываем - ECharts автоматически использует feature.id из GeoJSON
        emphasis: {
          label: { show: true, color: "#fff" },
          itemStyle: {
            areaColor: "#fbbf24",
            borderWidth: 2,
            borderColor: "#fff",
            shadowBlur: 10,
            shadowColor: "rgba(251, 191, 36, 0.5)",
          },
        },
        select: {
          label: { show: true, color: "#fff" },
          itemStyle: { areaColor: "#f59e0b", borderColor: "#fff", borderWidth: 2 },
        },
        itemStyle: {
          borderColor: "#1e40af",
          borderWidth: 0.5,
          areaColor: "#1e3a8a",
        },
        label: { show: false, fontSize: 10, color: "#f1f5f9" },
        data: mapData,
      }],
    };

    return option;
  }, [data, entityType, isRegistered]);

  const handleEvents = useMemo(
    () => ({
      click: (params: any) => {
        if (params.data?.regionCode && onRegionClick) {
          onRegionClick(params.data.regionCode);
        }
      },
    }),
    [onRegionClick]
  );

  if (isLoading || isMapLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Распределение по регионам России</CardTitle>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-[500px] w-full" />
        </CardContent>
      </Card>
    );
  }

  if (mapError) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Распределение по регионам России</CardTitle>
        </CardHeader>
        <CardContent>
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Ошибка загрузки карты: {mapError.message}
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
  }

  if (!data || data.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Распределение по регионам России</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-[500px] flex items-center justify-center text-muted-foreground">
            Нет данных для отображения
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Распределение по регионам России</CardTitle>
        <p className="text-sm text-muted-foreground">
          Интенсивность цвета = количество организаций. Нажмите на регион для фильтрации. Используйте колесико мыши для масштабирования
        </p>
      </CardHeader>
      <CardContent>
        <ReactECharts
          key={`map-${resetCounter}`}
          ref={chartRef}
          echarts={echarts}
          option={chartOption}
          style={{ height: "600px", width: "100%" }}
          onEvents={handleEvents}
          opts={{ renderer: "canvas" }}
          theme="dark"
        />
      </CardContent>
    </Card>
  );
}
