"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Heart, Download, Share2, Bell, BellRing } from "lucide-react";
import { decodeHtmlEntities } from "@/lib/html-utils";
import { formatDate } from "@/lib/format-utils";
import { CompanyStatusBadge } from "./company-status-badge";
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
import type { LegalEntity } from "@/lib/api";

interface CompanyHeaderProps {
  company: LegalEntity;
}

export function CompanyHeader({ company }: CompanyHeaderProps) {
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
    EntityType.COMPANY,
    company.ogrn,
    { enabled: isAuthenticated }
  );

  const hasSubscription = hasSubscriptionData?.hasSubscription ?? false;

  // Check if in favorites
  const { data: hasFavoriteData } = useHasFavoriteQuery(
    EntityType.COMPANY,
    company.ogrn,
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
        queryKey: ["favorite", "has", EntityType.COMPANY, company.ogrn],
      });
      queryClient.invalidateQueries({
        queryKey: ["favorites", "my"],
      });
      toast({
        title: "Добавлено в избранное",
        description: "Компания добавлена в избранное"
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
        queryKey: ["favorite", "has", EntityType.COMPANY, company.ogrn],
      });
      queryClient.invalidateQueries({
        queryKey: ["favorites", "my"],
      });
      toast({
        title: "Удалено из избранного",
        description: "Компания удалена из избранного"
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

    if (isFavorite) {
      // Удаляем из избранного
      const favorite = favoritesData?.myFavorites?.find(
        f => f.entityType === EntityType.COMPANY && f.entityId === company.ogrn
      );
      if (favorite) {
        deleteFavorite.mutate(favorite.id);
      }
    } else {
      // Добавляем в избранное
      createFavorite.mutate({
        entityType: EntityType.COMPANY,
        entityId: company.ogrn,
        entityName: primaryName,
        notes: null
      });
    }
  };

  const handleDownloadExtract = () => {
    // TODO: Implement extract download
    console.log("Download extract:", company.ogrn);
  };

  const handleShare = () => {
    // TODO: Implement share functionality
    const decodedName = decodeHtmlEntities(company.fullName || company.shortName || "");
    if (navigator.share) {
      navigator.share({
        title: decodedName,
        text: `Информация о компании ${decodedName}`,
        url: window.location.href,
      });
    } else {
      navigator.clipboard.writeText(window.location.href);
    }
  };

  // Декодируем HTML-сущности в названиях
  const decodedFullName = company.fullName ? decodeHtmlEntities(company.fullName) : "";
  const decodedShortName = company.shortName ? decodeHtmlEntities(company.shortName) : null;

  // Определяем что показывать как основное название
  const primaryName = decodedFullName || decodedShortName || "Название не указано";
  const secondaryName = decodedFullName && decodedShortName && decodedFullName !== decodedShortName
    ? decodedShortName
    : null;

  return (
    <Card>
      <CardHeader className="pb-4">
        <div className="flex flex-col lg:flex-row lg:items-start lg:justify-between gap-4">
          <div className="flex-1 min-w-0">
            <div className="mb-2">
              <CompanyStatusBadge company={company} />
            </div>
            <div className="flex items-center gap-3 mb-2">
              <h1 className="text-2xl font-bold text-white break-words" style={{maxWidth: '100%', wordWrap: 'break-word'}}>
                {primaryName}
              </h1>
            </div>
            {secondaryName && (
              <p className="text-lg text-gray-200 mb-3 break-words">{secondaryName}</p>
            )}
            {company.registrationDate && (
              <p className="text-sm text-gray-300">
                Дата регистрации: {formatDate(company.registrationDate)}
              </p>
            )}
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
              {isFavorite ? "В избранном" : "В избранное"}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleDownloadExtract}
              className="flex items-center justify-center gap-2 w-full sm:w-auto"
            >
              <Download className="h-4 w-4" />
              Скачать выписку
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleShare}
              className="flex items-center justify-center gap-2 w-full sm:w-auto"
            >
              <Share2 className="h-4 w-4" />
              Поделиться
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
                  Подписка активна
                </>
              ) : (
                <>
                  <Bell className="h-4 w-4" />
                  Отслеживать изменения
                </>
              )}
            </Button>
          </div>
        </div>
      </CardHeader>

      <CardContent>
        <div className="grid grid-cols-3 gap-6">
          <div>
            <p className="text-sm text-gray-500 mb-1">ОГРН</p>
            <p className="font-mono text-lg font-semibold">{company.ogrn}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500 mb-1">ИНН</p>
            <p className="font-mono text-lg font-semibold">{company.inn}</p>
          </div>
          {company.kpp && (
            <div>
              <p className="text-sm text-gray-500 mb-1">КПП</p>
              <p className="font-mono text-lg font-semibold">{company.kpp}</p>
            </div>
          )}
        </div>
      </CardContent>

      <SubscriptionForm
        entityType="company"
        entityId={company.ogrn}
        entityName={primaryName}
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
