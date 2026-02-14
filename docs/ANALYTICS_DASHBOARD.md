# Аналитический дашборд ЕГРЮЛ/ЕГРИП

## Обзор

Интерактивный аналитический дашборд для визуализации агрегированных данных о юридических лицах и индивидуальных предпринимателях с правильной логикой определения статусов.

**Ключевые возможности:**
- 📊 KPI карточки с основными показателями (10 карточек: 5 для компаний + 5 для ИП)
- 🗺️ Интерактивная тепловая карта регионов России с динамическим выделением
- 📈 График динамики регистраций и ликвидаций
- 🔍 Фильтрация по региону, ОКВЭД, периоду (выбор месяцев)
- 💾 Сохранение настроек фильтров в localStorage
- ⚡ Кэширование запросов (5 минут)

## Используемые технологии

### Frontend
- **React 19** + **Next.js 15** (App Router) - с исправлением React Hydration Error через useEffect
- **Apache ECharts** - библиотека для графиков и карт с динамическим управлением выделением
- **TanStack Query** - управление состоянием и кэшированием
- **Radix UI** - UI компоненты
- **Tailwind CSS** - стилизация
- **date-fns** - работа с датами

### Backend
- **GraphQL** (gqlgen) - API слой
- **ClickHouse Cluster** - аналитическая БД с Materialized Views
- **Go** - бизнес-логика и репозиторий

## Архитектура решения

### Backend: Materialized Views

Для эффективной аналитики используются предагрегированные Materialized Views в ClickHouse:

1. **stats_companies_by_region** - агрегация компаний по регионам и статусам
2. **stats_entrepreneurs_by_region** - агрегация ИП по регионам и статусам
3. **stats_registrations_by_month** - временной ряд регистраций
4. **stats_terminations_by_month** - временной ряд прекращений деятельности

**Ключевые особенности:**
- Использует `AggregatingMergeTree` (без репликации для упрощения)
- Автоматически обновляется при вставке новых данных
- Хранит агрегатные функции (`countState()`)
- Чтение через `countMerge()` для получения итоговых значений

**Миграции:**
- `013_update_mv_status_logic.sql` - основная логика статусов
- `018_fix_terminations_mv_logic.sql` - исправление логики ликвидаций (COALESCE + status_code)

### Логика определения статусов

#### Компании (ЕГРЮЛ)

Логика полностью соответствует `company-status-badge.tsx`:

```sql
multiIf(
    -- Банкротство (приоритетнее чем просто ликвидация)
    status_code IN ('113', '114', '115', '116', '117'), 'bankrupt',

    -- Ликвидирована
    termination_date IS NOT NULL OR
    status_code IN ('101', '105', '106', '107', '113', '114', '115',
                    '116', '117', '701', '702', '801', '802'), 'liquidated',

    -- Активная (включая реорганизацию 121-139)
    'active'
) as status
```

**Статусы:**
- `active` - нет termination_date И код НЕ в списке недействующих
- `liquidated` - есть termination_date ИЛИ код недействующих (101, 105-107, 701-702, 801-802)
- `bankrupt` - коды банкротства (113-117), подмножество liquidated

#### Индивидуальные предприниматели (ЕГРИП)

Логика полностью соответствует `entrepreneur-status-badge.tsx`:

```sql
if(
    termination_date IS NULL AND status_code IS NULL,
    'active',
    'liquidated'
) as status
```

**Статусы:**
- `active` - НЕТ termination_date И НЕТ status_code (NULL)
- `liquidated` - ЕСТЬ termination_date ИЛИ ЕСТЬ любой status_code

### GraphQL API

**Endpoint:** `http://localhost:8080/graphql`

#### Новые типы

```graphql
"""
Точка временного ряда (регистрации/ликвидации)
"""
type TimeSeriesPoint {
  month: Date!
  registrationsCount: Int!
  terminationsCount: Int!
  netGrowth: Int!
}

"""
Расширенная статистика для дашборда
"""
type DashboardStatistics {
  # Временные ряды
  registrationsByMonth(dateFrom: Date, dateTo: Date): [TimeSeriesPoint!]!

  # Региональная статистика для тепловой карты (ВСЕ регионы)
  regionHeatmap: [RegionStatistics!]!
}
```

#### Query

```graphql
query GetDashboardStatistics($filter: StatsFilter) {
  dashboardStatistics(filter: $filter) {
    registrationsByMonth(dateFrom: $filter.dateFrom, dateTo: $filter.dateTo) {
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

  # Переиспользуем существующий query для KPI
  statistics(filter: $filter) {
    totalCompanies
    totalEntrepreneurs
    activeCompanies
    activeEntrepreneurs
    liquidatedCompanies
    liquidatedEntrepreneurs
  }
}
```

#### Примеры использования

**Все данные без фильтров:**
```graphql
query {
  dashboardStatistics {
    regionHeatmap {
      regionCode
      regionName
      companiesCount
      activeCount
    }
  }
}
```

**С фильтром по региону:**
```graphql
query {
  dashboardStatistics(filter: { regionCode: "77" }) {
    registrationsByMonth(
      dateFrom: "2024-01-01",
      dateTo: "2024-12-31"
    ) {
      month
      registrationsCount
      terminationsCount
      netGrowth
    }
  }
}
```

### Repository и Service слои

**Файл:** `services/api-gateway/internal/repository/clickhouse/statistics.go`

**Ключевые методы:**

```go
// GetCompanyCounts - подсчет компаний по статусам с фильтрацией
func (r *StatisticsRepository) GetCompanyCounts(
    ctx context.Context,
    filter *model.StatsFilter,
) (*model.EntityCounts, error) {
    // Использует stats_companies_by_region MV
    // Логика статусов согласно company-status-badge.tsx
    baseQuery := `
        SELECT
            countMerge(count) as total,
            countMergeIf(count, status = 'active') as active,
            countMergeIf(count, status IN ('liquidated', 'bankrupt')) as liquidated
        FROM egrul.stats_companies_by_region
    `
}

// GetEntrepreneurCounts - подсчет ИП по статусам с фильтрацией
func (r *StatisticsRepository) GetEntrepreneurCounts(
    ctx context.Context,
    filter *model.StatsFilter,
) (*model.EntityCounts, error) {
    // Использует stats_entrepreneurs_by_region MV
    // Логика статусов согласно entrepreneur-status-badge.tsx
    baseQuery := `
        SELECT
            countMerge(count) as total,
            countMergeIf(count, status = 'active') as active,
            countMergeIf(count, status = 'liquidated') as liquidated
        FROM egrul.stats_entrepreneurs_by_region
    `
}
```

## Frontend

### Структура компонентов

```
frontend/src/
├── app/(dashboard)/analytics/
│   └── page.tsx                               # Главная страница дашборда
├── components/analytics/
│   ├── use-analytics-filters.ts               # Хук управления фильтрами
│   ├── analytics-filters.tsx                  # Панель фильтров
│   ├── month-range-picker.tsx                 # Выбор месяцев (не дат)
│   ├── kpi-cards.tsx                          # KPI карточки (10 карточек: 5+5)
│   ├── region-map-chart.tsx                   # Карта регионов с динамическим выделением
│   └── registrations-timeline-chart.tsx       # График регистраций
└── lib/api/
    ├── hooks.ts                               # TanStack Query hooks
    └── queries/dashboard-statistics.graphql   # GraphQL queries
```

### Компоненты

#### 1. KPI Cards

**Файл:** `frontend/src/components/analytics/kpi-cards.tsx`

Отображает основные показатели в виде карточек: **10 карточек в 2 ряда по 5 колонок**

**Первый ряд (Компании):**
1. Всего компаний
2. Активные компании
3. Ликвидировано (ЮЛ) - всего в базе
4. Зарегистрировано (ЮЛ) - за выбранный период
5. Ликвидировано за период (ЮЛ) - за выбранный период

**Второй ряд (ИП):**
1. Всего ИП
2. Активные ИП
3. Ликвидировано (ИП) - всего в базе
4. Зарегистрировано (ИП) - за выбранный период
5. Ликвидировано за период (ИП) - за выбранный период

**Особенности:**
- Skeleton loading states
- Форматирование чисел (`toLocaleString('ru-RU')`)
- Адаптивная сетка (`grid-cols-2 md:grid-cols-3 lg:grid-cols-5`)
- Динамический расчет ликвидаций за период из `terminationsCount`
- Неактивный ряд становится полупрозрачным при фильтре по типу организации

#### 2. Region Map Chart

**Файл:** `frontend/src/components/analytics/region-map-chart.tsx`

Интерактивная хороплет-карта России с данными по регионам.

**Возможности:**
- Zoom и pan (roam: true)
- Tooltips с детальной информацией
- Клик по региону для фильтрации
- Цветовая шкала (visualMap) от светлого к темному
- Подсветка региона при hover
- **Динамический сброс выделения** через `selectedRegionCode` prop

**Управление выделением региона:**
```typescript
// Компонент принимает selectedRegionCode prop
interface RegionMapChartProps {
  selectedRegionCode?: string;
  onRegionClick?: (regionCode: string) => void;
}

// Использует ref для доступа к ECharts instance
const chartRef = useRef<any>(null);

// useEffect сбрасывает выделение при очистке фильтра
useEffect(() => {
  if (!selectedRegionCode && chartRef.current) {
    const instance = chartRef.current.getEchartsInstance();
    if (instance) {
      instance.dispatchAction({
        type: 'unselect',
        seriesIndex: 0,
      });
    }
  }
}, [selectedRegionCode]);
```

**Технические детали:**
```typescript
{
  visualMap: {
    min: 0,
    max: maxValue,
    inRange: {
      color: ['#e0f2fe', '#0ea5e9', '#0369a1', '#075985', '#0c4a6e']
    }
  },
  series: [{
    type: 'map',
    map: 'russia',  // Встроенная карта ECharts
    roam: true,
    emphasis: {
      itemStyle: { areaColor: '#fbbf24' }
    }
  }]
}
```

#### 3. Registrations Timeline Chart

**Файл:** `frontend/src/components/analytics/registrations-timeline-chart.tsx`

Линейный график с тремя линиями:
- 🟢 Регистрации (зеленая)
- 🔴 Ликвидации (красная)
- 🔵 Прирост (синяя, с заливкой)

**Особенности:**
- Smooth curves (smooth: true)
- Cross-axis pointer для точного чтения значений
- Легенда с возможностью скрытия линий
- Адаптивная высота (350px)

#### 4. Analytics Filters

**Файл:** `frontend/src/components/analytics/analytics-filters.tsx`

Панель фильтрации с четырьмя компонентами:
- Выбор типа организации (ЮЛ / ИП / Все)
- Выбор региона (RegionSelect)
- Выбор ОКВЭД (OkvedSelect)
- Выбор периода (MonthRangePicker)
- Кнопка сброса

**Сохранение в localStorage:**
```typescript
useEffect(() => {
  if (typeof window !== 'undefined') {
    localStorage.setItem('analytics-filters', JSON.stringify(filters));
  }
}, [filters]);
```

**Активные фильтры:**
- Отображаются как бейджи под панелью фильтров
- Каждый бейдж можно удалить кликом на "X"

#### 5. Month Range Picker

**Файл:** `frontend/src/components/analytics/month-range-picker.tsx`

Компонент выбора диапазона месяцев (не дат!) для фильтрации данных.

**Особенности:**
- Выбор "с месяца" и "по месяц" через отдельные Select
- Возвращает даты начала и конца выбранных месяцев
- Кнопка "Сбросить" для очистки выбора
- Кнопка "Применить" для сохранения выбора

**UI:**
- Кнопка с иконкой календаря и выбранным диапазоном
- Popover с двумя группами Select (месяц + год)
- Форматирование диапазона: "янв. 2024 — дек. 2024"

**Исправление React Hydration Error:**
```typescript
// Инициализация состояния как undefined для избежания hydration mismatch
const [fromMonth, setFromMonth] = useState<number | undefined>(undefined);
const [fromYear, setFromYear] = useState<number | undefined>(undefined);

// Синхронизация с props после монтирования через useEffect
useEffect(() => {
  setFromMonth(dateFrom ? dateFrom.getMonth() : undefined);
  setFromYear(dateFrom ? dateFrom.getFullYear() : undefined);
  setToMonth(dateTo ? dateTo.getMonth() : undefined);
  setToYear(dateTo ? dateTo.getFullYear() : undefined);
}, [dateFrom, dateTo]);
```

**Обработка дат:**
```typescript
// Первый день месяца "с"
newDateFrom = new Date(fromYear, fromMonth, 1);

// Последний день месяца "по" (23:59:59.999)
newDateTo = new Date(toYear, toMonth + 1, 0, 23, 59, 59, 999);
```

### React Hooks

#### useDashboardStatistics

**Файл:** `frontend/src/lib/api/hooks.ts`

```typescript
export function useDashboardStatistics({
  filter,
  dateFrom,
  dateTo,
}: {
  filter?: StatsFilter;
  dateFrom?: Date;
  dateTo?: Date;
}) {
  return useQuery({
    queryKey: ['dashboard-statistics', filter, dateFrom, dateTo],
    queryFn: async () => {
      const response = await graphQLClient.request(
        GetDashboardStatisticsDocument,
        {
          filter: filter || {},
          dateFrom: dateFrom?.toISOString(),
          dateTo: dateTo?.toISOString(),
        }
      );
      return response;
    },
    staleTime: 5 * 60 * 1000, // 5 минут кэш
    retry: 2,
  });
}
```

**Особенности:**
- Автоматическое кэширование на 5 минут
- Retry логика (2 попытки)
- Зависимости в queryKey для правильной инвалидации

#### useAnalyticsFilters

**Файл:** `frontend/src/components/analytics/use-analytics-filters.ts`

```typescript
export function useAnalyticsFilters() {
  const [filters, setFilters] = useState<AnalyticsFilters>(() => {
    // Загрузка из localStorage
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('analytics-filters');
      if (saved) return JSON.parse(saved);
    }
    return {};
  });

  useEffect(() => {
    // Сохранение в localStorage
    if (typeof window !== 'undefined') {
      localStorage.setItem('analytics-filters', JSON.stringify(filters));
    }
  }, [filters]);

  return { filters, setFilters, resetFilters };
}
```

## Производительность

### Backend оптимизации

1. **Materialized Views** - предагрегированные данные
   - Избегаем тяжелых COUNT(*) запросов на больших таблицах
   - Агрегация происходит инкрементально при вставке данных

2. **AggregatingMergeTree** - эффективное хранение агрегатов
   - countState() для хранения, countMerge() для чтения
   - Автоматическая агрегация при слиянии парт

3. **Опциональный Redis кэш** (5 минут TTL)
   ```go
   cacheKey := fmt.Sprintf("dashboard:stats:%s:%s:%s",
       filter.RegionCode, filter.Okved, filter.DateFrom)
   ```

### Frontend оптимизации

1. **TanStack Query кэширование** - 5 минут staleTime
2. **ECharts tree-shaking** - только используемые компоненты
   ```typescript
   import { LineChart, MapChart } from 'echarts/charts';
   // НЕ: import * as echarts from 'echarts';
   ```
3. **Lazy loading компонентов** (опционально)
   ```typescript
   const RegionMapChart = lazy(() => import('./region-map-chart'));
   ```

### Размер bundle

- ECharts после tree-shaking: ~150KB (gzip)
- Карта России: встроена в ECharts
- Общий прирост bundle: ~180KB

## Адаптивный дизайн

### Breakpoints

- **Mobile** (< 768px):
  - KPI cards: 1 колонка
  - Карта: 300px высота
  - Фильтры: вертикальный стек

- **Tablet** (768-1024px):
  - KPI cards: 2 колонки
  - Карта: 400px высота
  - Фильтры: горизонтальная панель

- **Desktop** (> 1024px):
  - KPI cards: 3 колонки
  - Карта: 500px высота
  - Фильтры: горизонтальная панель

### Touch-friendly

- Карта поддерживает touch-to-zoom и pan
- Tooltips показываются по tap на мобильных
- Фильтры адаптированы для touch интерфейса

## Доступность (a11y)

✅ **Keyboard navigation**
- Tab, Enter, Space для всех интерактивных элементов
- Focus visible (Radix UI)

✅ **ARIA labels**
- Графики имеют aria-label
- Кнопки и инпуты имеют aria-label/aria-labelledby

✅ **Screen reader support**
- ECharts ARIA описания для графиков
- Семантический HTML

✅ **Контрастность**
- Темная тема с правильными цветами
- WCAG AA compliance

## Установка и запуск

### Предварительные требования

1. ClickHouse кластер запущен
2. Применены миграции 013 и 018
3. Данные импортированы

### Команды

```bash
# 1. Пересоздать БД с автоматическим применением всех миграций (011-018)
make cluster-reset

# 2. Импортировать данные с автоматическим заполнением MV
make cluster-import

# 3. Запустить frontend dev server
cd frontend && pnpm dev

# 4. Открыть дашборд
# http://localhost:3000/analytics
```

**Что происходит автоматически:**
- `make cluster-reset` применяет миграции 011, 012, 013, 014, 015, 016, 017, 018
- `make cluster-import` импортирует данные и автоматически вызывает `make cluster-fill-mv` для заполнения Materialized Views
- Миграция 018 исправляет логику MV для ликвидаций (COALESCE + status_code)

### Проверка работоспособности

**1. Проверить MV в ClickHouse:**
```sql
-- Должны быть данные
SELECT status, countMerge(count) as total
FROM egrul.stats_companies_by_region
GROUP BY status;

-- Результат:
-- active:      ~182,847
-- liquidated:  ~26,685
-- bankrupt:    подмножество liquidated
```

**2. Проверить GraphQL API:**
```graphql
query {
  dashboardStatistics {
    regionHeatmap {
      regionCode
      regionName
      companiesCount
    }
  }
}
```

**3. Проверить frontend:**
- Открыть http://localhost:3000/analytics
- Все графики должны отобразиться
- Фильтры должны работать
- Клик на карту должен фильтровать данные

## Известные ограничения

1. **Batch updates MV** - при массовом импорте MV обновляются асинхронно, может быть задержка
2. **Историческая статистика** - изменение статуса существующих записей не обновляет MV автоматически (требуется пересчет)
3. **Bundle size** - ECharts добавляет ~180KB к bundle (trade-off за богатый функционал)

## Исправленные проблемы

1. **React Hydration Error #418** - исправлено через правильную инициализацию состояния
   - Проблема: MonthRangePicker инициализировал useState из props, вызывая несоответствие SSR/client
   - Решение: Инициализация как `undefined`, синхронизация через useEffect после монтирования

2. **Расхождение в цифрах ликвидаций** - исправлено миграцией 018
   - Проблема: MV использовала только `WHERE termination_date IS NOT NULL`, пропуская компании со status_code
   - Решение: Добавлена логика с COALESCE и проверкой status_code согласно Makefile

3. **Сброс фильтра не очищал карту** - исправлено через selectedRegionCode prop
   - Проблема: ECharts сохранял внутреннее состояние выделения
   - Решение: ref + useEffect + dispatchAction('unselect') при сбросе фильтра

## Дальнейшее развитие

### Запланировано в следующих версиях

1. **Экспорт данных**
   - PDF отчёт с графиками (jspdf + html2canvas)
   - Excel с таблицами данных (xlsx)

2. **Дополнительные виджеты**
   - Pie chart распределения по ОПФ
   - Bar chart топ отраслей по ОКВЭД
   - Bar chart среднего капитала по отраслям

3. **Продвинутые фильтры**
   - Диапазон капитала (min/max)
   - Множественный выбор регионов
   - Сравнение периодов (year-over-year)

4. **Интерактивность**
   - Drill-down с карты в детальные таблицы
   - Кросс-фильтрация между графиками
   - Bookmarks (сохранение конфигураций дашборда)

## Ссылки

- [План реализации](/Users/konstantin/.claude/plans/polished-percolating-cat.md)
- [Миграция 013](../infrastructure/migrations/clickhouse/cluster/013_update_mv_status_logic.sql) - основная логика статусов
- [Миграция 018](../infrastructure/migrations/clickhouse/cluster/018_fix_terminations_mv_logic.sql) - исправление логики ликвидаций
- [GraphQL Schema](../services/api-gateway/internal/graph/schema.graphqls)
- [Statistics Repository](../services/api-gateway/internal/repository/clickhouse/statistics.go)
- [Frontend Page](../frontend/src/app/(dashboard)/analytics/page.tsx)
- [Month Range Picker](../frontend/src/components/analytics/month-range-picker.tsx)
- [KPI Cards](../frontend/src/components/analytics/kpi-cards.tsx)

## Контрибьюторы

- Реализация backend: миграция 013, GraphQL API, repository/service слои
- Реализация frontend: React компоненты, ECharts интеграция, фильтры
- Документация: ANALYTICS_DASHBOARD.md

---

**Версия:** 1.1.0
**Дата:** 2025-02-11
**Статус:** ✅ Готово к использованию

**Изменения в версии 1.1.0:**
- Добавлена миграция 018 для исправления логики ликвидаций
- KPI карточки обновлены до 10 (5+5) с адаптивным grid 2/3/5
- Добавлена карточка "Ликвидировано за период"
- Замена DateRangePicker на MonthRangePicker для выбора месяцев
- Исправлен React Hydration Error #418
- Реализован механизм сброса выделения региона на карте
- Улучшена синхронизация фильтров с картой через selectedRegionCode prop
