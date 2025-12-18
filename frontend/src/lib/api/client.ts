import type {
  LegalEntity,
  IndividualEntrepreneur,
  SearchResult,
  PaginatedResponse,
} from "@/types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

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
      const error = await response.json().catch(() => ({
        message: `API Error: ${response.status} ${response.statusText}`,
      }));
      throw new Error(error.message || `API Error: ${response.status}`);
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

export const apiClient = new ApiClient(API_BASE_URL);

