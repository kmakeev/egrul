const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export interface LegalEntity {
  id: string;
  ogrn: string;
  inn: string;
  kpp?: string;
  fullName: string;
  shortName?: string;
  registrationDate?: string;
  address?: Address;
  status: string;
  mainActivity?: Activity;
  capital?: Capital;
  head?: Person;
}

export interface IndividualEntrepreneur {
  id: string;
  ogrnip: string;
  inn: string;
  lastName: string;
  firstName: string;
  middleName?: string;
  registrationDate?: string;
  status: string;
  mainActivity?: Activity;
}

export interface Address {
  postalCode?: string;
  region?: string;
  city?: string;
  street?: string;
  house?: string;
  office?: string;
  fullAddress?: string;
}

export interface Activity {
  code: string;
  name: string;
}

export interface Capital {
  amount: number;
  currency: string;
}

export interface Person {
  lastName: string;
  firstName: string;
  middleName?: string;
  inn?: string;
  position?: string;
}

export interface SearchResult {
  legalEntities: LegalEntity[];
  entrepreneurs: IndividualEntrepreneur[];
  total: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
    });

    if (!response.ok) {
      throw new Error(`API Error: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }

  // Юридические лица
  async getLegalEntities(
    page = 1,
    pageSize = 20
  ): Promise<PaginatedResponse<LegalEntity>> {
    return this.request(`/legal-entities?page=${page}&page_size=${pageSize}`);
  }

  async getLegalEntity(ogrn: string): Promise<LegalEntity> {
    return this.request(`/legal-entities/${ogrn}`);
  }

  async searchLegalEntities(query: string): Promise<{ results: LegalEntity[] }> {
    return this.request(`/legal-entities/search?q=${encodeURIComponent(query)}`);
  }

  // Индивидуальные предприниматели
  async getEntrepreneurs(
    page = 1,
    pageSize = 20
  ): Promise<PaginatedResponse<IndividualEntrepreneur>> {
    return this.request(`/entrepreneurs?page=${page}&page_size=${pageSize}`);
  }

  async getEntrepreneur(ogrnip: string): Promise<IndividualEntrepreneur> {
    return this.request(`/entrepreneurs/${ogrnip}`);
  }

  async searchEntrepreneurs(
    query: string
  ): Promise<{ results: IndividualEntrepreneur[] }> {
    return this.request(`/entrepreneurs/search?q=${encodeURIComponent(query)}`);
  }

  // Глобальный поиск
  async globalSearch(query: string): Promise<SearchResult> {
    return this.request(`/search?q=${encodeURIComponent(query)}`);
  }
}

export const api = new ApiClient(API_BASE_URL);

