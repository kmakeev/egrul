"use client";

import { useState, useEffect } from "react";
import * as echarts from "echarts/core";

/**
 * Hook для загрузки и регистрации GeoJSON карты России в ECharts
 * Загружает карту с публичного CDN и регистрирует её для использования
 */
export function useRussiaMap() {
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [isRegistered, setIsRegistered] = useState(false);

  useEffect(() => {
    // Проверяем, не зарегистрирована ли карта уже
    if (isRegistered) {
      setIsLoading(false);
      return;
    }

    const loadMap = async () => {
      try {
        setIsLoading(true);
        setError(null);

        // Загружаем GeoJSON карту России из локального файла
        // Карта содержит все федеральные субъекты РФ с ID в формате RU.XX
        const response = await fetch("/maps/russia.geo.json");

        if (!response.ok) {
          throw new Error(`Failed to load map: ${response.statusText}`);
        }

        const geoJson = await response.json();

        // ВАЖНО: Заменяем properties.name на id для правильного сопоставления данных
        // ECharts по умолчанию использует properties.name для поиска данных
        geoJson.features.forEach((feature: any) => {
          if (feature.id && feature.properties) {
            feature.properties.name = feature.id;
          }
        });

        // Регистрируем карту в ECharts
        echarts.registerMap("russia", geoJson);

        setIsRegistered(true);
        setIsLoading(false);
      } catch (err) {
        console.error("Error loading Russia map:", err);
        setError(err instanceof Error ? err : new Error("Unknown error"));
        setIsLoading(false);
      }
    };

    loadMap();
  }, [isRegistered]);

  return { isLoading, error, isRegistered };
}
