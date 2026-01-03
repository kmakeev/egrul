"use client";

import { useQuery, type UseQueryOptions } from "@tanstack/react-query";
import { defaultGraphQLClient } from "@/lib/api/graphql-client";
import type { IndividualEntrepreneur } from "@/lib/api";

// ==================== Типы ответов GraphQL ====================

interface GetEntrepreneurResponse {
  entrepreneur: IndividualEntrepreneur | null;
}

interface GetEntrepreneurActivitiesResponse {
  entrepreneur: {
    activities: Array<{
      code: string;
      name: string;
      isMain?: boolean;
    }>;
  } | null;
}

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

// ==================== Хуки ====================

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

export function useEntrepreneurActivitiesQuery(
  ogrnip: string,
  options?: UseQueryOptions<GetEntrepreneurActivitiesResponse, Error>
) {
  return useQuery<GetEntrepreneurActivitiesResponse, Error>({
    ...options,
    queryKey: ["entrepreneur-activities", ogrnip],
    queryFn: () =>
      defaultGraphQLClient.request<GetEntrepreneurActivitiesResponse, { ogrnip: string }>(
        /* GraphQL */ `
          query GetEntrepreneurActivities($ogrnip: ID!) {
            entrepreneur(ogrnip: $ogrnip) {
              activities {
                code
                name
                isMain
              }
            }
          }
        `,
        { ogrnip }
      ),
  });
}

export function useEntrepreneurHistoryQuery(
  ogrnip: string,
  limit: number = 20,
  offset: number = 0,
  options?: UseQueryOptions<GetEntrepreneurHistoryResponse, Error>
) {
  return useQuery<GetEntrepreneurHistoryResponse, Error>({
    ...options,
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
  });
}