"use client";

import { useQuery, type UseQueryOptions } from "@tanstack/react-query";
import { defaultGraphQLClient } from "@/lib/api/graphql-client";
import type { LegalEntity, IndividualEntrepreneur } from "@/lib/api";


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
              history(limit: 50, offset: 0) {
                id
                grn
                date
                reasonCode
                reasonDescription
                authority {
                  code
                  name
                }
                certificateSeries
                certificateNumber
                certificateDate
                snapshotFullName
                snapshotStatus
                snapshotAddress
              }
              historyCount
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
              id
              ogrnip
              ogrnipDate
              inn
              lastName
              firstName
              middleName
              registrationDate
              terminationDate
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
                kladrCode
              }
              status
              statusCode
              citizenshipType
              citizenshipCountryCode
              citizenshipCountryName
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
              regAuthority {
                code
                name
              }
              taxAuthority {
                code
                name
              }
              history(limit: 50, offset: 0) {
                id
                grn
                date
                reasonCode
                reasonDescription
                authority {
                  code
                  name
                }
                certificateSeries
                certificateNumber
                certificateDate
                snapshotFullName
                snapshotStatus
                snapshotAddress
              }
              historyCount
              licensesCount
              extractDate
              lastGrn
              lastGrnDate
              sourceFile
              versionDate
              createdAt
              updatedAt
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

  return useQuery({
    ...(options as UseQueryOptions<
      SearchCompaniesResponse,
      Error,
      TSelect,
      ["search-companies", typeof variables]
    >),
    queryKey,
    queryFn: async () => {
      
      
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

  return useQuery({
    ...(options as UseQueryOptions<
      SearchEntrepreneursResponse,
      Error,
      TSelect,
      ["search-entrepreneurs", typeof variables]
    >),
    queryKey,
    queryFn: async () => {
      
      
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

// ==================== Хуки для ИП ====================

interface GetEntrepreneurHistoryResponse {
  entrepreneur: {
    history: Array<{
      id: string;
      grn: string;
      date: string | null;
      reasonCode?: string | null;
      reasonDescription?: string | null;
      authority?: {
        code?: string | null;
        name?: string | null;
      } | null;
      certificateSeries?: string | null;
      certificateNumber?: string | null;
      certificateDate?: string | null;
      snapshotFullName?: string | null;
      snapshotStatus?: string | null;
      snapshotAddress?: string | null;
    }>;
    historyCount: number;
  } | null;
}

export function useEntrepreneurHistoryQuery(
  ogrnip: string,
  limit: number = 20,
  offset: number = 0,
  options?: { enabled?: boolean }
) {
  return useQuery({
    queryKey: ["entrepreneur-history", ogrnip, limit, offset],
    queryFn: () =>
      defaultGraphQLClient.request<GetEntrepreneurHistoryResponse, { ogrnip: string; limit: number; offset: number }>(
        /* GraphQL */ `
          query GetEntrepreneurHistory($ogrnip: ID!, $limit: Int!, $offset: Int!) {
            entrepreneur(ogrnip: $ogrnip) {
              history(limit: $limit, offset: $offset) {
                id
                grn
                date
                reasonCode
                reasonDescription
                authority {
                  code
                  name
                }
                certificateSeries
                certificateNumber
                certificateDate
                snapshotFullName
                snapshotStatus
                snapshotAddress
              }
              historyCount
            }
          }
        `,
        { ogrnip, limit, offset }
      ),
    ...options,
  });
}