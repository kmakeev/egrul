"use client";

import { useAuthStore } from "@/store/auth-store";

const GRAPHQL_ENDPOINT =
  process.env.NEXT_PUBLIC_GRAPHQL_URL || "http://localhost:8080/graphql";

export interface GraphQLErrorItem {
  message: string;
  path?: readonly (string | number)[];
  extensions?: Record<string, unknown>;
}

export interface GraphQLResponse<TData> {
  data?: TData;
  errors?: GraphQLErrorItem[];
}

export interface GraphQLRequestOptions<TVariables extends Record<string, unknown>> {
  query: string;
  variables?: TVariables;
  signal?: AbortSignal;
  headers?: HeadersInit;
  /** Количество повторов при временных ошибках (по умолчанию 1) */
  retries?: number;
  /** Базовая задержка между повторами (мс) */
  retryDelayMs?: number;
}

export class GraphQLRequestError extends Error {
  public readonly status?: number;
  public readonly graphQLErrors?: GraphQLErrorItem[];

  constructor(message: string, options?: { status?: number; errors?: GraphQLErrorItem[] }) {
    super(message);
    this.name = "GraphQLRequestError";
    this.status = options?.status;
    this.graphQLErrors = options?.errors;
  }
}

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function getAuthToken(): string | null {
  try {
    return useAuthStore.getState().token ?? null;
  } catch {
    return null;
  }
}

export async function rawGraphQLRequest<
  TData,
  TVariables extends Record<string, unknown> = Record<string, unknown>
>(
  options: GraphQLRequestOptions<TVariables>
): Promise<TData> {
  const {
    query,
    variables,
    signal,
    headers,
    retries = 1,
    retryDelayMs = 500,
  } = options;

  const token = getAuthToken();

  let attempt = 0;
  // eslint-disable-next-line no-constant-condition
  while (true) {
    try {
      const requestBody = {
        query,
        variables,
      };

      const response = await fetch(GRAPHQL_ENDPOINT, {
        method: "POST",
        signal,
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
          ...headers,
        },
        body: JSON.stringify(requestBody),
      });

      const json = (await response.json()) as GraphQLResponse<TData>;

      if (!response.ok) {
        throw new GraphQLRequestError(
          json.errors?.[0]?.message ||
            `GraphQL request failed with status ${response.status}`,
          { status: response.status, errors: json.errors }
        );
      }

      if (json.errors && json.errors.length > 0) {
        throw new GraphQLRequestError(json.errors[0].message, {
          status: response.status,
          errors: json.errors,
        });
      }

      if (!("data" in json)) {
        throw new GraphQLRequestError("GraphQL response does not contain data field");
      }

      return json.data as TData;
    } catch (error) {
      attempt += 1;

      const isLastAttempt = attempt > retries;
      const isAbortError =
        error instanceof DOMException && error.name === "AbortError";

      if (isAbortError || isLastAttempt) {
        throw error;
      }

      // Экспоненциальная задержка между повторами
      const delay = retryDelayMs * attempt;
      // eslint-disable-next-line no-await-in-loop
      await sleep(delay);
    }
  }
}

/**
 * Функция-обертка для использования в GraphQL Code Generator (fetcher)
 */
export function graphqlFetcher<
  TData,
  TVariables extends Record<string, unknown> = Record<string, unknown>
>(
  query: string,
  variables?: TVariables
) {
  return rawGraphQLRequest<TData, TVariables>({
    query,
    variables,
  });
}

export class GraphQLClient {
  private readonly endpoint: string;

  constructor(endpoint: string = GRAPHQL_ENDPOINT) {
    this.endpoint = endpoint;
  }

  request<
    TData,
    TVariables extends Record<string, unknown> = Record<string, unknown>
  >(
    query: string,
    variables?: TVariables,
    options?: Omit<GraphQLRequestOptions<TVariables>, "query" | "variables">
  ) {
    return rawGraphQLRequest<TData, TVariables>({
      query,
      variables,
      ...options,
    });
  }
}

export const defaultGraphQLClient = new GraphQLClient();
