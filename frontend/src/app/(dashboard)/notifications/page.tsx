/**
 * Страница уведомлений
 */

import { Metadata } from 'next';
import { Separator } from '@/components/ui/separator';
import { NotificationList } from '@/components/notifications/notification-list';
import { NotificationSettings } from '@/components/notifications/notification-settings';

export const metadata: Metadata = {
  title: 'Уведомления | ЕГРЮЛ/ЕГРИП',
  description: 'История уведомлений об изменениях в компаниях и ИП',
};

export default function NotificationsPage() {
  return (
    <div className="container max-w-5xl py-8 space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Уведомления</h1>
        <p className="text-muted-foreground mt-2">
          Отслеживайте изменения в компаниях и ИП, на которые вы подписаны
        </p>
      </div>

      <Separator />

      {/* Список уведомлений */}
      <NotificationList />

      <Separator />

      {/* Настройки */}
      <NotificationSettings />
    </div>
  );
}
