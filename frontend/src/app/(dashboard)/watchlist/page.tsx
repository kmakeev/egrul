"use client";

import { useMemo, useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Bell, AlertCircle } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { SubscriptionsList } from "@/components/subscriptions/subscriptions-list";
import { SubscriptionHelpDialog } from "@/components/subscriptions/subscription-help-dialog";
import { FavoritesList } from "@/components/favorites/favorites-list";
import { SearchPagination } from "@/components/search/search-pagination";
import { useMySubscriptionsQuery } from "@/lib/api/subscription-hooks";
import { useMyFavoritesQuery } from "@/lib/api/favorites-hooks";
import { useAuthStore } from "@/store/auth-store";

export default function WatchlistPage() {
  const router = useRouter();
  const { isAuthenticated, isHydrated } = useAuthStore();

  const {
    data,
    isLoading,
    error,
  } = useMySubscriptionsQuery({
    enabled: isAuthenticated,
  });

  const {
    data: favoritesData,
    isLoading: favoritesLoading,
    error: favoritesError,
  } = useMyFavoritesQuery({
    enabled: isAuthenticated,
  });

  // Вкладки
  const [activeTab, setActiveTab] = useState<"subscriptions" | "favorites">("subscriptions");

  // Фильтрация для подписок
  const [searchQuery, setSearchQuery] = useState("");
  const [typeFilter, setTypeFilter] = useState<"all" | "COMPANY" | "ENTREPRENEUR">("all");
  const [statusFilter, setStatusFilter] = useState<"all" | "active" | "paused">("all");

  // Пагинация для подписок
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // Фильтрация для избранного
  const [favSearchQuery, setFavSearchQuery] = useState("");
  const [favTypeFilter, setFavTypeFilter] = useState<"all" | "COMPANY" | "ENTREPRENEUR">("all");

  // Пагинация для избранного
  const [favPage, setFavPage] = useState(1);
  const [favPageSize, setFavPageSize] = useState(10);

  // Логика фильтрации
  const filteredSubscriptions = useMemo(() => {
    if (!data?.mySubscriptions) return [];

    return data.mySubscriptions.filter((sub) => {
      // Поиск по имени
      const matchesSearch =
        searchQuery === "" ||
        sub.entityName.toLowerCase().includes(searchQuery.toLowerCase());

      // Фильтр по типу
      const matchesType = typeFilter === "all" || sub.entityType === typeFilter;

      // Фильтр по статусу
      const matchesStatus =
        statusFilter === "all" ||
        (statusFilter === "active" && sub.isActive) ||
        (statusFilter === "paused" && !sub.isActive);

      return matchesSearch && matchesType && matchesStatus;
    });
  }, [data?.mySubscriptions, searchQuery, typeFilter, statusFilter]);

  // Логика пагинации
  const paginatedSubscriptions = useMemo(() => {
    const start = (page - 1) * pageSize;
    const end = start + pageSize;
    return filteredSubscriptions.slice(start, end);
  }, [filteredSubscriptions, page, pageSize]);

  // Логика фильтрации для избранного
  const filteredFavorites = useMemo(() => {
    if (!favoritesData?.myFavorites) return [];

    return favoritesData.myFavorites.filter((fav) => {
      // Поиск по имени
      const matchesSearch =
        favSearchQuery === "" ||
        fav.entityName.toLowerCase().includes(favSearchQuery.toLowerCase());

      // Фильтр по типу
      const matchesType = favTypeFilter === "all" || fav.entityType === favTypeFilter;

      return matchesSearch && matchesType;
    });
  }, [favoritesData?.myFavorites, favSearchQuery, favTypeFilter]);

  // Логика пагинации для избранного
  const paginatedFavorites = useMemo(() => {
    const start = (favPage - 1) * favPageSize;
    const end = start + favPageSize;
    return filteredFavorites.slice(start, end);
  }, [filteredFavorites, favPage, favPageSize]);

  // Сброс страницы при изменении фильтров подписок
  useEffect(() => {
    setPage(1);
  }, [searchQuery, typeFilter, statusFilter]);

  // Сброс страницы при изменении фильтров избранного
  useEffect(() => {
    setFavPage(1);
  }, [favSearchQuery, favTypeFilter]);

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

        {/* Иконка помощи */}
        <SubscriptionHelpDialog />
      </div>

      {/* Вкладки */}
      <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as "subscriptions" | "favorites")}>
        <TabsList className="grid w-full max-w-md grid-cols-2">
          <TabsTrigger value="subscriptions">
            Подписки {data?.mySubscriptions && `(${data.mySubscriptions.length})`}
          </TabsTrigger>
          <TabsTrigger value="favorites">
            Избранное {favoritesData?.myFavorites && `(${favoritesData.myFavorites.length})`}
          </TabsTrigger>
        </TabsList>

        {/* Вкладка: Подписки */}
        <TabsContent value="subscriptions" className="space-y-6">
          {/* Фильтры */}
          <Card>
            <CardContent className="pt-6">
              <div className="flex flex-col sm:flex-row gap-4">
                {/* Поиск по имени */}
                <div className="flex-1">
                  <Input
                    placeholder="Поиск по названию..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                  />
                </div>

                {/* Фильтр по типу */}
                <Select value={typeFilter} onValueChange={(value: "all" | "COMPANY" | "ENTREPRENEUR") => setTypeFilter(value)}>
                  <SelectTrigger className="w-full sm:w-[180px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Все типы</SelectItem>
                    <SelectItem value="COMPANY">Компании</SelectItem>
                    <SelectItem value="ENTREPRENEUR">ИП</SelectItem>
                  </SelectContent>
                </Select>

                {/* Фильтр по статусу */}
                <Select value={statusFilter} onValueChange={(value: "all" | "active" | "paused") => setStatusFilter(value)}>
                  <SelectTrigger className="w-full sm:w-[180px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Все статусы</SelectItem>
                    <SelectItem value="active">Активные</SelectItem>
                    <SelectItem value="paused">Приостановленные</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>

          {/* Статистика подписок */}
          {data?.mySubscriptions && (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium text-muted-foreground">
                    {searchQuery || typeFilter !== "all" || statusFilter !== "all"
                      ? "Найдено подписок"
                      : "Всего подписок"}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-3xl font-bold">
                    {filteredSubscriptions.length}
                    {(searchQuery || typeFilter !== "all" || statusFilter !== "all") && (
                      <span className="text-sm font-normal text-muted-foreground ml-2">
                        из {data.mySubscriptions.length}
                      </span>
                    )}
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
                    {filteredSubscriptions.filter((s) => s.isActive).length}
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
                    {filteredSubscriptions.filter((s) => !s.isActive).length}
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
              ) : filteredSubscriptions.length > 0 ? (
                <>
                  <SubscriptionsList subscriptions={paginatedSubscriptions} />
                  {filteredSubscriptions.length > pageSize && (
                    <div className="mt-4">
                      <SearchPagination
                        page={page}
                        pageSize={pageSize}
                        total={filteredSubscriptions.length}
                        onPageChange={setPage}
                        onPageSizeChange={(size) => {
                          setPageSize(size);
                          setPage(1);
                        }}
                      />
                    </div>
                  )}
                </>
              ) : data?.mySubscriptions && data.mySubscriptions.length > 0 ? (
                <div className="text-center py-8">
                  <p className="text-muted-foreground mb-4">
                    Подписок не найдено. Попробуйте изменить фильтры.
                  </p>
                </div>
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
        </TabsContent>

        {/* Вкладка: Избранное */}
        <TabsContent value="favorites" className="space-y-6">
          {/* Фильтры */}
          <Card>
            <CardContent className="pt-6">
              <div className="flex flex-col sm:flex-row gap-4">
                {/* Поиск по имени */}
                <div className="flex-1">
                  <Input
                    placeholder="Поиск по названию..."
                    value={favSearchQuery}
                    onChange={(e) => setFavSearchQuery(e.target.value)}
                  />
                </div>

                {/* Фильтр по типу */}
                <Select value={favTypeFilter} onValueChange={(value: "all" | "COMPANY" | "ENTREPRENEUR") => setFavTypeFilter(value)}>
                  <SelectTrigger className="w-full sm:w-[180px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Все типы</SelectItem>
                    <SelectItem value="COMPANY">Компании</SelectItem>
                    <SelectItem value="ENTREPRENEUR">ИП</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>

          {/* Статистика избранного */}
          {favoritesData?.myFavorites && (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium text-muted-foreground">
                    {favSearchQuery || favTypeFilter !== "all"
                      ? "Найдено записей"
                      : "Всего записей"}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-3xl font-bold">
                    {filteredFavorites.length}
                    {(favSearchQuery || favTypeFilter !== "all") && (
                      <span className="text-sm font-normal text-muted-foreground ml-2">
                        из {favoritesData.myFavorites.length}
                      </span>
                    )}
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium text-muted-foreground">
                    Компаний
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-3xl font-bold text-blue-600">
                    {filteredFavorites.filter((f) => f.entityType === "COMPANY").length}
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium text-muted-foreground">
                    ИП
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-3xl font-bold text-purple-600">
                    {filteredFavorites.filter((f) => f.entityType === "ENTREPRENEUR").length}
                  </p>
                </CardContent>
              </Card>
            </div>
          )}

          {/* Список избранного */}
          <Card>
            <CardHeader>
              <CardTitle>Избранное</CardTitle>
              <CardDescription>
                Список сохраненных компаний и ИП
              </CardDescription>
            </CardHeader>
            <CardContent>
              {favoritesLoading ? (
                <div className="flex justify-center py-8">
                  <p className="text-muted-foreground">Загрузка...</p>
                </div>
              ) : favoritesError ? (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>
                    Ошибка при загрузке избранного: {favoritesError.message}
                  </AlertDescription>
                </Alert>
              ) : filteredFavorites.length > 0 ? (
                <>
                  <FavoritesList favorites={paginatedFavorites} />
                  {filteredFavorites.length > favPageSize && (
                    <div className="mt-4">
                      <SearchPagination
                        page={favPage}
                        pageSize={favPageSize}
                        total={filteredFavorites.length}
                        onPageChange={setFavPage}
                        onPageSizeChange={(size) => {
                          setFavPageSize(size);
                          setFavPage(1);
                        }}
                      />
                    </div>
                  )}
                </>
              ) : favoritesData?.myFavorites && favoritesData.myFavorites.length > 0 ? (
                <div className="text-center py-8">
                  <p className="text-muted-foreground mb-4">
                    Записей не найдено. Попробуйте изменить фильтры.
                  </p>
                </div>
              ) : (
                <div className="text-center py-8">
                  <p className="text-muted-foreground mb-4">
                    У вас пока нет избранных записей
                  </p>
                  <Button onClick={() => router.push("/search")}>
                    Перейти к поиску
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
