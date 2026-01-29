"use client";

import { useRouter } from "next/navigation";
import { Bell, AlertCircle, User } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { SubscriptionsList } from "@/components/subscriptions/subscriptions-list";
import { useMySubscriptionsQuery } from "@/lib/api/subscription-hooks";
import { useAuthStore } from "@/store/auth-store";

export default function WatchlistPage() {
  const router = useRouter();
  const { user, isAuthenticated, isHydrated } = useAuthStore();

  const {
    data,
    isLoading,
    error,
  } = useMySubscriptionsQuery({
    enabled: isAuthenticated,
  });

  // Wait for hydration before checking auth
  if (!isHydrated) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <p className="text-muted-foreground">Загрузка...</p>
      </div>
    );
  }

  // Redirect to login if not authenticated (after hydration)
  if (!isAuthenticated) {
    router.push("/login");
    return null;
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* Заголовок */}
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Bell className="h-8 w-8" />
            Отслеживание изменений
          </h1>
          <p className="text-muted-foreground">
            Управление подписками на уведомления об изменениях в данных ЕГРЮЛ/ЕГРИП
          </p>
        </div>
      </div>

      {/* Информация о текущем пользователе */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">Текущий пользователь</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2">
            <User className="h-4 w-4 text-muted-foreground" />
            <span className="font-medium">
              {user?.firstName} {user?.lastName}
            </span>
            <span className="text-muted-foreground">({user?.email})</span>
          </div>
          <p className="text-sm text-muted-foreground mt-2">
            Уведомления будут приходить на указанный при регистрации email
          </p>
        </CardContent>
      </Card>

      {/* Статистика подписок */}
      {data?.mySubscriptions && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Всего подписок
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-bold">
                {data.mySubscriptions.length}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Активных
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-bold text-green-600">
                {data.mySubscriptions.filter((s) => s.isActive).length}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Приостановлено
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-bold text-yellow-600">
                {data.mySubscriptions.filter((s) => !s.isActive).length}
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Список подписок */}
      <Card>
        <CardHeader>
          <CardTitle>Мои подписки</CardTitle>
          <CardDescription>
            Список компаний и ИП, за изменениями в которых вы следите
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center py-8">
              <p className="text-muted-foreground">Загрузка...</p>
            </div>
          ) : error ? (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Ошибка при загрузке подписок: {error.message}
              </AlertDescription>
            </Alert>
          ) : data?.mySubscriptions && data.mySubscriptions.length > 0 ? (
            <SubscriptionsList subscriptions={data.mySubscriptions} />
          ) : (
            <div className="text-center py-8">
              <p className="text-muted-foreground mb-4">
                У вас пока нет подписок
              </p>
              <Button onClick={() => router.push("/search")}>
                Перейти к поиску
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Инструкция */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Как подписаться на изменения?</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <ol className="list-decimal list-inside space-y-2">
            <li>Найдите компанию или ИП через поиск</li>
            <li>Откройте карточку организации</li>
            <li>
              Нажмите кнопку{" "}
              <span className="inline-flex items-center gap-1 px-2 py-1 bg-muted rounded">
                <Bell className="h-3 w-3" />
                Отслеживать изменения
              </span>
            </li>
            <li>Выберите типы изменений, о которых хотите получать уведомления</li>
            <li>Подтвердите подписку</li>
          </ol>
        </CardContent>
      </Card>
    </div>
  );
}
