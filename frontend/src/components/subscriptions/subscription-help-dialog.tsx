"use client";

import { HelpCircle, Bell } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

export function SubscriptionHelpDialog() {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="ghost" size="icon" className="h-9 w-9">
          <HelpCircle className="h-5 w-5" />
          <span className="sr-only">Помощь</span>
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Как подписаться на изменения?</DialogTitle>
          <DialogDescription>
            Пошаговая инструкция по созданию подписки
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-2 text-sm">
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
        </div>
      </DialogContent>
    </Dialog>
  );
}
