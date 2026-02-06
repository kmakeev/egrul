/**
 * Zustand Store для управления уведомлениями
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { NotificationEvent } from '@/lib/notifications/sse-client';

export interface Notification extends NotificationEvent {
  isRead: boolean;
  readAt?: Date;
  createdAt: Date;
}

interface NotificationSettings {
  showToasts: boolean;
  showOnlySignificant: boolean;
}

interface NotificationsState {
  // Уведомления (последние 100)
  notifications: Notification[];
  unreadCount: number;
  isConnected: boolean;

  // Настройки
  settings: NotificationSettings;

  // Actions
  addNotification: (event: NotificationEvent) => void;
  markAsRead: (notificationId: string) => void;
  markAllAsRead: () => void;
  clearAll: () => void;
  setConnected: (connected: boolean) => void;
  updateSettings: (settings: Partial<NotificationSettings>) => void;
  loadNotifications: (notifications: Notification[]) => void;
}

const MAX_NOTIFICATIONS = 100;

export const useNotificationsStore = create<NotificationsState>()(
  persist(
    (set, get) => ({
      // Initial state
      notifications: [],
      unreadCount: 0,
      isConnected: false,
      settings: {
        showToasts: true,
        showOnlySignificant: false,
      },

      // Actions
      addNotification: (event: NotificationEvent) => {
        const { notifications, settings } = get();

        // Фильтрация по важности, если включена настройка
        if (settings.showOnlySignificant && !event.is_significant) {
          return;
        }

        // Проверить дублирование по ID
        if (notifications.some((n) => n.id === event.id)) {
          console.warn('[NotificationsStore] Duplicate notification:', event.id);
          return;
        }

        const newNotification: Notification = {
          ...event,
          isRead: false,
          createdAt: new Date(event.timestamp),
        };

        set((state) => {
          // Добавить в начало списка
          const updatedNotifications = [newNotification, ...state.notifications];

          // Ограничить размер до MAX_NOTIFICATIONS
          const limitedNotifications = updatedNotifications.slice(0, MAX_NOTIFICATIONS);

          // Пересчитать unreadCount
          const unreadCount = limitedNotifications.filter((n) => !n.isRead).length;

          return {
            notifications: limitedNotifications,
            unreadCount,
          };
        });
      },

      markAsRead: (notificationId: string) => {
        set((state) => {
          const updatedNotifications = state.notifications.map((n) =>
            n.id === notificationId
              ? { ...n, isRead: true, readAt: new Date() }
              : n
          );

          const unreadCount = updatedNotifications.filter((n) => !n.isRead).length;

          return {
            notifications: updatedNotifications,
            unreadCount,
          };
        });
      },

      markAllAsRead: () => {
        set((state) => ({
          notifications: state.notifications.map((n) => ({
            ...n,
            isRead: true,
            readAt: n.readAt || new Date(),
          })),
          unreadCount: 0,
        }));
      },

      clearAll: () => {
        set({
          notifications: [],
          unreadCount: 0,
        });
      },

      setConnected: (connected: boolean) => {
        set({ isConnected: connected });
      },

      updateSettings: (newSettings: Partial<NotificationSettings>) => {
        set((state) => ({
          settings: {
            ...state.settings,
            ...newSettings,
          },
        }));
      },

      loadNotifications: (loadedNotifications: Notification[]) => {
        set({
          notifications: loadedNotifications.slice(0, MAX_NOTIFICATIONS),
          unreadCount: loadedNotifications.filter((n) => !n.isRead).length,
        });
      },
    }),
    {
      name: 'notifications-storage',
      // Сохраняем только настройки, не уведомления (они загружаются с сервера)
      partialize: (state) => ({ settings: state.settings }),
    }
  )
);
