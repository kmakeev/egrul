/**
 * Dropdown с последними уведомлениями
 */

'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { CheckCheck } from 'lucide-react';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { NotificationBadge } from './notification-badge';
import { NotificationItem } from './notification-item';
import { useNotificationsStore } from '@/store/notifications-store';
import type { Notification } from '@/store/notifications-store';

export function NotificationDropdown() {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const notifications = useNotificationsStore((state) => state.notifications);
  const markAsRead = useNotificationsStore((state) => state.markAsRead);
  const markAllAsRead = useNotificationsStore((state) => state.markAllAsRead);

  // Последние 5 уведомлений
  const recentNotifications = notifications.slice(0, 5);

  /**
   * Формирует URL страницы компании или ИП
   */
  const getEntityUrl = (notification: Notification): string => {
    if (notification.entity_type === 'company') {
      return `/companies/${notification.entity_id}`;
    } else if (notification.entity_type === 'entrepreneur') {
      return `/entrepreneurs/${notification.entity_id}`;
    }
    return '#';
  };

  const handleNotificationClick = async (notification: Notification) => {
    // Отметить как прочитанное
    markAsRead(notification.id);

    // Вызвать API для синхронизации с сервером
    try {
      await fetch(`http://localhost:8080/notifications/${notification.id}/read`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });
    } catch (error) {
      console.error('Failed to mark notification as read:', error);
    }

    // Закрыть dropdown
    setIsOpen(false);

    // Перейти на страницу компании/ИП
    const url = getEntityUrl(notification);
    if (url !== '#') {
      router.push(url);
    }
  };

  const handleMarkAllAsRead = async () => {
    markAllAsRead();

    // Вызвать API для синхронизации с сервером
    try {
      await fetch('http://localhost:8080/notifications/read-all', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });
    } catch (error) {
      console.error('Failed to mark all as read:', error);
    }
  };

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <div>
          <NotificationBadge onClick={() => setIsOpen(!isOpen)} />
        </div>
      </PopoverTrigger>

      <PopoverContent
        className="w-96 p-0"
        align="end"
        sideOffset={8}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3">
          <h3 className="font-semibold">Уведомления</h3>
          {notifications.length > 0 && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleMarkAllAsRead}
              className="h-8 text-xs"
            >
              <CheckCheck className="mr-1 h-3 w-3" />
              Прочитать все
            </Button>
          )}
        </div>

        <Separator />

        {/* Content */}
        {notifications.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
            <p className="text-sm text-muted-foreground">
              Нет уведомлений
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              Подпишитесь на компании, чтобы получать уведомления об изменениях
            </p>
          </div>
        ) : (
          <>
            <ScrollArea className="h-[400px]">
              <div className="p-2 space-y-1">
                {recentNotifications.map((notification) => (
                  <NotificationItem
                    key={notification.id}
                    notification={notification}
                    onClick={() => handleNotificationClick(notification)}
                    compact
                  />
                ))}
              </div>
            </ScrollArea>

            {notifications.length > 5 && (
              <>
                <Separator />
                <div className="p-2">
                  <Link href="/notifications">
                    <Button
                      variant="ghost"
                      className="w-full text-sm"
                      onClick={() => setIsOpen(false)}
                    >
                      Показать все ({notifications.length})
                    </Button>
                  </Link>
                </div>
              </>
            )}
          </>
        )}
      </PopoverContent>
    </Popover>
  );
}
