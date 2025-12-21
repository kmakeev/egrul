const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export interface LegalEntity {
  id?: string;
  ogrn: string;
  ogrnDate?: string;
  inn: string;
  kpp?: string;
  fullName?: string;
  shortName?: string;
  brandName?: string;
  legalForm?: string;
  status?: string;
  statusCode?: string;
  terminationMethod?: string;
  registrationDate?: string;
  terminationDate?: string;
  extractDate?: string;
  address?: Address;
  email?: string;
  capital?: Capital;
  director?: Person;
  mainActivity?: Activity;
  activities?: Activity[];
  regAuthority?: string;
  taxAuthority?: string;
  pfrRegNumber?: string;
  fssRegNumber?: string;
  foundersCount?: number;
  licensesCount?: number;
  branchesCount?: number;
  isBankrupt?: boolean;
  bankruptcyStage?: string;
  isLiquidating?: boolean;
  isReorganizing?: boolean;
  lastGrn?: string;
  lastGrnDate?: string;
  sourceFile?: string;
  versionDate?: string;
  createdAt?: string;
  updatedAt?: string;
  // Legacy fields for backward compatibility
  head?: Person;
  founders?: Founder[];
  history?: HistoryRecord[];
  relatedCompanies?: RelatedCompany[];
  registrationAuthority?: string;
}

export interface Founder {
  id: string;
  type: "individual" | "legal";
  name: string;
  inn?: string;
  share?: number;
  amount?: number;
  currency?: string;
}

export interface HistoryRecord {
  id: string;
  date: string;
  type: string;
  description: string;
  details?: Record<string, unknown>;
}

export interface RelatedCompany {
  id: string;
  ogrn: string;
  name: string;
  relationshipType: string;
  status: string;
}

export interface IndividualEntrepreneur {
  id: string;
  ogrnip: string;
  inn: string;
  lastName: string;
  firstName: string;
  middleName?: string;
  registrationDate?: string;
  address?: Address;
  status: string;
  mainActivity?: Activity;
}

export interface Address {
  postalCode?: string;
  regionCode?: string;
  region?: string;
  district?: string;
  city?: string;
  locality?: string;
  street?: string;
  house?: string;
  building?: string;
  flat?: string;
  office?: string;
  fullAddress?: string;
  fiasId?: string;
  kladrCode?: string;
}

export interface Activity {
  code: string;
  name: string;
  isMain?: boolean;
}

export interface Capital {
  amount: number;
  currency: string;
}

export interface Person {
  lastName?: string;
  firstName?: string;
  middleName?: string;
  inn?: string;
  position?: string;
  positionCode?: string;
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

