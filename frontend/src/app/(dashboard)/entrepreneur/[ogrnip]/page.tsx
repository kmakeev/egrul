"use client";

import { useParams } from "next/navigation";
import { EntrepreneurHeader } from "@/components/entrepreneur/entrepreneur-header";
import { EntrepreneurTabs } from "@/components/entrepreneur/entrepreneur-tabs";
import { useEntrepreneurQuery } from "@/lib/api/hooks";
import { Card, CardContent } from "@/components/ui/card";
import { getMockEntrepreneurByOgrnip, useMockData } from "@/components/entrepreneur/mock-data";

export default function EntrepreneurPage() {
  const params = useParams();
  const ogrnip = params.ogrnip as string;

  const { data, isLoading, error } = useEntrepreneurQuery(ogrnip);

  // Используем mock данные в режиме разработки или при ошибке API
  let entrepreneur = data?.entrepreneur;
  
  if (useMockData || (!entrepreneur && !isLoading && error)) {
    entrepreneur = getMockEntrepreneurByOgrnip(ogrnip);
  }

  if (isLoading && !useMockData) {
    return <EntrepreneurPageLoading />;
  }

  if (error && !useMockData && !entrepreneur) {
    return <EntrepreneurPageError error={error} />;
  }

  if (!entrepreneur) {
    return <EntrepreneurPageNotFound />;
  }

  return (
    <>
      <div className="container mx-auto p-6 space-y-6">
        {/* Показываем предупреждение о mock данных */}
        {(useMockData || (error && entrepreneur)) && (
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6">
            <div className="flex items-center">
              <div className="text-yellow-600">
                <svg className="w-5 h-5 mr-2 inline" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
                Используются тестовые данные для разработки
              </div>
            </div>
          </div>
        )}
        
        <EntrepreneurHeader entrepreneur={entrepreneur} />
        <EntrepreneurTabs entrepreneur={entrepreneur} />
      </div>
    </>
  );
}

function EntrepreneurPageLoading() {
  return (
    <div className="container mx-auto p-6">
      <div className="flex items-center justify-center min-h-[400px]">
        <p className="text-lg text-muted-foreground">Загрузка информации об ИП...</p>
      </div>
    </div>
  );
}

function EntrepreneurPageError({ error }: { error: Error }) {
  return (
    <div className="container mx-auto p-6">
      <Card>
        <CardContent className="pt-6">
          <div className="text-center space-y-4">
            <h2 className="text-xl font-semibold text-destructive">Ошибка загрузки</h2>
            <p className="text-muted-foreground">
              {error.message || "Не удалось загрузить информацию об ИП"}
            </p>
            <button 
              onClick={() => window.location.reload()} 
              className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
            >
              Попробовать снова
            </button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function EntrepreneurPageNotFound() {
  return (
    <div className="container mx-auto p-6">
      <Card>
        <CardContent className="pt-6">
          <div className="text-center space-y-4">
            <h2 className="text-xl font-semibold">ИП не найден</h2>
            <p className="text-muted-foreground">
              Индивидуальный предприниматель с указанным ОГРНИП не найден в базе данных
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

