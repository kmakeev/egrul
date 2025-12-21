"use client";

import { useParams } from "next/navigation";
import { CompanyHeader } from "@/components/company/company-header";
import { CompanyTabs } from "@/components/company/company-tabs";
import { useCompanyQuery } from "@/lib/api/hooks";
import { generateCompanyJsonLd } from "./metadata";
import { getMockCompanyByOgrn, useMockData } from "@/components/company/mock-data";

export default function CompanyPage() {
  const params = useParams();
  const ogrn = params.ogrn as string;

  const { data, isLoading, error } = useCompanyQuery(ogrn);

  // Используем mock данные в режиме разработки или при ошибке API
  let company = data?.company;
  
  if (useMockData || (!company && !isLoading && error)) {
    company = getMockCompanyByOgrn(ogrn);
  }

  if (isLoading && !useMockData) {
    return <CompanyPageLoading />;
  }

  if (error && !useMockData && !company) {
    return <CompanyPageError error={error} />;
  }

  if (!company) {
    return <CompanyPageNotFound />;
  }

  return (
    <>
      {/* JSON-LD структурированные данные */}
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(generateCompanyJsonLd(company))
        }}
      />
      
      <div className="container mx-auto p-6 space-y-6">
        {/* Показываем предупреждение о mock данных */}
        {(useMockData || (error && company)) && (
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6">
            <div className="flex items-center">
              <div className="text-yellow-600">
                <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-yellow-800">
                  {useMockData 
                    ? "Используются тестовые данные (режим разработки)"
                    : "API недоступно, показаны демонстрационные данные"
                  }
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Отладочная информация в режиме разработки */}
        {process.env.NODE_ENV === "development" && !useMockData && (
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
            <details>
              <summary className="cursor-pointer text-sm font-medium text-blue-800">
                Отладочная информация (только в dev режиме)
              </summary>
              <div className="mt-2 text-xs text-blue-700">
                <p><strong>ОГРН:</strong> {ogrn}</p>
                <p><strong>Название:</strong> {company?.fullName || "не загружено"}</p>
                <p><strong>Статус:</strong> &quot;{company?.status || "не загружено"}&quot; (тип: {typeof company?.status})</p>
                <p><strong>Статус загрузки:</strong> {isLoading ? "загружается" : "загружено"}</p>
                <p><strong>Ошибка:</strong> {error ? error.message : "нет"}</p>
                <p><strong>Данные получены:</strong> {data ? "да" : "нет"}</p>
              </div>
            </details>
          </div>
        )}
        
        <CompanyHeader company={company} />
        <CompanyTabs company={company} />
      </div>
    </>
  );
}

function CompanyPageLoading() {
  return (
    <div className="container mx-auto p-6">
      <div className="space-y-6">
        {/* Skeleton для заголовка */}
        <div className="bg-muted rounded-lg border p-6">
          <div className="animate-pulse">
            <div className="h-8 bg-muted-foreground/20 rounded w-3/4 mb-4"></div>
            <div className="h-4 bg-muted-foreground/20 rounded w-1/2 mb-6"></div>
            <div className="grid grid-cols-3 gap-6">
              <div className="h-16 bg-muted-foreground/20 rounded"></div>
              <div className="h-16 bg-muted-foreground/20 rounded"></div>
              <div className="h-16 bg-muted-foreground/20 rounded"></div>
            </div>
          </div>
        </div>
        
        {/* Skeleton для табов */}
        <div className="bg-muted rounded-lg border p-6">
          <div className="animate-pulse">
            <div className="flex space-x-4 mb-6">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="h-10 bg-muted-foreground/20 rounded w-32"></div>
              ))}
            </div>
            <div className="space-y-4">
              <div className="h-32 bg-muted-foreground/20 rounded"></div>
              <div className="h-32 bg-muted-foreground/20 rounded"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function CompanyPageError({ error }: { error: Error }) {
  return (
    <div className="container mx-auto p-6">
      <div className="bg-muted rounded-lg border p-6 text-center">
        <div className="text-red-600 mb-4">
          <svg className="h-12 w-12 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
          </svg>
        </div>
        <h2 className="text-xl font-semibold text-gray-900 mb-2">Ошибка загрузки</h2>
        <p className="text-gray-600 mb-4">
          {error instanceof Error ? error.message : "Произошла неизвестная ошибка"}
        </p>
        <button 
          onClick={() => window.location.reload()} 
          className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition-colors"
        >
          Попробовать снова
        </button>
      </div>
    </div>
  );
}

function CompanyPageNotFound() {
  return (
    <div className="container mx-auto p-6">
      <div className="bg-muted rounded-lg border p-6 text-center">
        <div className="text-gray-400 mb-4">
          <svg className="h-12 w-12 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-4m-5 0H9m0 0H5m0 0h2M7 7h10M7 11h6m-6 4h6" />
          </svg>
        </div>
        <h2 className="text-xl font-semibold text-gray-900 mb-2">Компания не найдена</h2>
        <p className="text-gray-600">
          Компания с указанным ОГРН не найдена в базе данных
        </p>
      </div>
    </div>
  );
}

