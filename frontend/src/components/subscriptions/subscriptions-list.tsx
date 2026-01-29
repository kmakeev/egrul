"use client";

import { Building2, User, Trash2, BellOff, BellRing } from "lucide-react";
import Link from "next/link";
import { useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { useToast } from "@/hooks/use-toast";
import {
  type EntitySubscription,
  EntityType,
  useDeleteSubscriptionMutation,
  useToggleSubscriptionMutation,
} from "@/lib/api/subscription-hooks";

interface SubscriptionsListProps {
  subscriptions: EntitySubscription[];
}

export function SubscriptionsList({ subscriptions }: SubscriptionsListProps) {
  const { toast } = useToast();
  const queryClient = useQueryClient();

  const deleteSubscription = useDeleteSubscriptionMutation({
    onSuccess: () => {
      // Явно инвалидируем кэш для немедленного обновления UI
      queryClient.invalidateQueries({
        queryKey: ["subscriptions", "my"],
      });
      queryClient.invalidateQueries({
        queryKey: ["subscription"],
      });
      toast({
        title: "Подписка удалена",
        description: "Вы больше не будете получать уведомления",
      });
    },
    onError: (error) => {
      toast({
        title: "Ошибка при удалении",
        description: error.message,
        variant: "destructive",
      });
    },
  });

  const toggleSubscription = useToggleSubscriptionMutation({
    onSuccess: (data) => {
      // Явно инвалидируем кэш для немедленного обновления UI
      queryClient.invalidateQueries({
        queryKey: ["subscriptions"],
      });
      queryClient.invalidateQueries({
        queryKey: ["subscription"],
      });
      toast({
        title: data.toggleSubscription.isActive
          ? "Подписка включена"
          : "Подписка приостановлена",
        description: data.toggleSubscription.isActive
          ? "Уведомления возобновлены"
          : "Уведомления временно отключены",
      });
    },
    onError: (error) => {
      toast({
        title: "Ошибка",
        description: error.message,
        variant: "destructive",
      });
    },
  });

  const handleDelete = (id: string) => {
    deleteSubscription.mutate(id);
  };

  const handleToggle = (id: string, currentStatus: boolean) => {
    toggleSubscription.mutate({ id, isActive: !currentStatus });
  };

  if (subscriptions.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <BellOff className="h-12 w-12 text-muted-foreground mb-4" />
          <p className="text-lg font-medium text-muted-foreground">
            У вас пока нет подписок
          </p>
          <p className="text-sm text-muted-foreground mt-1">
            Подпишитесь на изменения в карточках компаний или ИП
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {subscriptions.map((subscription) => {
        const isCompany = subscription.entityType === EntityType.COMPANY;
        const entityUrl = isCompany
          ? `/company/${subscription.entityId}`
          : `/entrepreneur/${subscription.entityId}`;

        const activeFilters = subscription.changeFilters
          ? Object.entries(subscription.changeFilters)
              .filter(([_, value]) => value)
              .map(([key]) => key)
          : [];

        return (
          <Card
            key={subscription.id}
            className={subscription.isActive ? "" : "opacity-60"}
          >
            <CardHeader className="pb-3">
              <div className="flex items-start justify-between">
                <div className="flex items-start gap-3 flex-1">
                  <div className="p-2 rounded-lg bg-primary/10">
                    {isCompany ? (
                      <Building2 className="h-5 w-5 text-primary" />
                    ) : (
                      <User className="h-5 w-5 text-primary" />
                    )}
                  </div>
                  <div className="flex-1">
                    <Link
                      href={entityUrl}
                      className="hover:underline cursor-pointer"
                    >
                      <CardTitle className="text-base mb-1">
                        {subscription.entityName}
                      </CardTitle>
                    </Link>
                    <p className="text-sm text-muted-foreground">
                      {isCompany ? "ОГРН" : "ОГРНИП"}: {subscription.entityId}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  {subscription.isActive ? (
                    <Badge variant="default" className="gap-1">
                      <BellRing className="h-3 w-3" />
                      Активна
                    </Badge>
                  ) : (
                    <Badge variant="secondary" className="gap-1">
                      <BellOff className="h-3 w-3" />
                      Приостановлена
                    </Badge>
                  )}
                </div>
              </div>
            </CardHeader>

            <CardContent className="space-y-3">
              {/* Фильтры */}
              <div>
                <p className="text-sm font-medium mb-2">
                  Отслеживаемые изменения:
                </p>
                <div className="flex flex-wrap gap-1">
                  {activeFilters.length > 0 ? (
                    activeFilters.map((filter) => (
                      <Badge key={filter} variant="outline" className="text-xs">
                        {getFilterLabel(filter)}
                      </Badge>
                    ))
                  ) : (
                    <span className="text-sm text-muted-foreground">
                      Не выбрано
                    </span>
                  )}
                </div>
              </div>

              {/* Метаданные */}
              <div className="flex items-center justify-between text-xs text-muted-foreground pt-2 border-t">
                <span>
                  Создана: {new Date(subscription.createdAt).toLocaleDateString("ru-RU")}
                </span>
                {subscription.lastNotifiedAt && (
                  <span>
                    Последнее уведомление:{" "}
                    {new Date(subscription.lastNotifiedAt).toLocaleDateString("ru-RU")}
                  </span>
                )}
              </div>

              {/* Действия */}
              <div className="flex gap-2 pt-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    handleToggle(subscription.id, subscription.isActive)
                  }
                  disabled={toggleSubscription.isPending}
                >
                  {subscription.isActive ? (
                    <>
                      <BellOff className="h-4 w-4 mr-1" />
                      Приостановить
                    </>
                  ) : (
                    <>
                      <BellRing className="h-4 w-4 mr-1" />
                      Возобновить
                    </>
                  )}
                </Button>

                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={deleteSubscription.isPending}
                    >
                      <Trash2 className="h-4 w-4 mr-1" />
                      Удалить
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>Удалить подписку?</AlertDialogTitle>
                      <AlertDialogDescription>
                        Вы больше не будете получать уведомления об изменениях
                        в {subscription.entityName}. Это действие нельзя
                        отменить.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>Отмена</AlertDialogCancel>
                      <AlertDialogAction
                        onClick={() => handleDelete(subscription.id)}
                      >
                        Удалить
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}

function getFilterLabel(filter: string): string {
  const labels: Record<string, string> = {
    status: "Статус",
    director: "Руководитель",
    founders: "Учредители",
    address: "Адрес",
    capital: "Капитал",
    activities: "ОКВЭД",
  };
  return labels[filter] || filter;
}
