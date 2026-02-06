import { useMemo, useEffect, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useDebounce } from "@/hooks/use-debounce";
import {
  useSearchCompaniesQuery,
  useSearchEntrepreneursQuery,
} from "@/lib/api/hooks";
import type { LegalEntity, IndividualEntrepreneur } from "@/lib/api";
import { searchFiltersSchema, type SearchFiltersInput } from "@/lib/validations";
import { getRegionNameByCode } from "@/lib/regions";

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



export type SearchRow = {
  id: string;
  type: "company" | "entrepreneur";
  name: string;
  inn: string;
  ogrn?: string | null; // ОГРН для компаний или ОГРНИП для ИП
  region?: string | null;
  registrationDate?: string | null;
};

function parseSearchParams(params: URLSearchParams): SearchFiltersInput {
  const raw = {
    q: params.get("q") ?? undefined,
    innOgrn: params.get("innOgrn") ?? undefined,
    region: params.get("region") ?? undefined,
    okved: params.get("okved") ?? undefined,
    // Если статус не передан в URL, не подставляем "all", чтобы не слать statusCode = "all" на бэкенд
    status: (params.get("status") as SearchFiltersInput["status"]) ?? undefined,
    founderName: params.get("founderName") ?? undefined,
    entityType: (params.get("entityType") as SearchFiltersInput["entityType"]) ?? "all",
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

  return searchFiltersSchema.parse(raw);
}

function buildSearchParams(filters: SearchFiltersInput): URLSearchParams {
  const params = new URLSearchParams();

  if (filters.q) params.set("q", filters.q);
  if (filters.innOgrn) params.set("innOgrn", filters.innOgrn);
  if (filters.region) params.set("region", filters.region);
  if (filters.okved) params.set("okved", filters.okved);
  if (filters.status && filters.status !== "all")
    params.set("status", filters.status);
  if (filters.founderName) params.set("founderName", filters.founderName);
  if (filters.entityType && filters.entityType !== "all")
    params.set("entityType", filters.entityType);
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
    setUrlKey((prev: number) => prev + 1);
  }, [searchParams]);

  const filters = useMemo(() => {
    const parsed = parseSearchParams(searchParams);
    return parsed;
  }, [searchParams, urlKey]);

  const debouncedQ = useDebounce(filters.q ?? "", 1000);
  
  // Поиск активен только если фильтры применены (applied=true) или есть текстовый запрос >= 2 символа
  // Для текстового поиска оставляем автоматический режим, для расширенных фильтров - ручной
  const hasQuickSearch = debouncedQ.length >= 2;
  const _hasAdvancedFilters =
    !!filters.region ||
    !!filters.okved ||
    !!filters.status ||
    !!filters.founderName ||
    !!filters.dateFrom ||
    !!filters.dateTo ||
    (filters.entityType && filters.entityType !== "all");
  
  // Поиск активен если: есть текстовый запрос (автоматически) ИЛИ применены любые фильтры (включая пустые)
  const enabled = hasQuickSearch || filters.applied;


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
      const innOgrnType = detectQueryType(filters.innOgrn);
      const trimmedInnOgrn = filters.innOgrn.trim();
      
      if (innOgrnType === 'inn') {
        // ИНН используется для обоих типов
        companyFilter.inn = trimmedInnOgrn;
        entrepreneurFilter.inn = trimmedInnOgrn;
      } else if (innOgrnType === 'ogrn') {
        if (trimmedInnOgrn.length === 13) {
          // ОГРН для юридических лиц (13 цифр) - только в фильтр компаний
          companyFilter.ogrn = trimmedInnOgrn;
        } else if (trimmedInnOgrn.length === 15) {
          // ОГРНИП для индивидуальных предпринимателей (15 цифр) - только в фильтр ИП
          entrepreneurFilter.ogrnip = trimmedInnOgrn;
        } else {
          // Если формат не соответствует ожиданиям, пробуем в оба
          companyFilter.ogrn = trimmedInnOgrn;
          entrepreneurFilter.ogrnip = trimmedInnOgrn;
        }
      } else {
        // Если это не ИНН и не ОГРН, считаем ИНН (для обратной совместимости)
        companyFilter.inn = trimmedInnOgrn;
        entrepreneurFilter.inn = trimmedInnOgrn;
      }
    }
    // ИСПРАВЛЕНИЕ: Логика добавления дополнительных фильтров
    // Если есть основное поле поиска (q или innOgrn), то дополнительные фильтры добавляются 
    // только к тем фильтрам, которые имеют основной идентификатор.
    // Если основного поля нет (только расширенные параметры), то дополнительные фильтры 
    // добавляются в оба фильтра для широкого поиска.
    // НОВОЕ: Учитываем фильтр по типу организации (entityType)
    
    const hasMainSearchField = !!(debouncedQ || filters.innOgrn);
    const hasCompanyMainFilter = !!(companyFilter.ogrn || companyFilter.inn || companyFilter.name);
    const hasEntrepreneurMainFilter = !!(entrepreneurFilter.ogrnip || entrepreneurFilter.inn || entrepreneurFilter.name);
    
    // Определяем, какие типы организаций нужно искать
    const shouldSearchCompanies = filters.entityType === "all" || filters.entityType === "company";
    const shouldSearchEntrepreneurs = filters.entityType === "all" || filters.entityType === "entrepreneur";
    
    // Frontend отправляет код региона (например, "77", "78"), поэтому используем regionCode вместо region
    // Backend ищет по regionCode для точного совпадения, а по region - для поиска по названию через ILIKE
    if (filters.region) {
      if (shouldSearchCompanies && (!hasMainSearchField || hasCompanyMainFilter)) {
        companyFilter.regionCode = filters.region;
      }
      if (shouldSearchEntrepreneurs && (!hasMainSearchField || hasEntrepreneurMainFilter)) {
        entrepreneurFilter.regionCode = filters.region;
      }
    }
    if (filters.okved) {
      if (shouldSearchCompanies && (!hasMainSearchField || hasCompanyMainFilter)) {
        companyFilter.okved = filters.okved;
      }
      if (shouldSearchEntrepreneurs && (!hasMainSearchField || hasEntrepreneurMainFilter)) {
        entrepreneurFilter.okved = filters.okved;
      }
    }
    // Фильтрация по коду статуса (status_code) вместо текстового статуса
    if (filters.status) {
      if (shouldSearchCompanies && (!hasMainSearchField || hasCompanyMainFilter)) {
        companyFilter.statusCode = filters.status;
      }
      if (shouldSearchEntrepreneurs && (!hasMainSearchField || hasEntrepreneurMainFilter)) {
        entrepreneurFilter.statusCode = filters.status;
      }
    }
    // Поиск по ФИО учредителя (только для ЮЛ)
    if (filters.founderName) {
      if (shouldSearchCompanies && (!hasMainSearchField || hasCompanyMainFilter)) {
        companyFilter.founderName = filters.founderName;
      }
    }
    if (filters.dateFrom) {
      if (shouldSearchCompanies && (!hasMainSearchField || hasCompanyMainFilter)) {
        companyFilter.registrationDateFrom = filters.dateFrom;
      }
      if (shouldSearchEntrepreneurs && (!hasMainSearchField || hasEntrepreneurMainFilter)) {
        entrepreneurFilter.registrationDateFrom = filters.dateFrom;
      }
    }
    if (filters.dateTo) {
      if (shouldSearchCompanies && (!hasMainSearchField || hasCompanyMainFilter)) {
        companyFilter.registrationDateTo = filters.dateTo;
      }
      if (shouldSearchEntrepreneurs && (!hasMainSearchField || hasEntrepreneurMainFilter)) {
        entrepreneurFilter.registrationDateTo = filters.dateTo;
      }
    }
    
    // Если выбран конкретный тип организации, очищаем фильтр для другого типа
    if (!shouldSearchCompanies) {
      Object.keys(companyFilter).forEach(key => delete companyFilter[key]);
    }
    if (!shouldSearchEntrepreneurs) {
      Object.keys(entrepreneurFilter).forEach(key => delete entrepreneurFilter[key]);
    }

    // Возвращаем объект с отдельными фильтрами для компаний и предпринимателей
    return { companyFilter, entrepreneurFilter };
  }, [debouncedQ, filters, urlKey, enabled]);


  // Создаем стабильный ключ для запроса, чтобы React Query правильно видел изменения
  // Используем отдельные объекты фильтров для компаний и предпринимателей
  // Важно: отправляем undefined вместо пустого объекта, чтобы бэкенд правильно обработал отсутствие фильтра
  const companyQueryVariables = useMemo(() => {
    const companyFilter = filterParams.companyFilter;
    
    
    // Удаляем undefined значения и проверяем, есть ли хотя бы одно поле
    const cleanedCompanyFilter: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(companyFilter)) {
      if (value !== undefined && value !== null && value !== '') {
        cleanedCompanyFilter[key] = value;
      }
    }
    const hasCompanyFilter = Object.keys(cleanedCompanyFilter).length > 0;
    
    
    // ИСПРАВЛЕНИЕ: Создаем уникальный ключ кеширования на основе всех параметров фильтра
    const cacheKey = JSON.stringify({
      filter: cleanedCompanyFilter,
      entityType: filters.entityType,
      applied: filters.applied,
      urlKey, // Добавляем urlKey для принудительного обновления кеша при изменении URL
    });
    
    const vars: { 
      filter?: Record<string, unknown>; 
      pagination: { limit: number; offset: number };
      sort?: Record<string, unknown>;
      cacheKey: string;
    } = {
      pagination: { limit, offset },
      cacheKey,
    };
    if (hasCompanyFilter) {
      vars.filter = cleanedCompanyFilter;
    }
    
    // Добавляем сортировку если указана
    if (filters.sortBy && filters.sortOrder) {
      // Маппинг полей сортировки фронтенда на GraphQL поля
      const sortFieldMap: Record<string, string> = {
        'name': 'FULL_NAME',
        'inn': 'INN', 
        'ogrn': 'OGRN',
        'registrationDate': 'REGISTRATION_DATE'
      };
      
      const graphqlField = sortFieldMap[filters.sortBy];
      if (graphqlField) {
        vars.sort = {
          field: graphqlField,
          order: filters.sortOrder.toUpperCase()
        };
        
      }
    }
    
    
    
    return vars;
  }, [filterParams, limit, offset, filters.sortBy, filters.sortOrder, debouncedQ, urlKey, filters.applied, filters.entityType]);
  
  const entrepreneurQueryVariables = useMemo(() => {
    const entrepreneurFilter = filterParams.entrepreneurFilter;
    
    
    // Удаляем undefined значения и проверяем, есть ли хотя бы одно поле
    const cleanedEntrepreneurFilter: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(entrepreneurFilter)) {
      if (value !== undefined && value !== null && value !== '') {
        cleanedEntrepreneurFilter[key] = value;
      }
    }
    const hasEntrepreneurFilter = Object.keys(cleanedEntrepreneurFilter).length > 0;
    
    
    // ИСПРАВЛЕНИЕ: Создаем уникальный ключ кеширования на основе всех параметров фильтра
    const cacheKey = JSON.stringify({
      filter: cleanedEntrepreneurFilter,
      entityType: filters.entityType,
      applied: filters.applied,
      urlKey, // Добавляем urlKey для принудительного обновления кеша при изменении URL
    });
    
    const vars: { 
      filter?: Record<string, unknown>; 
      pagination: { limit: number; offset: number };
      sort?: Record<string, unknown>;
      cacheKey: string;
    } = {
      pagination: { limit, offset },
      cacheKey,
    };
    if (hasEntrepreneurFilter) {
      vars.filter = cleanedEntrepreneurFilter;
    }
    
    // Добавляем сортировку если указана
    if (filters.sortBy && filters.sortOrder) {
      // Маппинг полей сортировки фронтенда на GraphQL поля для предпринимателей
      const sortFieldMap: Record<string, string> = {
        'name': 'FULL_NAME',
        'inn': 'INN', 
        'ogrn': 'OGRNIP', // Для предпринимателей используем OGRNIP
        'registrationDate': 'REGISTRATION_DATE'
      };
      
      const graphqlField = sortFieldMap[filters.sortBy];
      if (graphqlField) {
        vars.sort = {
          field: graphqlField,
          order: filters.sortOrder.toUpperCase()
        };
        
      }
    }
    
    
    
    return vars;
  }, [filterParams, limit, offset, filters.sortBy, filters.sortOrder, debouncedQ, urlKey, filters.applied, filters.entityType]);

  // Проверяем, нужно ли выполнять запросы для каждого типа
  // Запрос выполняется если: enabled=true И (есть фильтр ИЛИ выбран соответствующий тип организации)
  const shouldSearchCompanies = filters.entityType === "all" || filters.entityType === "company";
  const shouldSearchEntrepreneurs = filters.entityType === "all" || filters.entityType === "entrepreneur";
  
  const hasCompanyFilterInVars = 'filter' in companyQueryVariables && companyQueryVariables.filter !== undefined;
  const hasEntrepreneurFilterInVars = 'filter' in entrepreneurQueryVariables && entrepreneurQueryVariables.filter !== undefined;
  
  // Запросы выполняются если enabled=true И нужно искать соответствующий тип И есть соответствующий фильтр
  // Исключение: если нет основного поискового поля (только расширенные фильтры), то выполняем оба запроса
  const hasMainSearchField = !!(debouncedQ || filters.innOgrn);
  const companyQueryEnabled = enabled && shouldSearchCompanies && (hasCompanyFilterInVars || !hasMainSearchField);
  const entrepreneurQueryEnabled = enabled && shouldSearchEntrepreneurs && (hasEntrepreneurFilterInVars || !hasMainSearchField);
  
  
  const companiesQuery = useSearchCompaniesQuery(
    companyQueryVariables,
    {
      // Запрос выполняется если enabled=true И нужно искать компании
      enabled: companyQueryEnabled,
      // Не используем placeholderData, чтобы при отсутствии фильтра не показывать старые данные
      // ИСПРАВЛЕНИЕ: Отключаем кеширование для предотвращения показа данных от предыдущих поисков
      staleTime: 0,
      gcTime: 0,
    }
  );

  const entrepreneursQuery = useSearchEntrepreneursQuery(
    entrepreneurQueryVariables,
    {
      // Запрос выполняется если enabled=true И нужно искать предпринимателей
      enabled: entrepreneurQueryEnabled,
      // Не используем placeholderData, чтобы при отсутствии фильтра не показывать старые данные
      // ИСПРАВЛЕНИЕ: Отключаем кеширование для предотвращения показа данных от предыдущих поисков
      staleTime: 0,
      gcTime: 0,
    }
  );


  const isLoading = companiesQuery.isLoading || entrepreneursQuery.isLoading;
  const isFetching = companiesQuery.isFetching || entrepreneursQuery.isFetching;
  const error = companiesQuery.error || entrepreneursQuery.error;


  // Если фильтры не применены (applied: false), не показываем данные, даже если они есть в кэше
  // Это предотвращает показ данных для старых фильтров при изменении региона/других параметров
  const shouldShowData = enabled || filters.applied;


  const totalCompanies = shouldShowData && companyQueryEnabled ? (companiesQuery.data?.companies.pageInfo.totalCount ?? 0) : 0;
  const totalEntrepreneurs = shouldShowData && entrepreneurQueryEnabled ? (entrepreneursQuery.data?.entrepreneurs.pageInfo.totalCount ?? 0) : 0;
  const total = totalCompanies + totalEntrepreneurs;

  const rows: SearchRow[] = useMemo(() => {
    // Если данные не должны показываться, возвращаем пустой массив
    if (!shouldShowData) {
      return [];
    }
    
    // ИСПРАВЛЕНИЕ: Используем данные только если соответствующий запрос активен И выполнен
    // Это предотвращает показ кешированных данных от предыдущих поисков
    const companyRows: SearchRow[] =
      (companyQueryEnabled && companiesQuery.data?.companies.edges) 
        ? companiesQuery.data.companies.edges.map(
          ({ node }: { node: LegalEntity }) => ({
          id: node.ogrn,
          type: "company",
          name: node.fullName ?? node.shortName ?? "",
          inn: node.inn ?? "",
          ogrn: node.ogrn ?? null,
          region: node.address?.regionCode ? getRegionNameByCode(node.address.regionCode) : null,
          registrationDate: node.registrationDate ?? null,
          })
        )
        : [];

    const entrepreneurRows: SearchRow[] =
      (entrepreneurQueryEnabled && entrepreneursQuery.data?.entrepreneurs.edges)
        ? entrepreneursQuery.data.entrepreneurs.edges.map(
          ({ node }: { node: IndividualEntrepreneur }) => ({
          id: node.ogrnip,
          type: "entrepreneur",
          name: `${node.lastName} ${node.firstName} ${node.middleName ?? ""}`.trim(),
          inn: node.inn ?? "",
          ogrn: node.ogrnip ?? null,
          region: node.address?.regionCode ? getRegionNameByCode(node.address.regionCode) : null,
          registrationDate: node.registrationDate ?? null,
          })
        )
        : [];

    const allRows = [...companyRows, ...entrepreneurRows];
    
    
    return allRows;
  }, [companiesQuery.data, entrepreneursQuery.data, shouldShowData, companyQueryEnabled, entrepreneurQueryEnabled, enabled, hasCompanyFilterInVars, hasEntrepreneurFilterInVars]);

  function updateFilters(partial: Partial<SearchFiltersInput>) {

    // Определяем, изменяются ли фильтры поиска (не пагинация/сортировка/applied)
    const isFilterChange = Object.keys(partial).some(
      (key) => !["page", "pageSize", "sortBy", "sortOrder", "applied"].includes(key)
    );


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
            "entityType",
          ].includes(key)
        ) {
          // помечаем поле как undefined, чтобы buildSearchParams не добавлял его в URL
          cleanedPartial[key] = undefined;
        }
      } else {
        cleanedPartial[key] = value;
      }
    }

    const next: SearchFiltersInput = {
      ...filters,
      ...(cleanedPartial as Partial<SearchFiltersInput>),
      // Сбрасываем страницу на 1 при изменении фильтров поиска
      page:
        (cleanedPartial.page as number | undefined) ??
        (isFilterChange ? 1 : filters.page),
      // При изменении фильтров (не пагинации) сбрасываем applied, если не указано явно
      applied: cleanedPartial.applied !== undefined 
        ? (cleanedPartial.applied as boolean)
        : (isFilterChange ? false : filters.applied),
    };


    const params = buildSearchParams(next);
    const search = params.toString();

    // Используем replace для обновления URL без добавления в историю
    const newUrl = search ? `${pathname}?${search}` : pathname;
    router.replace(newUrl, { scroll: false });
    // Принудительно обновляем urlKey для пересчета фильтров
    setUrlKey((prev: number) => prev + 1);

  }

  function resetFilters() {
    // Полностью очищаем все фильтры, переходя на чистый URL без параметров
    router.replace(pathname, { scroll: false });
    // Принудительно обновляем urlKey для пересчета фильтров
    setUrlKey((prev: number) => prev + 1);

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


