"use client";

import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
  type UseMutationOptions,
} from "@tanstack/react-query";
import { defaultGraphQLClient } from "@/lib/api/graphql-client";
import { EntityType } from "./subscription-hooks";

// ==================== Типы ====================

export interface Favorite {
  id: string;
  userId: string;
  entityType: EntityType;
  entityId: string;
  entityName: string;
  notes: string | null;
  createdAt: string;
}

export interface CreateFavoriteInput {
  entityType: EntityType;
  entityId: string;
  entityName: string;
  notes: string | null;
}

export interface UpdateFavoriteNotesInput {
  id: string;
  notes: string | null;
}

// ==================== Response Types ====================

interface MyFavoritesResponse {
  myFavorites: Favorite[];
}

interface HasFavoriteResponse {
  hasFavorite: boolean;
}

interface CreateFavoriteResponse {
  createFavorite: Favorite;
}

interface UpdateFavoriteNotesResponse {
  updateFavoriteNotes: Favorite;
}

interface DeleteFavoriteResponse {
  deleteFavorite: boolean;
}

// ==================== Queries ====================

export function useMyFavoritesQuery(
  options?: Omit<
    UseQueryOptions<MyFavoritesResponse, Error>,
    "queryKey" | "queryFn"
  >
) {
  return useQuery<MyFavoritesResponse, Error>({
    queryKey: ["favorites", "my"],
    queryFn: () =>
      defaultGraphQLClient.request<MyFavoritesResponse>(
        /* GraphQL */ `
          query MyFavorites {
            myFavorites {
              id
              userId
              entityType
              entityId
              entityName
              notes
              createdAt
            }
          }
        `
      ),
    ...options,
  });
}

export function useHasFavoriteQuery(
  entityType: EntityType,
  entityId: string,
  options?: Omit<
    UseQueryOptions<HasFavoriteResponse, Error>,
    "queryKey" | "queryFn"
  >
) {
  return useQuery<HasFavoriteResponse, Error>({
    queryKey: ["favorite", "has", entityType, entityId],
    queryFn: () =>
      defaultGraphQLClient.request<
        HasFavoriteResponse,
        { entityType: EntityType; entityId: string }
      >(
        /* GraphQL */ `
          query HasFavorite(
            $entityType: EntityType!
            $entityId: String!
          ) {
            hasFavorite(
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

// ==================== Mutations ====================

export function useCreateFavoriteMutation(
  options?: UseMutationOptions<
    CreateFavoriteResponse,
    Error,
    CreateFavoriteInput
  >
) {
  const queryClient = useQueryClient();

  return useMutation<
    CreateFavoriteResponse,
    Error,
    CreateFavoriteInput
  >({
    mutationFn: (input: CreateFavoriteInput) =>
      defaultGraphQLClient.request<
        CreateFavoriteResponse,
        { input: CreateFavoriteInput }
      >(
        /* GraphQL */ `
          mutation CreateFavorite($input: CreateFavoriteInput!) {
            createFavorite(input: $input) {
              id
              userId
              entityType
              entityId
              entityName
              notes
              createdAt
            }
          }
        `,
        { input }
      ),
    onSuccess: (data) => {
      // Инвалидируем кэш избранного пользователя
      queryClient.invalidateQueries({
        queryKey: ["favorites", "my"],
      });
      // Инвалидируем проверку наличия в избранном
      queryClient.invalidateQueries({
        queryKey: [
          "favorite",
          "has",
          data.createFavorite.entityType,
          data.createFavorite.entityId,
        ],
      });
    },
    ...options,
  });
}

export function useUpdateFavoriteNotesMutation(
  options?: UseMutationOptions<
    UpdateFavoriteNotesResponse,
    Error,
    UpdateFavoriteNotesInput
  >
) {
  const queryClient = useQueryClient();

  return useMutation<
    UpdateFavoriteNotesResponse,
    Error,
    UpdateFavoriteNotesInput
  >({
    mutationFn: (input: UpdateFavoriteNotesInput) =>
      defaultGraphQLClient.request<
        UpdateFavoriteNotesResponse,
        { input: UpdateFavoriteNotesInput }
      >(
        /* GraphQL */ `
          mutation UpdateFavoriteNotes(
            $input: UpdateFavoriteNotesInput!
          ) {
            updateFavoriteNotes(input: $input) {
              id
              notes
            }
          }
        `,
        { input }
      ),
    onSuccess: () => {
      // Инвалидируем список избранного
      queryClient.invalidateQueries({ queryKey: ["favorites", "my"] });
    },
    ...options,
  });
}

export function useDeleteFavoriteMutation(
  options?: UseMutationOptions<DeleteFavoriteResponse, Error, string>
) {
  const queryClient = useQueryClient();

  return useMutation<DeleteFavoriteResponse, Error, string>({
    mutationFn: (id: string) =>
      defaultGraphQLClient.request<
        DeleteFavoriteResponse,
        { id: string }
      >(
        /* GraphQL */ `
          mutation DeleteFavorite($id: ID!) {
            deleteFavorite(id: $id)
          }
        `,
        { id }
      ),
    onSuccess: () => {
      // Инвалидируем все избранное
      queryClient.invalidateQueries({ queryKey: ["favorites"] });
      queryClient.invalidateQueries({ queryKey: ["favorite"] });
    },
    ...options,
  });
}
