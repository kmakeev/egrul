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
  historyCount?: number;
  relatedCompanies?: RelatedCompany[];
  registrationAuthority?: string;
}

export interface Founder {
  type: "PERSON" | "RUSSIAN_COMPANY" | "FOREIGN_COMPANY" | "PUBLIC_ENTITY" | "FUND";
  ogrn?: string;
  inn?: string;
  name: string;
  lastName?: string;
  firstName?: string;
  middleName?: string;
  country?: string;
  citizenship?: string;
  shareNominalValue?: number;
  sharePercent?: number;
}

export interface HistoryRecord {
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
}

export interface RelatedCompany {
  id: string;
  ogrn: string;
  name: string;
  relationshipType: string;
  status: string;
  company?: LegalEntity;
  description?: string;
}

export interface IndividualEntrepreneur {
  id: string;
  ogrnip: string;
  ogrnipDate?: string;
  inn: string;
  lastName: string;
  firstName: string;
  middleName?: string;
  registrationDate?: string;
  terminationDate?: string;
  address?: Address;
  status?: string;
  statusCode?: string;
  citizenshipType?: string;
  citizenshipCountryCode?: string;
  citizenshipCountryName?: string;
  mainActivity?: Activity;
  activities?: Activity[];
  regAuthority?: Authority;
  taxAuthority?: Authority;
  history?: HistoryRecord[];
  historyCount?: number;
  licensesCount?: number;
  extractDate?: string;
  lastGrn?: string;
  lastGrnDate?: string;
  sourceFile?: string;
  versionDate?: string;
  createdAt?: string;
  updatedAt?: string;
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
  fullAddress?: string;
  fiasId?: string;
  kladrCode?: string;
}

export interface Authority {
  code?: string;
  name?: string;
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

export interface License {
  id: string;
  number: string;
  series?: string;
  activity?: string;
  startDate?: string;
  endDate?: string;
  authority?: string;
  status?: string;
}

export interface Branch {
  id: string;
  type: 'BRANCH' | 'REPRESENTATIVE';
  name?: string;
  kpp?: string;
  address?: Address;
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

  // Лицензии компании
  async getCompanyLicenses(ogrn: string): Promise<License[]> {
    const query = `
      query GetCompanyLicenses($ogrn: ID!) {
        company(ogrn: $ogrn) {
          licenses {
            id
            number
            series
            activity
            startDate
            endDate
            authority
            status
          }
        }
      }
    `;
    
    const response = await this.graphqlRequest<{ company: { licenses: License[] } }>(query, { ogrn });
    return response.company?.licenses || [];
  }

  // Филиалы компании
  async getCompanyBranches(ogrn: string): Promise<Branch[]> {
    const query = `
      query GetCompanyBranches($ogrn: ID!) {
        company(ogrn: $ogrn) {
          branches {
            id
            type
            name
            kpp
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
            }
          }
        }
      }
    `;
    
    const response = await this.graphqlRequest<{ company: { branches: Branch[] } }>(query, { ogrn });
    return response.company?.branches || [];
  }

  private async graphqlRequest<T>(query: string, variables?: Record<string, unknown>): Promise<T> {
    const response = await fetch(`${this.baseUrl.replace('/api/v1', '')}/graphql`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query,
        variables,
      }),
    });

    if (!response.ok) {
      throw new Error(`GraphQL Error: ${response.status} ${response.statusText}`);
    }

    const result = await response.json();
    
    if (result.errors) {
      throw new Error(`GraphQL Error: ${result.errors[0].message}`);
    }

    return result.data;
  }

  // Глобальный поиск
  async globalSearch(query: string): Promise<SearchResult> {
    return this.request(`/search?q=${encodeURIComponent(query)}`);
  }
}

export const api = new ApiClient(API_BASE_URL);

