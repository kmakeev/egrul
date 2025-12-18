import { useMemo, useEffect, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useDebounce } from "@/hooks/use-debounce";
import {
  useSearchCompaniesQuery,
  useSearchEntrepreneursQuery,
} from "@/lib/api/hooks";
import type { LegalEntity, IndividualEntrepreneur } from "@/lib/api";
import { searchFiltersSchema, type SearchFiltersInput } from "@/lib/validations";

/**
 * Определяет тип поискового запроса
 * @param query - Поисковый запрос
 * @returns Тип запроса: 'inn' | 'ogrn' | 'name'
 */
function detectQueryType(query: string): 'inn' | 'ogrn' | 'name' {
  const trimmed = query.trim();
  
  // Проверяем, что запрос состоит только из цифр
  if (!/^\d+$/.test(trimmed)) {
    return 'name';
  }
  
  // ИНН: 10 или 12 цифр
  if (/^\d{10}$|^\d{12}$/.test(trimmed)) {
    return 'inn';
  }
  
  // ОГРН: 13 цифр (для ЮЛ) или 15 цифр (для ИП)
  if (/^\d{13}$|^\d{15}$/.test(trimmed)) {
    return 'ogrn';
  }
  
  // Если не подходит ни под один формат, считаем названием
  return 'name';
}

// Временно отключаем логи фронтенда для просмотра логов бэкенда
const ENABLE_FRONTEND_LOGS = false;

// Включение/выключение отладочных логов для диагностики проблем с поиском
// Установите в true для включения подробного логирования
const ENABLE_DEBUG_LOGS = false;

// Хелпер для отправки отладочных логов (выполняется только если ENABLE_DEBUG_LOGS = true)
function debugLog(data: {
  sessionId?: string;
  runId?: string;
  hypothesisId?: string;
  location: string;
  message: string;
  data: unknown;
  timestamp?: number;
}) {
  if (!ENABLE_DEBUG_LOGS) return;
  
  fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      sessionId: "debug-session",
      runId: "debug-ogrn",
      ...data,
      timestamp: data.timestamp ?? Date.now(),
    }),
  }).catch(() => {});
}

export type SearchRow = {
  id: string;
  type: "company" | "entrepreneur";
  name: string;
  inn: string;
  region?: string | null;
  status?: string | null;
  registrationDate?: string | null;
};

function parseSearchParams(params: URLSearchParams): SearchFiltersInput & { applied?: boolean } {
  const raw = {
    q: params.get("q") ?? undefined,
    innOgrn: params.get("innOgrn") ?? undefined,
    region: params.get("region") ?? undefined,
    okved: params.get("okved") ?? undefined,
    // Если статус не передан в URL, не подставляем "all", чтобы не слать statusCode = "all" на бэкенд
    status: (params.get("status") as SearchFiltersInput["status"]) ?? undefined,
    founderName: params.get("founderName") ?? undefined,
    dateFrom: params.get("dateFrom") ?? undefined,
    dateTo: params.get("dateTo") ?? undefined,
    page: params.get("page") ?? "1",
    pageSize: params.get("pageSize") ?? "20",
    sortBy: (params.get("sortBy") ??
      undefined) as SearchFiltersInput["sortBy"],
    sortOrder: (params.get("sortOrder") as
      | SearchFiltersInput["sortOrder"]
      | null) ?? "asc",
    applied: params.get("applied") === "true",
  };

  return { ...searchFiltersSchema.parse(raw), applied: raw.applied };
}

function buildSearchParams(filters: SearchFiltersInput & { applied?: boolean }): URLSearchParams {
  const params = new URLSearchParams();

  if (filters.q) params.set("q", filters.q);
  if (filters.innOgrn) params.set("innOgrn", filters.innOgrn);
  if (filters.region) params.set("region", filters.region);
  if (filters.okved) params.set("okved", filters.okved);
  if (filters.status && filters.status !== "all")
    params.set("status", filters.status);
  if (filters.founderName) params.set("founderName", filters.founderName);
  if (filters.dateFrom) params.set("dateFrom", filters.dateFrom);
  if (filters.dateTo) params.set("dateTo", filters.dateTo);
  if (filters.page && filters.page !== 1)
    params.set("page", String(filters.page));
  if (filters.pageSize && filters.pageSize !== 20)
    params.set("pageSize", String(filters.pageSize));
  if (filters.sortBy) params.set("sortBy", filters.sortBy);
  if (filters.sortOrder && filters.sortOrder !== "asc")
    params.set("sortOrder", filters.sortOrder);
  if (filters.applied) params.set("applied", "true");

  return params;
}

export function useSearch() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  
  // Состояние для принудительного обновления при изменении URL
  const [urlKey, setUrlKey] = useState(0);

  // Отслеживаем изменения URL через useEffect
  useEffect(() => {
    // #region agent log: searchParams changed
    debugLog({
      runId: "run-filters",
      hypothesisId: "H4",
      location: "use-search.ts:useEffect",
      message: "searchParams changed",
      data: { 
        urlParams: Object.fromEntries(searchParams.entries()),
        searchParamsString: searchParams.toString(),
        currentUrl: typeof window !== "undefined" ? window.location.href : "SSR",
      },
    });
    // #endregion agent log: searchParams changed
    setUrlKey((prev: number) => prev + 1);
  }, [searchParams.toString()]);

  const filters = useMemo(() => {
    const parsed = parseSearchParams(searchParams);
    // #region agent log: parseSearchParams
    debugLog({
      runId: "run-filters",
      hypothesisId: "H4",
      location: "use-search.ts:parseSearchParams",
      message: "Parsed filters from URL",
      data: { 
        urlParams: Object.fromEntries(searchParams.entries()),
        parsedFilters: parsed,
        searchParamsString: searchParams.toString(),
        urlKey,
      },
    });
    // #endregion agent log: parseSearchParams
    return parsed;
  }, [searchParams, urlKey]);

  const debouncedQ = useDebounce(filters.q ?? "", 300);
  
  // Поиск активен только если фильтры применены (applied=true) или есть текстовый запрос >= 2 символа
  // Для текстового поиска оставляем автоматический режим, для расширенных фильтров - ручной
  const hasQuickSearch = debouncedQ.length >= 2;
  const hasAdvancedFilters =
    !!filters.region ||
    !!filters.okved ||
    !!filters.status ||
    !!filters.founderName ||
    !!filters.dateFrom ||
    !!filters.dateTo;
  
  // Поиск активен если: есть текстовый запрос (автоматически) ИЛИ применены расширенные фильтры
  const enabled = hasQuickSearch || (hasAdvancedFilters && filters.applied);

  // #region agent log: useSearch filters
  if (ENABLE_FRONTEND_LOGS) {
    fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        sessionId: "debug-session",
        runId: "run-filters",
        hypothesisId: "H1",
        location: "use-search.ts:filters",
        message: "Current search filters and enabled state",
        data: { filters, debouncedQ, hasQuickSearch, hasAdvancedFilters, enabled },
        timestamp: Date.now(),
      }),
    }).catch(() => {});
  }
  // #endregion agent log: useSearch filters

  const limit = filters.pageSize;
  const offset = (filters.page - 1) * limit;

  // Создаем отдельные объекты фильтров для компаний и предпринимателей
  // так как у них разные GraphQL схемы (CompanyFilter и EntrepreneurFilter)
  const filterParams = useMemo((): { companyFilter: Record<string, unknown>; entrepreneurFilter: Record<string, unknown> } => {
    const companyFilter: Record<string, unknown> = {};
    const entrepreneurFilter: Record<string, unknown> = {};
    
    // Определяем тип запроса из поля быстрого поиска
    if (debouncedQ) {
      const queryType = detectQueryType(debouncedQ);
      const trimmedQuery = debouncedQ.trim();
      
      if (queryType === 'inn') {
        // ИНН используется для обоих типов
        companyFilter.inn = trimmedQuery;
        entrepreneurFilter.inn = trimmedQuery;
      } else if (queryType === 'ogrn') {
        if (trimmedQuery.length === 13) {
          // ОГРН для юридических лиц (13 цифр) - только в фильтр компаний
          companyFilter.ogrn = trimmedQuery;
        } else if (trimmedQuery.length === 15) {
          // ОГРНИП для индивидуальных предпринимателей (15 цифр) - только в фильтр ИП
          entrepreneurFilter.ogrnip = trimmedQuery;
        } else {
          // Если формат не соответствует ожиданиям, пробуем в оба
          companyFilter.ogrn = trimmedQuery;
          entrepreneurFilter.ogrnip = trimmedQuery;
        }
      } else {
        // Если это не ИНН и не ОГРН, считаем названием
        companyFilter.name = debouncedQ;
        entrepreneurFilter.name = debouncedQ;
      }
    }
    
    // Дополнительный фильтр по ИНН/ОГРН из расширенного поиска
    if (filters.innOgrn) {
      companyFilter.inn = filters.innOgrn;
      entrepreneurFilter.inn = filters.innOgrn;
    }
    // Frontend отправляет код региона (например, "77", "78"), поэтому используем regionCode вместо region
    // Backend ищет по regionCode для точного совпадения, а по region - для поиска по названию через ILIKE
    if (filters.region) {
      companyFilter.regionCode = filters.region;
      entrepreneurFilter.regionCode = filters.region;
    }
    if (filters.okved) {
      companyFilter.okved = filters.okved;
      entrepreneurFilter.okved = filters.okved;
    }
    // Фильтрация по коду статуса (status_code) вместо текстового статуса
    if (filters.status) {
      companyFilter.statusCode = filters.status;
      entrepreneurFilter.statusCode = filters.status;
    }
    // Поиск по ФИО учредителя (только для ЮЛ)
    if (filters.founderName) {
      companyFilter.founderName = filters.founderName;
    }
    if (filters.dateFrom) {
      companyFilter.registrationDateFrom = filters.dateFrom;
      entrepreneurFilter.registrationDateFrom = filters.dateFrom;
    }
    if (filters.dateTo) {
      companyFilter.registrationDateTo = filters.dateTo;
      entrepreneurFilter.registrationDateTo = filters.dateTo;
    }
    
    // Создаем стабильный ключ для логирования
    const filterKey = JSON.stringify({ companyFilter, entrepreneurFilter });
    
    // #region agent log: filterParams computed
    debugLog({
      runId: "run-filters",
      hypothesisId: "H2",
      location: "use-search.ts:filterParams",
      message: "filterParams computed",
      data: { 
        companyFilter,
        entrepreneurFilter,
        filterKey,
        filters,
        debouncedQ,
        urlKey,
        enabled,
      },
    });
    // #endregion agent log: filterParams computed
    
    // Возвращаем объект с отдельными фильтрами для компаний и предпринимателей
    return { companyFilter, entrepreneurFilter };
  }, [debouncedQ, filters.innOgrn, filters.region, filters.okved, filters.status, filters.founderName, filters.dateFrom, filters.dateTo, urlKey, enabled]);

    // #region agent log: useSearch filterParams
    if (ENABLE_FRONTEND_LOGS) {
      fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          sessionId: "debug-session",
          runId: "run-filters",
          hypothesisId: "H2",
          location: "use-search.ts:filterParams",
          message: "Computed filter params and pagination",
          data: { filterParams, limit, offset },
          timestamp: Date.now(),
        }),
      }).catch(() => {});
    }
    // #endregion agent log: useSearch filterParams

  // Создаем стабильный ключ для запроса, чтобы React Query правильно видел изменения
  // Используем отдельные объекты фильтров для компаний и предпринимателей
  // Важно: отправляем undefined вместо пустого объекта, чтобы бэкенд правильно обработал отсутствие фильтра
  const companyQueryVariables = useMemo(() => {
    const companyFilter = filterParams.companyFilter;
    
    // #region agent log: before cleaning companyFilter
    debugLog({
      hypothesisId: "H1",
      location: "use-search.ts:companyQueryVariables:before-cleaning",
      message: "Before cleaning companyFilter",
      data: { 
        companyFilter,
        companyFilterKeys: Object.keys(companyFilter),
        companyFilterEntries: Object.entries(companyFilter),
        debouncedQ,
      },
    });
    // #endregion agent log: before cleaning companyFilter
    
    // Удаляем undefined значения и проверяем, есть ли хотя бы одно поле
    const cleanedCompanyFilter: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(companyFilter)) {
      if (value !== undefined && value !== null && value !== '') {
        cleanedCompanyFilter[key] = value;
      }
    }
    const hasCompanyFilter = Object.keys(cleanedCompanyFilter).length > 0;
    
    // #region agent log: after cleaning companyFilter
    debugLog({
      hypothesisId: "H3",
      location: "use-search.ts:companyQueryVariables:after-cleaning",
      message: "After cleaning companyFilter",
      data: { 
        cleanedCompanyFilter,
        cleanedCompanyFilterKeys: Object.keys(cleanedCompanyFilter),
        hasCompanyFilter,
      },
    });
    // #endregion agent log: after cleaning companyFilter
    
    const vars: { filter?: Record<string, unknown>; pagination: { limit: number; offset: number } } = {
      pagination: { limit, offset },
    };
    if (hasCompanyFilter) {
      vars.filter = cleanedCompanyFilter;
    }
    
    // #region agent log: vars before GraphQL request
    debugLog({
      hypothesisId: "H4",
      location: "use-search.ts:companyQueryVariables:vars-created",
      message: "Company query variables created",
      data: { 
        vars,
        varsHasFilter: 'filter' in vars,
        varsFilterValue: vars.filter,
        varsStringified: JSON.stringify(vars),
      },
    });
    // #endregion agent log: vars before GraphQL request
    
    // #region agent log: companyQueryVariables memoized
    debugLog({
      runId: "run-filters",
      hypothesisId: "H8",
      location: "use-search.ts:companyQueryVariables:memoized",
      message: "companyQueryVariables memoized",
      data: { 
        queryVariables: vars,
        companyFilter: filterParams.companyFilter,
        cleanedCompanyFilter,
        hasCompanyFilter,
        limit,
        offset,
        queryKey: JSON.stringify(vars),
      },
    });
    // #endregion agent log: companyQueryVariables memoized
    
    return vars;
  }, [filterParams, limit, offset]);
  
  const entrepreneurQueryVariables = useMemo(() => {
    const entrepreneurFilter = filterParams.entrepreneurFilter;
    
    // #region agent log: before cleaning entrepreneurFilter
    debugLog({
      hypothesisId: "H2",
      location: "use-search.ts:entrepreneurQueryVariables:before-cleaning",
      message: "Before cleaning entrepreneurFilter",
      data: { 
        entrepreneurFilter,
        entrepreneurFilterKeys: Object.keys(entrepreneurFilter),
        entrepreneurFilterEntries: Object.entries(entrepreneurFilter),
        debouncedQ,
      },
    });
    // #endregion agent log: before cleaning entrepreneurFilter
    
    // Удаляем undefined значения и проверяем, есть ли хотя бы одно поле
    const cleanedEntrepreneurFilter: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(entrepreneurFilter)) {
      if (value !== undefined && value !== null && value !== '') {
        cleanedEntrepreneurFilter[key] = value;
      }
    }
    const hasEntrepreneurFilter = Object.keys(cleanedEntrepreneurFilter).length > 0;
    
    // #region agent log: after cleaning entrepreneurFilter
    debugLog({
      hypothesisId: "H3",
      location: "use-search.ts:entrepreneurQueryVariables:after-cleaning",
      message: "After cleaning entrepreneurFilter",
      data: { 
        cleanedEntrepreneurFilter,
        cleanedEntrepreneurFilterKeys: Object.keys(cleanedEntrepreneurFilter),
        hasEntrepreneurFilter,
      },
    });
    // #endregion agent log: after cleaning entrepreneurFilter
    
    const vars: { filter?: Record<string, unknown>; pagination: { limit: number; offset: number } } = {
      pagination: { limit, offset },
    };
    if (hasEntrepreneurFilter) {
      vars.filter = cleanedEntrepreneurFilter;
    }
    
    // #region agent log: vars before GraphQL request entrepreneur
    debugLog({
      hypothesisId: "H4",
      location: "use-search.ts:entrepreneurQueryVariables:vars-created",
      message: "Entrepreneur query variables created",
      data: { 
        vars,
        varsHasFilter: 'filter' in vars,
        varsFilterValue: vars.filter,
        varsStringified: JSON.stringify(vars),
      },
    });
    // #endregion agent log: vars before GraphQL request entrepreneur
    
    // #region agent log: entrepreneurQueryVariables memoized
    debugLog({
      runId: "run-filters",
      hypothesisId: "H8",
      location: "use-search.ts:entrepreneurQueryVariables:memoized",
      message: "entrepreneurQueryVariables memoized",
      data: { 
        queryVariables: vars,
        entrepreneurFilter: filterParams.entrepreneurFilter,
        cleanedEntrepreneurFilter,
        hasEntrepreneurFilter,
        limit,
        offset,
        queryKey: JSON.stringify(vars),
      },
    });
    // #endregion agent log: entrepreneurQueryVariables memoized
    
    return vars;
  }, [filterParams, limit, offset]);
  
  // Для обратной совместимости оставляем queryVariables
  const queryVariables = companyQueryVariables;
    

  // #region agent log: before queries
  debugLog({
    runId: "run-filters",
    hypothesisId: "H3",
    location: "use-search.ts:before-queries",
    message: "About to call React Query hooks",
    data: { 
      queryVariables,
      enabled,
      queryKey: JSON.stringify(queryVariables),
      filters,
      hasQuickSearch,
      hasAdvancedFilters,
      filtersApplied: filters.applied,
    },
  });
  // #endregion agent log: before queries

  // Проверяем, есть ли фильтр для каждого типа запроса
  // Запрос должен выполняться только если есть хотя бы один фильтр в соответствующем объекте
  const hasCompanyFilterInVars = 'filter' in companyQueryVariables && companyQueryVariables.filter !== undefined;
  const hasEntrepreneurFilterInVars = 'filter' in entrepreneurQueryVariables && entrepreneurQueryVariables.filter !== undefined;
  
  // #region agent log: filter check before queries
  debugLog({
    hypothesisId: "H1",
    location: "use-search.ts:before-queries:filter-check",
    message: "Filter check before queries",
    data: {
      enabled,
      hasCompanyFilterInVars,
      hasEntrepreneurFilterInVars,
      companyQueryVariablesFilter: companyQueryVariables.filter,
      entrepreneurQueryVariablesFilter: entrepreneurQueryVariables.filter,
      companyEnabled: enabled && hasCompanyFilterInVars,
      entrepreneurEnabled: enabled && hasEntrepreneurFilterInVars,
    },
  });
  // #endregion agent log: filter check before queries
  
  const companiesQuery = useSearchCompaniesQuery(
    companyQueryVariables,
    {
      // Запрос выполняется только если enabled=true И есть фильтр
      enabled: enabled && hasCompanyFilterInVars,
      // Не используем placeholderData, чтобы при отсутствии фильтра не показывать старые данные
    }
  );

  const entrepreneursQuery = useSearchEntrepreneursQuery(
    entrepreneurQueryVariables,
    {
      // Запрос выполняется только если enabled=true И есть фильтр
      enabled: enabled && hasEntrepreneurFilterInVars,
      // Не используем placeholderData, чтобы при отсутствии фильтра не показывать старые данные
    }
  );

  // #region agent log: after queries
  if (ENABLE_FRONTEND_LOGS) {
    fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        sessionId: "debug-session",
        runId: "run-filters",
        hypothesisId: "H3",
        location: "use-search.ts:after-queries",
        message: "React Query hooks called",
        data: {
          companiesEnabled: companiesQuery.isEnabled,
          companiesFetching: companiesQuery.isFetching,
          companiesData: companiesQuery.data ? "has data" : "no data",
          entrepreneursEnabled: entrepreneursQuery.isEnabled,
          entrepreneursFetching: entrepreneursQuery.isFetching,
          entrepreneursData: entrepreneursQuery.data ? "has data" : "no data",
        },
        timestamp: Date.now(),
      }),
    }).catch(() => {});
  }
  // #endregion agent log: after queries

  const isLoading = companiesQuery.isLoading || entrepreneursQuery.isLoading;
  const isFetching = companiesQuery.isFetching || entrepreneursQuery.isFetching;
  const error = companiesQuery.error || entrepreneursQuery.error;

  // #region agent log: data access
  if (ENABLE_FRONTEND_LOGS) {
    fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        sessionId: "debug-session",
        runId: "run-filters",
        hypothesisId: "H6",
        location: "use-search.ts:data-access",
        message: "Accessing query data",
        data: {
          companiesData: companiesQuery.data ? "has data" : "no data",
          companiesDataValue: companiesQuery.data ? JSON.stringify(companiesQuery.data).substring(0, 100) : null,
          entrepreneursData: entrepreneursQuery.data ? "has data" : "no data",
          entrepreneursDataValue: entrepreneursQuery.data ? JSON.stringify(entrepreneursQuery.data).substring(0, 100) : null,
          enabled,
          companiesStatus: companiesQuery.status,
          entrepreneursStatus: entrepreneursQuery.status,
        },
        timestamp: Date.now(),
      }),
    }).catch(() => {});
  }
  // #endregion agent log: data access

  // Если фильтры не применены (applied: false), не показываем данные, даже если они есть в кэше
  // Это предотвращает показ данных для старых фильтров при изменении региона/других параметров
  const shouldShowData = enabled || filters.applied;

  // #region agent log: shouldShowData
  debugLog({
    runId: "run-filters",
    hypothesisId: "H7",
    location: "use-search.ts:shouldShowData",
    message: "shouldShowData computed",
    data: { 
      enabled,
      filtersApplied: filters.applied,
      shouldShowData,
      hasCompaniesData: !!companiesQuery.data,
      hasEntrepreneursData: !!entrepreneursQuery.data,
    },
  });
  // #endregion agent log: shouldShowData

  const totalCompanies = shouldShowData ? (companiesQuery.data?.companies.totalCount ?? 0) : 0;
  const totalEntrepreneurs = shouldShowData ? (entrepreneursQuery.data?.entrepreneurs.totalCount ?? 0) : 0;
  const total = totalCompanies + totalEntrepreneurs;

  const rows: SearchRow[] = useMemo(() => {
    // Если данные не должны показываться, возвращаем пустой массив
    if (!shouldShowData) {
      return [];
    }
    
    // Используем данные только если запрос был выполнен (enabled был true)
    // Если запрос отключен, не используем кэшированные данные
    const companyRows: SearchRow[] =
      (hasCompanyFilterInVars && companiesQuery.data?.companies.edges) 
        ? companiesQuery.data.companies.edges.map(
          ({ node }: { node: LegalEntity }) => ({
          id: node.ogrn,
          type: "company",
          name: node.fullName ?? node.shortName ?? "",
          inn: node.inn ?? "",
          region: undefined,
          status: node.status ?? null,
          registrationDate: node.registrationDate ?? null,
          })
        )
        : [];

    const entrepreneurRows: SearchRow[] =
      (hasEntrepreneurFilterInVars && entrepreneursQuery.data?.entrepreneurs.edges)
        ? entrepreneursQuery.data.entrepreneurs.edges.map(
          ({ node }: { node: IndividualEntrepreneur }) => ({
          id: node.ogrnip,
          type: "entrepreneur",
          name: `${node.lastName} ${node.firstName} ${node.middleName ?? ""}`.trim(),
          inn: node.inn ?? "",
          region: undefined,
          status: node.status ?? null,
          registrationDate: node.registrationDate ?? null,
          })
        )
        : [];

    const allRows = [...companyRows, ...entrepreneurRows];
    
    // #region agent log: rows computed
    debugLog({
      hypothesisId: "H7",
      location: "use-search.ts:rows",
      message: "rows computed",
      data: { 
        shouldShowData,
        hasCompanyFilterInVars,
        hasEntrepreneurFilterInVars,
        companyRowsCount: companyRows.length,
        entrepreneurRowsCount: entrepreneurRows.length,
        totalRowsCount: allRows.length,
        hasCompaniesData: !!companiesQuery.data,
        hasEntrepreneursData: !!entrepreneursQuery.data,
      },
    });
    // #endregion agent log: rows computed
    
    return allRows;
  }, [companiesQuery.data, entrepreneursQuery.data, shouldShowData, hasCompanyFilterInVars, hasEntrepreneurFilterInVars]);

  function updateFilters(partial: Partial<SearchFiltersInput & { applied?: boolean }>) {
    // #region agent log: updateFilters entry
    debugLog({
      runId: "run-filters",
      hypothesisId: "H1",
      location: "use-search.ts:updateFilters:entry",
      message: "updateFilters called with partial",
      data: { 
        partial,
        currentFilters: filters,
        partialKeys: Object.keys(partial),
      },
    });
    // #endregion agent log: updateFilters entry

    // Определяем, изменяются ли фильтры поиска (не пагинация/сортировка/applied)
    const isFilterChange = Object.keys(partial).some(
      (key) => !["page", "pageSize", "sortBy", "sortOrder", "applied"].includes(key)
    );

    // #region agent log: isFilterChange computed
    debugLog({
      runId: "run-filters",
      hypothesisId: "H1",
      location: "use-search.ts:updateFilters:isFilterChange",
      message: "isFilterChange computed",
      data: { 
        isFilterChange,
        partialKeys: Object.keys(partial),
        excludedKeys: ["page", "pageSize", "sortBy", "sortOrder", "applied"],
      },
    });
    // #endregion agent log: isFilterChange computed

    // Явно обрабатываем undefined значения - удаляем их из фильтров
    const cleanedPartial: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(partial)) {
      // Если значение undefined/null/пустая строка и это поле фильтра, удаляем его
      if (value === undefined || value === null || value === "") {
        if (
          [
            "region",
            "okved",
            "innOgrn",
            "dateFrom",
            "dateTo",
            "q",
            "status",
            "founderName",
          ].includes(key)
        ) {
          // помечаем поле как undefined, чтобы buildSearchParams не добавлял его в URL
          cleanedPartial[key] = undefined;
        }
      } else {
        cleanedPartial[key] = value;
      }
    }

    const next: SearchFiltersInput & { applied?: boolean } = {
      ...filters,
      ...(cleanedPartial as Partial<SearchFiltersInput & { applied?: boolean }>),
      // Сбрасываем страницу на 1 при изменении фильтров поиска
      page:
        (cleanedPartial.page as number | undefined) ??
        (isFilterChange ? 1 : filters.page),
      // При изменении фильтров (не пагинации) сбрасываем applied, если не указано явно
      applied: cleanedPartial.applied !== undefined 
        ? (cleanedPartial.applied as boolean)
        : (isFilterChange ? false : filters.applied),
    };

    // #region agent log: next filters computed
    debugLog({
      runId: "run-filters",
      hypothesisId: "H1",
      location: "use-search.ts:updateFilters:next",
      message: "Next filters computed",
      data: { 
        filtersBefore: filters,
        filtersAfter: next,
        appliedBefore: filters.applied,
        appliedAfter: next.applied,
        isFilterChange,
        cleanedPartial,
      },
    });
    // #endregion agent log: next filters computed

    const params = buildSearchParams(next);
    const search = params.toString();

    // Используем replace для обновления URL без добавления в историю
    const newUrl = search ? `${pathname}?${search}` : pathname;
    // #region agent log: router.replace
    debugLog({
      runId: "run-filters",
      hypothesisId: "H1",
      location: "use-search.ts:router.replace",
      message: "Router replace called",
      data: { 
        currentUrl: typeof window !== "undefined" ? window.location.href : "SSR",
        newUrl,
        searchParams: search,
        filtersBefore: filters,
        filtersAfter: next,
      },
    });
    // #endregion agent log: router.replace
    router.replace(newUrl);
    // Принудительно обновляем urlKey для пересчета фильтров
    setUrlKey((prev: number) => prev + 1);

    // #region agent log: updateFilters
    debugLog({
      runId: "run-filters",
      hypothesisId: "H3",
      location: "use-search.ts:updateFilters",
      message: "updateFilters called",
      data: { partial, next, newUrl, isFilterChange },
    });
    // #endregion agent log: updateFilters
  }

  function resetFilters() {
    // Полностью очищаем все фильтры, переходя на чистый URL без параметров
    router.replace(pathname);
    // Принудительно обновляем urlKey для пересчета фильтров
    setUrlKey((prev: number) => prev + 1);

    // #region agent log: resetFilters
    debugLog({
      runId: "run-filters",
      hypothesisId: "H4",
      location: "use-search.ts:resetFilters",
      message: "resetFilters called",
      data: { pathname, urlKey },
    });
    // #endregion agent log: resetFilters
  }

  // Функция для применения фильтров (вызывается по кнопке "Найти")
  function applyFilters() {
    updateFilters({ applied: true, page: 1 });
  }

  return {
    filters,
    debouncedQ,
    enabled,
    isLoading,
    isFetching,
    error,
    total,
    rows,
    page: filters.page,
    pageSize: filters.pageSize,
    updateFilters,
    resetFilters,
    applyFilters,
  };
}


