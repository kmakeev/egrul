"use client";

import { useParams } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useCompanyQuery } from "@/lib/api/hooks";

export default function CompanyPage() {
  const params = useParams();
  const ogrn = params.ogrn as string;

  const { data, isLoading, error } = useCompanyQuery(ogrn);

  const company = data?.company;

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
              Ошибка: {error instanceof Error ? error.message : "Неизвестная ошибка"}
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!company) {
    return null;
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>{company.fullName}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-sm text-muted-foreground">ОГРН</p>
              <p className="font-medium">{company.ogrn}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">ИНН</p>
              <p className="font-medium">{company.inn}</p>
            </div>
            {company.kpp && (
              <div>
                <p className="text-sm text-muted-foreground">КПП</p>
                <p className="font-medium">{company.kpp}</p>
              </div>
            )}
            <div>
              <p className="text-sm text-muted-foreground">Статус</p>
              <p className="font-medium">{company.status}</p>
            </div>
            {company.registrationDate && (
              <div>
                <p className="text-sm text-muted-foreground">Дата регистрации</p>
                <p className="font-medium">
                  {new Date(company.registrationDate).toLocaleDateString("ru-RU")}
                </p>
              </div>
            )}
          </div>
          {company.address && (
            <div>
              <p className="text-sm text-muted-foreground mb-2">Адрес</p>
              <p className="font-medium">
                {company.address.fullAddress ||
                  `${company.address.region || ""} ${company.address.city || ""} ${company.address.street || ""} ${company.address.house || ""}`.trim()}
              </p>
            </div>
          )}
          {company.mainActivity && (
            <div>
              <p className="text-sm text-muted-foreground mb-2">Основной вид деятельности</p>
              <p className="font-medium">
                {company.mainActivity.code} - {company.mainActivity.name}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

