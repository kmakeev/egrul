"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Card, CardHeader } from "@/components/ui/card";
import { Heart, Download, Share2, Bell, BellRing } from "lucide-react";
import { formatDate } from "@/lib/format-utils";
import { EntrepreneurStatusBadge } from "./entrepreneur-status-badge";
import { SubscriptionForm } from "@/components/subscriptions/subscription-form";
import { LoginDialog } from "@/components/auth/login-dialog";
import { useHasSubscriptionQuery } from "@/lib/api/subscription-hooks";
import { EntityType } from "@/lib/api/subscription-hooks";
import {
  useHasFavoriteQuery,
  useCreateFavoriteMutation,
  useDeleteFavoriteMutation,
  useMyFavoritesQuery
} from "@/lib/api/favorites-hooks";
import { useAuthStore } from "@/store/auth-store";
import { useToast } from "@/hooks/use-toast";
import type { IndividualEntrepreneur } from "@/lib/api";

interface EntrepreneurHeaderProps {
  entrepreneur: IndividualEntrepreneur;
}

export function EntrepreneurHeader({ entrepreneur }: EntrepreneurHeaderProps) {
  const { toast } = useToast();
  const queryClient = useQueryClient();

  // State for subscription modal
  const [subscriptionDialogOpen, setSubscriptionDialogOpen] = useState(false);

  // State for login modal
  const [loginDialogOpen, setLoginDialogOpen] = useState(false);

  // Get user from auth store
  const { isAuthenticated } = useAuthStore();

  // Check if already subscribed
  const { data: hasSubscriptionData } = useHasSubscriptionQuery(
    EntityType.ENTREPRENEUR,
    entrepreneur.ogrnip,
    { enabled: isAuthenticated }
  );

  const hasSubscription = hasSubscriptionData?.hasSubscription ?? false;

  // Check if in favorites
  const { data: hasFavoriteData } = useHasFavoriteQuery(
    EntityType.ENTREPRENEUR,
    entrepreneur.ogrnip,
    { enabled: isAuthenticated }
  );

  const isFavorite = hasFavoriteData?.hasFavorite ?? false;

  // Get all favorites to find favorite ID for deletion
  const { data: favoritesData } = useMyFavoritesQuery({
    enabled: isAuthenticated && isFavorite
  });

  // Mutations for favorites
  const createFavorite = useCreateFavoriteMutation({
    onSuccess: () => {
      // Явно инвалидируем кэш для немедленного обновления UI
      queryClient.invalidateQueries({
        queryKey: ["favorite", "has", EntityType.ENTREPRENEUR, entrepreneur.ogrnip],
      });
      queryClient.invalidateQueries({
        queryKey: ["favorites", "my"],
      });
      toast({
        title: "Добавлено в избранное",
        description: "ИП добавлен в избранное"
      });
    },
    onError: (error: unknown) => {
      const message = error instanceof Error ? error.message : "Не удалось добавить в избранное";
      toast({
        title: "Ошибка",
        description: message,
        variant: "destructive"
      });
    }
  });

  const deleteFavorite = useDeleteFavoriteMutation({
    onSuccess: () => {
      // Явно инвалидируем кэш для немедленного обновления UI
      queryClient.invalidateQueries({
        queryKey: ["favorite", "has", EntityType.ENTREPRENEUR, entrepreneur.ogrnip],
      });
      queryClient.invalidateQueries({
        queryKey: ["favorites", "my"],
      });
      toast({
        title: "Удалено из избранного",
        description: "ИП удален из избранного"
      });
    },
    onError: (error: unknown) => {
      const message = error instanceof Error ? error.message : "Не удалось удалить из избранного";
      toast({
        title: "Ошибка",
        description: message,
        variant: "destructive"
      });
    }
  });

  const handleSubscribeClick = () => {
    if (!isAuthenticated) {
      setLoginDialogOpen(true);
      return;
    }
    setSubscriptionDialogOpen(true);
  };

  const handleAddToFavorites = () => {
    if (!isAuthenticated) {
      setLoginDialogOpen(true);
      return;
    }

    // Формируем полное имя для сохранения
    const fullName = `${entrepreneur.lastName} ${entrepreneur.firstName} ${entrepreneur.middleName || ""}`.trim();

    if (isFavorite) {
      // Удаляем из избранного
      const favorite = favoritesData?.myFavorites?.find(
        f => f.entityType === EntityType.ENTREPRENEUR && f.entityId === entrepreneur.ogrnip
      );
      if (favorite) {
        deleteFavorite.mutate(favorite.id);
      }
    } else {
      // Добавляем в избранное
      createFavorite.mutate({
        entityType: EntityType.ENTREPRENEUR,
        entityId: entrepreneur.ogrnip,
        entityName: fullName,
        notes: null
      });
    }
  };

  const handleDownloadExtract = () => {
    // TODO: Implement extract download
    console.log("Download extract:", entrepreneur.ogrnip);
  };

  const handleShare = () => {
    // TODO: Implement share functionality
    const fullName = `${entrepreneur.lastName} ${entrepreneur.firstName} ${entrepreneur.middleName || ""}`.trim();
    if (navigator.share) {
      navigator.share({
        title: fullName,
        text: `Информация об ИП ${fullName}`,
        url: window.location.href,
      });
    } else {
      navigator.clipboard.writeText(window.location.href);
    }
  };

  // Формируем полное имя ИП
  const fullName = `${entrepreneur.lastName} ${entrepreneur.firstName} ${entrepreneur.middleName || ""}`.trim();

  return (
    <Card>
      <CardHeader className="pb-4">
        <div className="flex flex-col lg:flex-row lg:items-start lg:justify-between gap-4">
          <div className="flex-1 min-w-0">
            <div className="mb-2">
              <EntrepreneurStatusBadge entrepreneur={entrepreneur} />
            </div>
            <div className="flex items-center gap-3 mb-2">
              <h1 className="text-2xl font-bold text-white break-words" style={{maxWidth: '100%', wordWrap: 'break-word'}}>
                {fullName}
              </h1>
            </div>
            <div className="space-y-1">
              <p className="text-sm text-gray-300">
                ОГРНИП: {entrepreneur.ogrnip}
              </p>
              <p className="text-sm text-gray-300">
                ИНН: {entrepreneur.inn}
              </p>
              {entrepreneur.registrationDate && (
                <p className="text-sm text-gray-300">
                  Дата регистрации: {formatDate(entrepreneur.registrationDate)}
                </p>
              )}
            </div>
          </div>

          <div className="flex flex-col sm:flex-row gap-2">
            <Button
              variant={isFavorite ? "default" : "outline"}
              size="sm"
              onClick={handleAddToFavorites}
              className="flex items-center justify-center gap-2 w-full sm:w-auto"
              disabled={createFavorite.isPending || deleteFavorite.isPending}
            >
              <Heart className={`h-4 w-4 ${isFavorite ? "fill-current" : ""}`} />
              <span className="hidden sm:inline">{isFavorite ? "В избранном" : "В избранное"}</span>
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleDownloadExtract}
              className="flex items-center justify-center gap-2 w-full sm:w-auto"
            >
              <Download className="h-4 w-4" />
              <span className="hidden sm:inline">Выписка</span>
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleShare}
              className="flex items-center justify-center gap-2 w-full sm:w-auto"
            >
              <Share2 className="h-4 w-4" />
              <span className="hidden sm:inline">Поделиться</span>
            </Button>
            <Button
              variant={hasSubscription ? "default" : "outline"}
              size="sm"
              onClick={handleSubscribeClick}
              className="flex items-center justify-center gap-2 w-full sm:w-auto"
            >
              {hasSubscription ? (
                <>
                  <BellRing className="h-4 w-4 fill-current" />
                  <span className="hidden sm:inline">Подписка активна</span>
                </>
              ) : (
                <>
                  <Bell className="h-4 w-4" />
                  <span className="hidden sm:inline">Отслеживать</span>
                </>
              )}
            </Button>
          </div>
        </div>
      </CardHeader>

      <SubscriptionForm
        entityType="entrepreneur"
        entityId={entrepreneur.ogrnip}
        entityName={fullName}
        open={subscriptionDialogOpen}
        onOpenChange={setSubscriptionDialogOpen}
      />

      <LoginDialog
        open={loginDialogOpen}
        onOpenChange={setLoginDialogOpen}
      />
    </Card>
  );
}
