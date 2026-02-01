"use client";

import { useState, useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Settings, Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useToast } from "@/hooks/use-toast";
import {
  type EntitySubscription,
  type ChangeFilters,
  useUpdateSubscriptionFiltersMutation,
} from "@/lib/api/subscription-hooks";

interface EditSubscriptionFiltersDialogProps {
  subscription: EntitySubscription | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const FILTER_LABELS: Record<keyof ChangeFilters, string> = {
  status: "Статус (ликвидация, реорганизация)",
  director: "Руководитель",
  founders: "Учредители",
  address: "Адрес",
  capital: "Уставный капитал",
  activities: "Виды деятельности (ОКВЭД)",
};

export function EditSubscriptionFiltersDialog({
  subscription,
  open,
  onOpenChange,
}: EditSubscriptionFiltersDialogProps) {
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<ChangeFilters>({
    status: false,
    director: false,
    founders: false,
    address: false,
    capital: false,
    activities: false,
  });

  // Загрузка фильтров из подписки при монтировании или изменении
  useEffect(() => {
    if (subscription?.changeFilters) {
      setFilters(subscription.changeFilters);
    }
  }, [subscription]);

  const updateFilters = useUpdateSubscriptionFiltersMutation({
    onSuccess: () => {
      // Явно инвалидируем кэш для немедленного обновления UI
      queryClient.invalidateQueries({
        queryKey: ["subscription", subscription?.id],
      });
      queryClient.invalidateQueries({
        queryKey: ["subscriptions", "my"],
      });
      toast({
        title: "Фильтры обновлены",
        description: "Изменения сохранены успешно",
      });
      onOpenChange(false);
    },
    onError: (error: unknown) => {
      const message = error instanceof Error ? error.message : "Не удалось обновить фильтры";
      toast({
        title: "Ошибка при обновлении фильтров",
        description: message,
        variant: "destructive",
      });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!subscription) return;

    updateFilters.mutate({
      id: subscription.id,
      changeFilters: filters,
    });
  };

  const handleFilterChange = (key: keyof ChangeFilters, checked: boolean) => {
    setFilters((prev) => ({
      ...prev,
      [key]: checked,
    }));
  };

  const selectAll = () => {
    setFilters({
      status: true,
      director: true,
      founders: true,
      address: true,
      capital: true,
      activities: true,
    });
  };

  const deselectAll = () => {
    setFilters({
      status: false,
      director: false,
      founders: false,
      address: false,
      capital: false,
      activities: false,
    });
  };

  const selectedCount = Object.values(filters).filter(Boolean).length;

  if (!subscription) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Settings className="h-5 w-5" />
            Редактирование подписки
          </DialogTitle>
          <DialogDescription>
            Измените типы отслеживаемых изменений
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Название организации */}
          <div className="rounded-lg border bg-muted/50 p-3">
            <p className="text-sm font-medium">{subscription.entityName}</p>
            <p className="text-xs text-muted-foreground mt-1">
              {subscription.entityType === "COMPANY" ? "ОГРН" : "ОГРНИП"}:{" "}
              {subscription.entityId}
            </p>
          </div>

          {/* Фильтры */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <Label>Типы изменений ({selectedCount} выбрано)</Label>
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={selectAll}
                  className="h-7 text-xs"
                >
                  Выбрать все
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={deselectAll}
                  className="h-7 text-xs"
                >
                  Снять все
                </Button>
              </div>
            </div>

            <div className="space-y-2 rounded-lg border p-3">
              {(Object.keys(filters) as Array<keyof ChangeFilters>).map(
                (key) => (
                  <div key={key} className="flex items-center gap-2">
                    <Checkbox
                      id={`filter-${key}`}
                      checked={filters[key]}
                      onCheckedChange={(checked) =>
                        handleFilterChange(key, checked === true)
                      }
                    />
                    <Label
                      htmlFor={`filter-${key}`}
                      className="text-sm font-normal cursor-pointer flex items-center gap-2"
                    >
                      {filters[key] && (
                        <Check className="h-3 w-3 text-green-600" />
                      )}
                      {FILTER_LABELS[key]}
                    </Label>
                  </div>
                )
              )}
            </div>

            <p className="text-xs text-muted-foreground">
              {selectedCount === 0
                ? "Подписка будет сохранена без отслеживания изменений (только в избранном)"
                : "Уведомления будут приходить на email, указанный при регистрации"
              }
            </p>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Отмена
            </Button>
            <Button
              type="submit"
              disabled={updateFilters.isPending}
            >
              {updateFilters.isPending ? "Сохранение..." : "Сохранить"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
