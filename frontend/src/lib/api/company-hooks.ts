"use client";

import { useQuery } from "@tanstack/react-query";
import { defaultGraphQLClient } from "@/lib/api/graphql-client";
import type { HistoryRecord, RelatedCompany, Activity } from "@/lib/api";

// ==================== Типы ответов GraphQL ====================

export interface GetCompanyHistoryResponse {
  company: {
    ogrn: string;
    history: HistoryRecord[];
    historyCount?: number;
    // Добавляем все остальные поля, которые может вернуть GraphQL
    [key: string]: any;
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

export function useCompanyHistoryQuery(ogrn: string, limit: number = 50, offset: number = 0, options?: { enabled?: boolean }) {
  return useQuery<GetCompanyHistoryResponse>({
    queryKey: ["company-history", ogrn, limit, offset],
    queryFn: () =>
      defaultGraphQLClient.request<GetCompanyHistoryResponse, { ogrn: string; limit: number; offset: number }>(
        /* GraphQL */ `
          query GetCompanyHistory($ogrn: ID!, $limit: Int!, $offset: Int!) {
            company(ogrn: $ogrn) {
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
        { ogrn, limit, offset }
      ),
    enabled: options?.enabled !== false && !!ogrn,
    staleTime: 0, // Данные сразу считаются устаревшими - как в поиске
    gcTime: 0, // Не кешируем данные - как в поиске
  });
}

// Новый хук для истории через отдельный запрос (обходит проблемы с резолверами)
export function useCompanyHistoryDirectQuery(ogrn: string, limit: number = 50, offset: number = 0, options?: { enabled?: boolean }) {
  console.log("🚀 CALLING NEW DIRECT HISTORY QUERY", { ogrn, limit, offset });
  
  return useQuery<{ entityHistory: HistoryRecord[] }>({
    queryKey: ["entity-history-direct", ogrn, limit, offset],
    queryFn: () =>
      defaultGraphQLClient.request<{ entityHistory: HistoryRecord[] }, { entityType: string; entityId: string; limit: number; offset: number }>(
        /* GraphQL */ `
          query GetEntityHistoryDirect($entityType: EntityType!, $entityId: ID!, $limit: Int!, $offset: Int!) {
            entityHistory(entityType: $entityType, entityId: $entityId, limit: $limit, offset: $offset) {
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
        `,
        { entityType: "COMPANY", entityId: ogrn, limit, offset }
      ),
    enabled: options?.enabled !== false && !!ogrn,
    staleTime: 0, // Данные сразу считаются устаревшими - как в поиске
    gcTime: 0, // Не кешируем данные - как в поиске
    retry: false, // Не повторяем запросы при ошибках
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