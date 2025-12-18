"use client";

import { useParams } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useEntrepreneurQuery } from "@/lib/api/hooks";

export default function EntrepreneurPage() {
  const params = useParams();
  const ogrnip = params.ogrnip as string;

  const { data, isLoading, error } = useEntrepreneurQuery(ogrnip);

  const entrepreneur = data?.entrepreneur;

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

  if (!entrepreneur) {
    return null;
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>
            {entrepreneur.lastName} {entrepreneur.firstName} {entrepreneur.middleName}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-sm text-muted-foreground">ОГРНИП</p>
              <p className="font-medium">{entrepreneur.ogrnip}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">ИНН</p>
              <p className="font-medium">{entrepreneur.inn}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Статус</p>
              <p className="font-medium">{entrepreneur.status}</p>
            </div>
            {entrepreneur.registrationDate && (
              <div>
                <p className="text-sm text-muted-foreground">Дата регистрации</p>
                <p className="font-medium">
                  {new Date(entrepreneur.registrationDate).toLocaleDateString("ru-RU")}
                </p>
              </div>
            )}
          </div>
          {entrepreneur.mainActivity && (
            <div>
              <p className="text-sm text-muted-foreground mb-2">Основной вид деятельности</p>
              <p className="font-medium">
                {entrepreneur.mainActivity.code} - {entrepreneur.mainActivity.name}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

