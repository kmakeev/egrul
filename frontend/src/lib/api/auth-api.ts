"use client";

import { graphqlFetcher } from "./graphql-client";
import type { User } from "@/types";

// GraphQL mutations and queries
const REGISTER_MUTATION = `
  mutation Register($input: RegisterInput!) {
    register(input: $input) {
      user {
        id
        email
        firstName
        lastName
        createdAt
      }
      token
      expiresAt
    }
  }
`;

const LOGIN_MUTATION = `
  mutation Login($input: LoginInput!) {
    login(input: $input) {
      user {
        id
        email
        firstName
        lastName
        createdAt
      }
      token
      expiresAt
    }
  }
`;

const ME_QUERY = `
  query Me {
    me {
      id
      email
      firstName
      lastName
      emailVerified
      createdAt
    }
  }
`;

// Types
export interface RegisterInput {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
}

export interface LoginInput {
  email: string;
  password: string;
}

export interface AuthResponse {
  user: User;
  token: string;
  expiresAt: string;
}

interface RegisterResponse {
  register: AuthResponse;
}

interface LoginResponse {
  login: AuthResponse;
}

interface MeResponse {
  me: User & { emailVerified?: boolean };
}

// API functions
export async function register(input: RegisterInput): Promise<AuthResponse> {
  const data = await graphqlFetcher<RegisterResponse>(REGISTER_MUTATION, {
    input,
  });
  return data.register;
}

export async function login(input: LoginInput): Promise<AuthResponse> {
  const data = await graphqlFetcher<LoginResponse>(LOGIN_MUTATION, {
    input,
  });
  return data.login;
}

export async function getCurrentUser(): Promise<User & { emailVerified?: boolean }> {
  const data = await graphqlFetcher<MeResponse>(ME_QUERY);
  return data.me;
}
