"use client";

import { useQuery } from "@tanstack/react-query";
import { defaultGraphQLClient } from "@/lib/api/graphql-client";
import type { Founder, HistoryRecord, RelatedCompany, Activity } from "@/lib/api";

// ==================== Типы ответов GraphQL ====================

interface GetCompanyFoundersResponse {
  company: {
    founders: Founder[];
  } | null;
}

interface GetCompanyHistoryResponse {
  company: {
    history: HistoryRecord[];
  } | null;
}

interface GetCompanyRelationsResponse {
  company: {
    relatedCompanies: RelatedCompany[];
  } | null;
}

interface GetCompanyActivitiesResponse {
  company: {
    activities: Activity[];
  } | null;
}

// ==================== Хуки для дополнительных данных ====================

export function useCompanyFoundersQuery(ogrn: string) {
  return useQuery<GetCompanyFoundersResponse, Error>({
    queryKey: ["company-founders", ogrn],
    queryFn: () =>
      defaultGraphQLClient.request<GetCompanyFoundersResponse, { ogrn: string }>(
        /* GraphQL */ `
          query GetCompanyFounders($ogrn: ID!) {
            company(ogrn: $ogrn) {
              founders {
                id
                type
                name
                inn
                share
                amount
                currency
              }
            }
          }
        `,
        { ogrn }
      ),
    enabled: !!ogrn,
  });
}

export function useCompanyHistoryQuery(ogrn: string) {
  return useQuery<GetCompanyHistoryResponse, Error>({
    queryKey: ["company-history", ogrn],
    queryFn: () =>
      defaultGraphQLClient.request<GetCompanyHistoryResponse, { ogrn: string }>(
        /* GraphQL */ `
          query GetCompanyHistory($ogrn: ID!) {
            company(ogrn: $ogrn) {
              history {
                id
                date
                type
                description
                details
              }
            }
          }
        `,
        { ogrn }
      ),
    enabled: !!ogrn,
  });
}

export function useCompanyRelationsQuery(ogrn: string) {
  return useQuery<GetCompanyRelationsResponse, Error>({
    queryKey: ["company-relations", ogrn],
    queryFn: () =>
      defaultGraphQLClient.request<GetCompanyRelationsResponse, { ogrn: string }>(
        /* GraphQL */ `
          query GetCompanyRelations($ogrn: ID!) {
            company(ogrn: $ogrn) {
              relatedCompanies {
                id
                ogrn
                name
                relationshipType
                status
              }
            }
          }
        `,
        { ogrn }
      ),
    enabled: !!ogrn,
  });
}

export function useCompanyActivitiesQuery(ogrn: string) {
  return useQuery<GetCompanyActivitiesResponse, Error>({
    queryKey: ["company-activities", ogrn],
    queryFn: () =>
      defaultGraphQLClient.request<GetCompanyActivitiesResponse, { ogrn: string }>(
        /* GraphQL */ `
          query GetCompanyActivities($ogrn: ID!) {
            company(ogrn: $ogrn) {
              activities {
                code
                name
              }
            }
          }
        `,
        { ogrn }
      ),
    enabled: !!ogrn,
  });
}