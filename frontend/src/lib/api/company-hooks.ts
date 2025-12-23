"use client";

import { useQuery } from "@tanstack/react-query";
import { defaultGraphQLClient } from "@/lib/api/graphql-client";
import type { HistoryRecord, RelatedCompany, Activity } from "@/lib/api";

// ==================== Типы ответов GraphQL ====================

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

export function useCompanyHistoryQuery(ogrn: string) {
  return useQuery<GetCompanyHistoryResponse, Error>({
    queryKey: ["company-history", ogrn],
    queryFn: () =>
      defaultGraphQLClient.request<GetCompanyHistoryResponse, { ogrn: string }>(
        /* GraphQL */ `
          query GetCompanyHistory($ogrn: ID!) {
            company(ogrn: $ogrn) {
              history(limit: 100, offset: 0) {
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