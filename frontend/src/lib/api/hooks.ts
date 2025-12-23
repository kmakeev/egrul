"use client";

import { useQuery, type UseQueryOptions } from "@tanstack/react-query";
import { defaultGraphQLClient } from "@/lib/api/graphql-client";
import type { LegalEntity, IndividualEntrepreneur } from "@/lib/api";

// Временно отключаем логи фронтенда для просмотра логов бэкенда
const ENABLE_FRONTEND_LOGS = true;

// ==================== Типы ответов GraphQL ====================

interface GetCompanyResponse {
  company: LegalEntity | null;
}

interface GetEntrepreneurResponse {
  entrepreneur: IndividualEntrepreneur | null;
}

export interface SearchCompaniesResponse {
  companies: {
    edges: {
      cursor: string;
      node: LegalEntity;
    }[];
    pageInfo: {
      hasNextPage: boolean;
      endCursor: string | null;
      totalCount: number;
    };
    totalCount: number;
  };
}

export interface SearchEntrepreneursResponse {
  entrepreneurs: {
    edges: {
      cursor: string;
      node: IndividualEntrepreneur;
    }[];
    pageInfo: {
      hasNextPage: boolean;
      endCursor: string | null;
      totalCount: number;
    };
    totalCount: number;
  };
}

interface GetStatisticsResponse {
  statistics: {
    totalCompanies: number;
    totalEntrepreneurs: number;
    activeCompanies: number;
    activeEntrepreneurs: number;
    liquidatedCompanies: number;
    liquidatedEntrepreneurs: number;
  };
}

// ==================== Хуки ====================

export function useCompanyQuery(ogrn: string) {
  return useQuery<GetCompanyResponse, Error>({
    queryKey: ["company", ogrn],
    queryFn: () =>
      defaultGraphQLClient.request<GetCompanyResponse, { ogrn: string }>(
        /* GraphQL */ `
          query GetCompany($ogrn: ID!) {
            company(ogrn: $ogrn) {
              ogrn
              ogrnDate
              inn
              kpp
              fullName
              shortName
              brandName
              legalForm
              status
              statusCode
              terminationMethod
              registrationDate
              terminationDate
              extractDate
              address {
                postalCode
                regionCode
                region
                district
                city
                locality
                street
                house
                building
                flat
                fullAddress
                fiasId
              }
              email
              capital {
                amount
                currency
              }
              director {
                lastName
                firstName
                middleName
                inn
                position
                positionCode
              }
              mainActivity {
                code
                name
                isMain
              }
              activities {
                code
                name
                isMain
              }
              regAuthority
              taxAuthority
              pfrRegNumber
              fssRegNumber
              founders {
                type
                ogrn
                inn
                name
                lastName
                firstName
                middleName
                country
                citizenship
                shareNominalValue
                sharePercent
              }
              foundersCount
              licensesCount
              branchesCount
              isBankrupt
              bankruptcyStage
              isLiquidating
              isReorganizing
              lastGrn
              lastGrnDate
              sourceFile
              versionDate
              createdAt
              updatedAt
            }
          }
        `,
        { ogrn }
      ),
  });
}

export function useEntrepreneurQuery(ogrnip: string) {
  return useQuery<GetEntrepreneurResponse, Error>({
    queryKey: ["entrepreneur", ogrnip],
    queryFn: () =>
      defaultGraphQLClient.request<GetEntrepreneurResponse, { ogrnip: string }>(
        /* GraphQL */ `
          query GetEntrepreneur($ogrnip: ID!) {
            entrepreneur(ogrnip: $ogrnip) {
              ogrnip
              inn
              lastName
              firstName
              middleName
              status
              registrationDate
            }
          }
        `,
        { ogrnip }
      ),
  });
}

export function useSearchCompaniesQuery<
  TData = SearchCompaniesResponse,
  TSelect = TData
>(
  variables: {
    filter?: Record<string, unknown>;
    pagination?: { limit?: number; offset?: number };
    sort?: Record<string, unknown>;
    searchType?: string;
  },
  options?: Omit<
    UseQueryOptions<
      SearchCompaniesResponse,
      Error,
      TSelect,
      ["search-companies", typeof variables]
    >,
    "queryKey" | "queryFn"
  >
) {
  // #region agent log: useQuery setup companies
  const queryKey: ["search-companies", typeof variables] = ["search-companies", variables];
  if (ENABLE_FRONTEND_LOGS) {
    fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      sessionId: "debug-session",
      runId: "run-filters",
      hypothesisId: "H8",
      location: "hooks.ts:useQuery:companies:setup",
      message: "useQuery setup for companies",
      data: { 
        queryKeyString: JSON.stringify(queryKey),
        variablesString: JSON.stringify(variables),
        filter: variables.filter,
        pagination: variables.pagination,
        searchType: variables.searchType,
      },
      timestamp: Date.now(),
    }),
  }).catch(() => {});
  }
  // #endregion agent log: useQuery setup companies

  return useQuery({
    ...(options as UseQueryOptions<
      SearchCompaniesResponse,
      Error,
      TSelect,
      ["search-companies", typeof variables]
    >),
    queryKey,
    queryFn: async () => {
      // #region agent log: GraphQL request companies
      if (ENABLE_FRONTEND_LOGS) {
        fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          sessionId: "debug-session",
          runId: "run-filters",
          hypothesisId: "H8",
          location: "hooks.ts:queryFn:companies",
          message: "Sending GraphQL request for companies",
          data: { 
            variables: JSON.stringify(variables),
            filter: variables.filter,
            pagination: variables.pagination,
          },
          timestamp: Date.now(),
        }),
      }).catch(() => {});
      }
      // #endregion agent log: GraphQL request companies
      
      // #region agent log: before GraphQL request companies
      fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          sessionId: "debug-session",
          runId: "debug-ogrn",
          hypothesisId: "H1",
          location: "hooks.ts:useSearchCompaniesQuery:before-request",
          message: "Before GraphQL request for companies",
          data: {
            variables,
            variablesStringified: JSON.stringify(variables),
            hasFilter: 'filter' in variables,
            filterValue: variables.filter,
          },
          timestamp: Date.now(),
        }),
      }).catch(() => {});
      // #endregion agent log: before GraphQL request companies
      
      const result = await defaultGraphQLClient.request<
        SearchCompaniesResponse,
        typeof variables
      >(
        /* GraphQL */ `
          query SearchCompanies($filter: CompanyFilter, $pagination: Pagination, $sort: CompanySort) {
            companies(filter: $filter, pagination: $pagination, sort: $sort) {
              edges {
                cursor
                node {
                  ogrn
                  inn
                  kpp
                  fullName
                  shortName
                  status
                  registrationDate
                  address {
                    regionCode
                  }
                }
              }
              pageInfo {
                hasNextPage
                endCursor
                totalCount
              }
              totalCount
            }
          }
        `,
        variables
      );
      
      // #region agent log: after GraphQL request companies
      fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          sessionId: "debug-session",
          runId: "debug-ogrn",
          hypothesisId: "H5",
          location: "hooks.ts:useSearchCompaniesQuery:after-request",
          message: "After GraphQL request for companies",
          data: {
            totalCount: result.companies?.totalCount ?? 0,
            edgesCount: result.companies?.edges?.length ?? 0,
            firstOgrn: result.companies?.edges?.[0]?.node?.ogrn ?? null,
            variablesFilter: variables.filter,
          },
          timestamp: Date.now(),
        }),
      }).catch(() => {});
      // #endregion agent log: after GraphQL request companies
      
      // #region agent log: GraphQL response companies
      if (ENABLE_FRONTEND_LOGS) {
        fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          sessionId: "debug-session",
          runId: "run-filters",
          hypothesisId: "H8",
          location: "hooks.ts:queryFn:companies:response",
          message: "Received GraphQL response for companies",
          data: { 
            totalCount: result.companies?.totalCount ?? 0,
            edgesCount: result.companies?.edges?.length ?? 0,
            firstOgrn: result.companies?.edges?.[0]?.node?.ogrn ?? null,
            firstInn: result.companies?.edges?.[0]?.node?.inn ?? null,
            filter: variables.filter,
          },
          timestamp: Date.now(),
        }),
      }).catch(() => {});
      }
      // #endregion agent log: GraphQL response companies
      
      return result;
    },
  });
}

export function useSearchEntrepreneursQuery<
  TData = SearchEntrepreneursResponse,
  TSelect = TData
>(
  variables: {
    filter?: Record<string, unknown>;
    pagination?: { limit?: number; offset?: number };
    sort?: Record<string, unknown>;
    searchType?: string;
  },
  options?: Omit<
    UseQueryOptions<
      SearchEntrepreneursResponse,
      Error,
      TSelect,
      ["search-entrepreneurs", typeof variables]
    >,
    "queryKey" | "queryFn"
  >
) {
  // #region agent log: useQuery setup entrepreneurs
  const queryKey: ["search-entrepreneurs", typeof variables] = ["search-entrepreneurs", variables];
  if (ENABLE_FRONTEND_LOGS) {
    fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      sessionId: "debug-session",
      runId: "run-filters",
      hypothesisId: "H8",
      location: "hooks.ts:useQuery:entrepreneurs:setup",
      message: "useQuery setup for entrepreneurs",
      data: { 
        queryKeyString: JSON.stringify(queryKey),
        variablesString: JSON.stringify(variables),
        filter: variables.filter,
        pagination: variables.pagination,
        searchType: variables.searchType,
      },
      timestamp: Date.now(),
    }),
  }).catch(() => {});
  }
  // #endregion agent log: useQuery setup entrepreneurs

  return useQuery({
    ...(options as UseQueryOptions<
      SearchEntrepreneursResponse,
      Error,
      TSelect,
      ["search-entrepreneurs", typeof variables]
    >),
    queryKey,
    queryFn: async () => {
      // #region agent log: GraphQL request entrepreneurs
      if (ENABLE_FRONTEND_LOGS) {
        fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          sessionId: "debug-session",
          runId: "run-filters",
          hypothesisId: "H8",
          location: "hooks.ts:queryFn:entrepreneurs",
          message: "Sending GraphQL request for entrepreneurs",
          data: { 
            variables: JSON.stringify(variables),
            filter: variables.filter,
            pagination: variables.pagination,
          },
          timestamp: Date.now(),
        }),
      }).catch(() => {});
      }
      // #endregion agent log: GraphQL request entrepreneurs
      
      // #region agent log: before GraphQL request entrepreneurs
      fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          sessionId: "debug-session",
          runId: "debug-ogrn",
          hypothesisId: "H2",
          location: "hooks.ts:useSearchEntrepreneursQuery:before-request",
          message: "Before GraphQL request for entrepreneurs",
          data: {
            variables,
            variablesStringified: JSON.stringify(variables),
            hasFilter: 'filter' in variables,
            filterValue: variables.filter,
          },
          timestamp: Date.now(),
        }),
      }).catch(() => {});
      // #endregion agent log: before GraphQL request entrepreneurs
      
      const result = await defaultGraphQLClient.request<
        SearchEntrepreneursResponse,
        typeof variables
      >(
        /* GraphQL */ `
          query SearchEntrepreneurs($filter: EntrepreneurFilter, $pagination: Pagination, $sort: EntrepreneurSort) {
            entrepreneurs(filter: $filter, pagination: $pagination, sort: $sort) {
              edges {
                cursor
                node {
                  ogrnip
                  inn
                  lastName
                  firstName
                  middleName
                  status
                  registrationDate
                  address {
                    regionCode
                  }
                }
              }
              pageInfo {
                hasNextPage
                endCursor
                totalCount
              }
              totalCount
            }
          }
        `,
        variables
      );
      
      // #region agent log: after GraphQL request entrepreneurs
      fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          sessionId: "debug-session",
          runId: "debug-ogrn",
          hypothesisId: "H5",
          location: "hooks.ts:useSearchEntrepreneursQuery:after-request",
          message: "After GraphQL request for entrepreneurs",
          data: {
            totalCount: result.entrepreneurs?.totalCount ?? 0,
            edgesCount: result.entrepreneurs?.edges?.length ?? 0,
            firstOgrnip: result.entrepreneurs?.edges?.[0]?.node?.ogrnip ?? null,
            variablesFilter: variables.filter,
          },
          timestamp: Date.now(),
        }),
      }).catch(() => {});
      // #endregion agent log: after GraphQL request entrepreneurs
      
      // #region agent log: GraphQL response entrepreneurs
      if (ENABLE_FRONTEND_LOGS) {
        fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          sessionId: "debug-session",
          runId: "run-filters",
          hypothesisId: "H8",
          location: "hooks.ts:queryFn:entrepreneurs:response",
          message: "Received GraphQL response for entrepreneurs",
          data: { 
            totalCount: result.entrepreneurs?.totalCount ?? 0,
            edgesCount: result.entrepreneurs?.edges?.length ?? 0,
            firstOgrnip: result.entrepreneurs?.edges?.[0]?.node?.ogrnip ?? null,
            firstInn: result.entrepreneurs?.edges?.[0]?.node?.inn ?? null,
            filter: variables.filter,
          },
          timestamp: Date.now(),
        }),
      }).catch(() => {});
      }
      // #endregion agent log: GraphQL response entrepreneurs
      
      return result;
    },
  });
}

export function useStatisticsQuery<
  TData = GetStatisticsResponse,
  TSelect = TData
>(
  variables: { filter?: Record<string, unknown> },
  options?: Omit<
    UseQueryOptions<
      GetStatisticsResponse,
      Error,
      TSelect,
      ["statistics", typeof variables]
    >,
    "queryKey" | "queryFn"
  >
) {
  return useQuery({
    ...(options as UseQueryOptions<
      GetStatisticsResponse,
      Error,
      TSelect,
      ["statistics", typeof variables]
    >),
    queryKey: ["statistics", variables],
    queryFn: () =>
      defaultGraphQLClient.request<
        GetStatisticsResponse,
        typeof variables
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
            }
          }
        `,
        variables
      ),
  });
}