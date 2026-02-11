/**
 * ECharts Configuration
 *
 * Оптимизированная конфигурация ECharts с tree-shaking и темной темой
 * Регистрируем только необходимые компоненты для уменьшения bundle size
 */

import * as echarts from 'echarts/core';
import { LineChart, MapChart } from 'echarts/charts';
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  VisualMapComponent,
  GeoComponent,
  TitleComponent,
  DataZoomComponent,
} from 'echarts/components';
import { CanvasRenderer } from 'echarts/renderers';

// Регистрация только используемых компонентов (tree-shaking)
echarts.use([
  LineChart,
  MapChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  VisualMapComponent,
  GeoComponent,
  TitleComponent,
  DataZoomComponent,
  CanvasRenderer,
]);

/**
 * Темная тема для графиков (совместима с Tailwind dark mode)
 */
export const darkTheme = {
  backgroundColor: 'transparent',
  textStyle: {
    color: '#e2e8f0', // slate-200
    fontFamily: 'Inter, -apple-system, sans-serif',
  },
  title: {
    textStyle: {
      color: '#f1f5f9', // slate-100
      fontWeight: 600,
    },
    subtextStyle: {
      color: '#94a3b8', // slate-400
    },
  },
  legend: {
    textStyle: {
      color: '#e2e8f0', // slate-200
    },
    pageTextStyle: {
      color: '#e2e8f0',
    },
  },
  tooltip: {
    backgroundColor: 'rgba(15, 23, 42, 0.9)', // slate-900 with opacity
    borderColor: '#334155', // slate-700
    borderWidth: 1,
    textStyle: {
      color: '#f1f5f9', // slate-100
    },
  },
  grid: {
    borderColor: '#334155', // slate-700
  },
  visualMap: {
    textStyle: {
      color: '#e2e8f0', // slate-200
    },
  },
};

/**
 * Общие настройки для всех графиков
 */
export const commonChartOptions = {
  // Анимация
  animation: true,
  animationDuration: 750,
  animationEasing: 'cubicOut',

  // Адаптивность
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    top: '10%',
    containLabel: true,
  },

  // Тултип
  tooltip: {
    trigger: 'axis',
    axisPointer: {
      type: 'cross',
      crossStyle: {
        color: '#94a3b8', // slate-400
      },
    },
  },
};

/**
 * Цветовая палитра для графиков (Tailwind colors)
 */
export const chartColors = {
  primary: '#3b82f6',   // blue-500
  success: '#10b981',   // green-500
  warning: '#f59e0b',   // amber-500
  danger: '#ef4444',    // red-500
  info: '#06b6d4',      // cyan-500
  purple: '#a855f7',    // purple-500
  pink: '#ec4899',      // pink-500
  indigo: '#6366f1',    // indigo-500
};

/**
 * Настройки оси для темной темы
 */
export const darkAxisConfig = {
  axisLine: {
    lineStyle: {
      color: '#64748b', // slate-500
    },
  },
  axisTick: {
    lineStyle: {
      color: '#64748b',
    },
  },
  axisLabel: {
    color: '#94a3b8', // slate-400
  },
  splitLine: {
    lineStyle: {
      color: '#334155', // slate-700
      type: 'dashed',
    },
  },
};

export default echarts;
