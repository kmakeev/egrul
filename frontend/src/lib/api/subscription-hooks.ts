"use client";

import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
  type UseMutationOptions,
} from "@tanstack/react-query";
import { defaultGraphQLClient } from "@/lib/api/graphql-client";

// ==================== Типы ====================

export enum EntityType {
  COMPANY = "COMPANY",
  ENTREPRENEUR = "ENTREPRENEUR",
}

export enum NotificationStatus {
  PENDING = "PENDING",
  SENT = "SENT",
  FAILED = "FAILED",
}

export interface ChangeFilters {
  status: boolean;
  director: boolean;
  founders: boolean;
  address: boolean;
  capital: boolean;
  activities: boolean;
}

export interface NotificationChannels {
  email: boolean;
}

export interface EntitySubscription {
  id: string;
  userId: string;
  entityType: EntityType;
  entityId: string;
  entityName: string;
  changeFilters: ChangeFilters;
  notificationChannels: NotificationChannels;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  lastNotifiedAt?: string | null;
}

export interface NotificationLogEntry {
  id: string;
  subscriptionId: string;
  changeEventId: string;
  entityType: EntityType;
  entityId: string;
  entityName: string;
  changeType: string;
  fieldName?: string | null;
  oldValue?: string | null;
  newValue?: string | null;
  detectedAt: string;
  channel: string;
  recipient: string;
  status: NotificationStatus;
  sentAt?: string | null;
  retryCount: number;
  errorMessage?: string | null;
  createdAt: string;
}

export interface CreateSubscriptionInput {
  entityType: EntityType;
  entityId: string;
  entityName: string;
  changeFilters: ChangeFilters;
  notificationChannels: NotificationChannels;
}

export interface UpdateSubscriptionFiltersInput {
  id: string;
  changeFilters: ChangeFilters;
}

export interface UpdateSubscriptionChannelsInput {
  id: string;
  notificationChannels: NotificationChannels;
}

export interface ToggleSubscriptionInput {
  id: string;
  isActive: boolean;
}

// ==================== Response Types ====================

interface MySubscriptionsResponse {
  mySubscriptions: EntitySubscription[];
}

interface GetSubscriptionResponse {
  subscription: EntitySubscription | null;
}

interface HasSubscriptionResponse {
  hasSubscription: boolean;
}

interface NotificationHistoryResponse {
  notificationHistory: NotificationLogEntry[];
}

interface CreateSubscriptionResponse {
  createSubscription: EntitySubscription;
}

interface UpdateSubscriptionFiltersResponse {
  updateSubscriptionFilters: EntitySubscription;
}

interface UpdateSubscriptionChannelsResponse {
  updateSubscriptionChannels: EntitySubscription;
}

interface DeleteSubscriptionResponse {
  deleteSubscription: boolean;
}

interface ToggleSubscriptionResponse {
  toggleSubscription: EntitySubscription;
}

// ==================== Queries ====================

export function useMySubscriptionsQuery(
  options?: Omit<
    UseQueryOptions<MySubscriptionsResponse, Error>,
    "queryKey" | "queryFn"
  >
) {
  return useQuery<MySubscriptionsResponse, Error>({
    queryKey: ["subscriptions", "my"],
    queryFn: () =>
      defaultGraphQLClient.request<MySubscriptionsResponse>(
        /* GraphQL */ `
          query MySubscriptions {
            mySubscriptions {
              id
              userId
              entityType
              entityId
              entityName
              changeFilters {
                status
                director
                founders
                address
                capital
                activities
              }
              notificationChannels {
                email
              }
              isActive
              createdAt
              updatedAt
              lastNotifiedAt
            }
          }
        `
      ),
    ...options,
  });
}

export function useSubscriptionQuery(
  id: string,
  options?: Omit<
    UseQueryOptions<GetSubscriptionResponse, Error>,
    "queryKey" | "queryFn"
  >
) {
  return useQuery<GetSubscriptionResponse, Error>({
    queryKey: ["subscription", id],
    queryFn: () =>
      defaultGraphQLClient.request<GetSubscriptionResponse, { id: string }>(
        /* GraphQL */ `
          query GetSubscription($id: ID!) {
            subscription(id: $id) {
              id
              userId
              entityType
              entityId
              entityName
              changeFilters {
                status
                director
                founders
                address
                capital
                activities
              }
              notificationChannels {
                email
              }
              isActive
              createdAt
              updatedAt
              lastNotifiedAt
            }
          }
        `,
        { id }
      ),
    enabled: !!id,
    ...options,
  });
}

export function useHasSubscriptionQuery(
  entityType: EntityType,
  entityId: string,
  options?: Omit<
    UseQueryOptions<HasSubscriptionResponse, Error>,
    "queryKey" | "queryFn"
  >
) {
  return useQuery<HasSubscriptionResponse, Error>({
    queryKey: ["subscription", "has", entityType, entityId],
    queryFn: () =>
      defaultGraphQLClient.request<
        HasSubscriptionResponse,
        { entityType: EntityType; entityId: string }
      >(
        /* GraphQL */ `
          query HasSubscription(
            $entityType: EntityType!
            $entityId: String!
          ) {
            hasSubscription(
              entityType: $entityType
              entityId: $entityId
            )
          }
        `,
        { entityType, entityId }
      ),
    enabled: !!entityType && !!entityId,
    ...options,
  });
}

export function useNotificationHistoryQuery(
  subscriptionId: string,
  limit: number = 20,
  offset: number = 0,
  options?: Omit<
    UseQueryOptions<NotificationHistoryResponse, Error>,
    "queryKey" | "queryFn"
  >
) {
  return useQuery<NotificationHistoryResponse, Error>({
    queryKey: ["notificationHistory", subscriptionId, limit, offset],
    queryFn: () =>
      defaultGraphQLClient.request<
        NotificationHistoryResponse,
        { subscriptionId: string; limit: number; offset: number }
      >(
        /* GraphQL */ `
          query NotificationHistory(
            $subscriptionId: String!
            $limit: Int
            $offset: Int
          ) {
            notificationHistory(
              subscriptionId: $subscriptionId
              limit: $limit
              offset: $offset
            ) {
              id
              subscriptionId
              changeEventId
              entityType
              entityId
              entityName
              changeType
              fieldName
              oldValue
              newValue
              detectedAt
              channel
              recipient
              status
              sentAt
              retryCount
              errorMessage
              createdAt
            }
          }
        `,
        { subscriptionId, limit, offset }
      ),
    enabled: !!subscriptionId,
    ...options,
  });
}

// ==================== Mutations ====================

export function useCreateSubscriptionMutation(
  options?: UseMutationOptions<
    CreateSubscriptionResponse,
    Error,
    CreateSubscriptionInput
  >
) {
  const queryClient = useQueryClient();

  return useMutation<
    CreateSubscriptionResponse,
    Error,
    CreateSubscriptionInput
  >({
    mutationFn: (input: CreateSubscriptionInput) =>
      defaultGraphQLClient.request<
        CreateSubscriptionResponse,
        { input: CreateSubscriptionInput }
      >(
        /* GraphQL */ `
          mutation CreateSubscription($input: CreateSubscriptionInput!) {
            createSubscription(input: $input) {
              id
              userId
              entityType
              entityId
              entityName
              changeFilters {
                status
                director
                founders
                address
                capital
                activities
              }
              notificationChannels {
                email
              }
              isActive
              createdAt
            }
          }
        `,
        { input }
      ),
    onSuccess: (data) => {
      // Инвалидируем кэш подписок пользователя
      queryClient.invalidateQueries({
        queryKey: ["subscriptions", "my"],
      });
      // Инвалидируем проверку наличия подписки
      queryClient.invalidateQueries({
        queryKey: [
          "subscription",
          "has",
          data.createSubscription.entityType,
          data.createSubscription.entityId,
        ],
      });
    },
    ...options,
  });
}

export function useUpdateSubscriptionFiltersMutation(
  options?: UseMutationOptions<
    UpdateSubscriptionFiltersResponse,
    Error,
    UpdateSubscriptionFiltersInput
  >
) {
  const queryClient = useQueryClient();

  return useMutation<
    UpdateSubscriptionFiltersResponse,
    Error,
    UpdateSubscriptionFiltersInput
  >({
    mutationFn: (input: UpdateSubscriptionFiltersInput) =>
      defaultGraphQLClient.request<
        UpdateSubscriptionFiltersResponse,
        { input: UpdateSubscriptionFiltersInput }
      >(
        /* GraphQL */ `
          mutation UpdateSubscriptionFilters(
            $input: UpdateSubscriptionFiltersInput!
          ) {
            updateSubscriptionFilters(input: $input) {
              id
              changeFilters {
                status
                director
                founders
                address
                capital
                activities
              }
              updatedAt
            }
          }
        `,
        { input }
      ),
    onSuccess: (data) => {
      // Инвалидируем подписку
      queryClient.invalidateQueries({
        queryKey: ["subscription", data.updateSubscriptionFilters.id],
      });
      // Инвалидируем список подписок (не знаем email, инвалидируем все)
      queryClient.invalidateQueries({ queryKey: ["subscriptions", "my"] });
    },
    ...options,
  });
}

export function useUpdateSubscriptionChannelsMutation(
  options?: UseMutationOptions<
    UpdateSubscriptionChannelsResponse,
    Error,
    UpdateSubscriptionChannelsInput
  >
) {
  const queryClient = useQueryClient();

  return useMutation<
    UpdateSubscriptionChannelsResponse,
    Error,
    UpdateSubscriptionChannelsInput
  >({
    mutationFn: (input: UpdateSubscriptionChannelsInput) =>
      defaultGraphQLClient.request<
        UpdateSubscriptionChannelsResponse,
        { input: UpdateSubscriptionChannelsInput }
      >(
        /* GraphQL */ `
          mutation UpdateSubscriptionChannels(
            $input: UpdateSubscriptionChannelsInput!
          ) {
            updateSubscriptionChannels(input: $input) {
              id
              notificationChannels {
                email
              }
              updatedAt
            }
          }
        `,
        { input }
      ),
    onSuccess: (data) => {
      queryClient.invalidateQueries({
        queryKey: ["subscription", data.updateSubscriptionChannels.id],
      });
      queryClient.invalidateQueries({ queryKey: ["subscriptions", "my"] });
    },
    ...options,
  });
}

export function useDeleteSubscriptionMutation(
  options?: UseMutationOptions<DeleteSubscriptionResponse, Error, string>
) {
  const queryClient = useQueryClient();

  return useMutation<DeleteSubscriptionResponse, Error, string>({
    mutationFn: (id: string) =>
      defaultGraphQLClient.request<
        DeleteSubscriptionResponse,
        { id: string }
      >(
        /* GraphQL */ `
          mutation DeleteSubscription($id: ID!) {
            deleteSubscription(id: $id)
          }
        `,
        { id }
      ),
    onSuccess: () => {
      // Инвалидируем все подписки
      queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
      queryClient.invalidateQueries({ queryKey: ["subscription"] });
    },
    ...options,
  });
}

export function useToggleSubscriptionMutation(
  options?: UseMutationOptions<
    ToggleSubscriptionResponse,
    Error,
    ToggleSubscriptionInput
  >
) {
  const queryClient = useQueryClient();

  return useMutation<
    ToggleSubscriptionResponse,
    Error,
    ToggleSubscriptionInput
  >({
    mutationFn: (input: ToggleSubscriptionInput) =>
      defaultGraphQLClient.request<
        ToggleSubscriptionResponse,
        { input: ToggleSubscriptionInput }
      >(
        /* GraphQL */ `
          mutation ToggleSubscription($input: ToggleSubscriptionInput!) {
            toggleSubscription(input: $input) {
              id
              isActive
              updatedAt
            }
          }
        `,
        { input }
      ),
    onSuccess: (data) => {
      queryClient.invalidateQueries({
        queryKey: ["subscription", data.toggleSubscription.id],
      });
      queryClient.invalidateQueries({ queryKey: ["subscriptions", "my"] });
    },
    ...options,
  });
}
