"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useStatisticsQuery } from "@/lib/api/hooks";

export default function AnalyticsPage() {
  const { data, isLoading, error } = useStatisticsQuery({ filter: {} }, {
    staleTime: 5 * 60 * 1000,
  });

  if (isLoading) {
    return (
      <div className="container mx-auto p-6">
        <p>Загрузка...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto p-6">
        <Card>
          <CardContent className="pt-6">
            <p className="text-destructive">
              Ошибка загрузки статистики: {error instanceof Error ? error.message : "Неизвестная ошибка"}
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  const stats = data?.statistics;

  return (
    <div className="container mx-auto p-6 space-y-6">
      <h1 className="mb-6 text-3xl font-bold">Аналитика</h1>

      {!stats ? (
        <p className="text-muted-foreground">Данные статистики недоступны</p>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          <Card>
            <CardHeader>
              <CardTitle>Всего компаний</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-semibold">{stats.totalCompanies.toLocaleString("ru-RU")}</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Активные компании</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-semibold">{stats.activeCompanies.toLocaleString("ru-RU")}</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Ликвидированные компании</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-semibold">{stats.liquidatedCompanies.toLocaleString("ru-RU")}</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Всего ИП</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-semibold">{stats.totalEntrepreneurs.toLocaleString("ru-RU")}</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Активные ИП</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-semibold">{stats.activeEntrepreneurs.toLocaleString("ru-RU")}</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Ликвидированные ИП</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-semibold">{stats.liquidatedEntrepreneurs.toLocaleString("ru-RU")}</p>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}

