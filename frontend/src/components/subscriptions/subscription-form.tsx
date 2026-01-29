"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Bell, Check } from "lucide-react";
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
  EntityType,
  type ChangeFilters,
  type CreateSubscriptionInput,
  useCreateSubscriptionMutation,
} from "@/lib/api/subscription-hooks";

interface SubscriptionFormProps {
  entityType: "company" | "entrepreneur";
  entityId: string;
  entityName: string;
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

export function SubscriptionForm({
  entityType,
  entityId,
  entityName,
  open,
  onOpenChange,
}: SubscriptionFormProps) {
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<ChangeFilters>({
    status: true,
    director: true,
    founders: true,
    address: true,
    capital: true,
    activities: true,
  });

  // Преобразуем entityType в EntityType enum
  const entityTypeEnum = entityType === "company" ? EntityType.COMPANY : EntityType.ENTREPRENEUR;

  const createSubscription = useCreateSubscriptionMutation({
    onSuccess: () => {
      // Явно инвалидируем кэш для немедленного обновления UI
      queryClient.invalidateQueries({
        queryKey: ["subscription", "has", entityTypeEnum, entityId],
      });
      queryClient.invalidateQueries({
        queryKey: ["subscriptions", "my"],
      });
      toast({
        title: "Подписка создана",
        description: `Вы будете получать уведомления об изменениях в ${entityName}`,
      });
      onOpenChange(false);
      setFilters({
        status: true,
        director: true,
        founders: true,
        address: true,
        capital: true,
        activities: true,
      });
    },
    onError: (error) => {
      toast({
        title: "Ошибка при создании подписки",
        description: error.message,
        variant: "destructive",
      });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const input: CreateSubscriptionInput = {
      entityType:
        entityType === "company" ? EntityType.COMPANY : EntityType.ENTREPRENEUR,
      entityId,
      entityName,
      changeFilters: filters,
      notificationChannels: { email: true },
    };

    createSubscription.mutate(input);
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

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Bell className="h-5 w-5" />
            Подписка на изменения
          </DialogTitle>
          <DialogDescription>
            Получайте уведомления об изменениях в карточке организации
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Название организации */}
          <div className="rounded-lg border bg-muted/50 p-3">
            <p className="text-sm font-medium">{entityName}</p>
            <p className="text-xs text-muted-foreground mt-1">
              {entityType === "company" ? "ОГРН" : "ОГРНИП"}: {entityId}
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
              Уведомления будут приходить на email, указанный при регистрации
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
              disabled={createSubscription.isPending || selectedCount === 0}
            >
              {createSubscription.isPending ? "Создание..." : "Подписаться"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
