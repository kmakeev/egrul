/**
 * Список всех уведомлений с фильтрацией
 */

'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { NotificationItem } from './notification-item';
import { useNotificationsStore } from '@/store/notifications-store';
import type { Notification } from '@/store/notifications-store';

export function NotificationList() {
  const router = useRouter();
  const notifications = useNotificationsStore((state) => state.notifications);
  const markAsRead = useNotificationsStore((state) => state.markAsRead);
  const clearAll = useNotificationsStore((state) => state.clearAll);

  const [activeTab, setActiveTab] = useState<'all' | 'unread'>('all');

  // Фильтрация по табам
  const filteredNotifications =
    activeTab === 'unread'
      ? notifications.filter((n) => !n.isRead)
      : notifications;

  const unreadCount = notifications.filter((n) => !n.isRead).length;

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
    markAsRead(notification.id);

    // Синхронизация с сервером
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

    // Перейти на страницу компании/ИП
    const url = getEntityUrl(notification);
    if (url !== '#') {
      router.push(url);
    }
  };

  return (
    <div className="space-y-4">
      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as 'all' | 'unread')}>
        <div className="flex items-center justify-between">
          <TabsList>
            <TabsTrigger value="all">
              Все ({notifications.length})
            </TabsTrigger>
            <TabsTrigger value="unread">
              Непрочитанные ({unreadCount})
            </TabsTrigger>
          </TabsList>

          {notifications.length > 0 && (
            <Button
              variant="outline"
              size="sm"
              onClick={clearAll}
            >
              Очистить все
            </Button>
          )}
        </div>

        {/* Все */}
        <TabsContent value="all" className="space-y-2">
          {filteredNotifications.length === 0 ? (
            <div className="py-12 text-center">
              <p className="text-muted-foreground">Нет уведомлений</p>
            </div>
          ) : (
            filteredNotifications.map((notification) => (
              <NotificationItem
                key={notification.id}
                notification={notification}
                onClick={() => handleNotificationClick(notification)}
              />
            ))
          )}
        </TabsContent>

        {/* Непрочитанные */}
        <TabsContent value="unread" className="space-y-2">
          {filteredNotifications.length === 0 ? (
            <div className="py-12 text-center">
              <p className="text-muted-foreground">
                Нет непрочитанных уведомлений
              </p>
            </div>
          ) : (
            filteredNotifications.map((notification) => (
              <NotificationItem
                key={notification.id}
                notification={notification}
                onClick={() => handleNotificationClick(notification)}
              />
            ))
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
