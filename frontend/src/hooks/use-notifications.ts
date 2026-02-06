/**
 * React Hook для управления SSE подключением и уведомлениями
 */

import { useEffect, useRef } from 'react';
import { SSEClient, type NotificationEvent } from '@/lib/notifications/sse-client';
import { useNotificationsStore } from '@/store/notifications-store';
import { useToast } from '@/hooks/use-toast';

export function useNotifications() {
  const sseClientRef = useRef<SSEClient | null>(null);
  const { toast } = useToast();

  const addNotification = useNotificationsStore((state) => state.addNotification);
  const setConnected = useNotificationsStore((state) => state.setConnected);
  const settings = useNotificationsStore((state) => state.settings);

  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    const email = localStorage.getItem('auth_email');

    if (!token || !email) {
      return;
    }

    // Загрузить историю уведомлений
    const loadHistory = async () => {
      try {
        const response = await fetch('http://localhost:8080/notifications/history?limit=50', {
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          console.error('Failed to load notification history:', response.statusText);
          return;
        }

        const data = await response.json();

        // Преобразовать формат API в формат Store
        const notifications = (data as Array<{
          ID: string;
          EntityType: string;
          EntityID: string;
          EntityName: string;
          ChangeType: string;
          FieldName: string;
          OldValue?: string;
          NewValue?: string;
          DetectedAt: string;
          IsRead: boolean;
          ReadAt?: string;
          CreatedAt: string;
        }>).map((item) => ({
          id: item.ID,
          type: 'change_detected',
          entity_type: item.EntityType,
          entity_id: item.EntityID,
          entity_name: item.EntityName,
          change_type: item.ChangeType,
          field_name: item.FieldName,
          old_value: item.OldValue || '',
          new_value: item.NewValue || '',
          is_significant: false,
          timestamp: item.DetectedAt,
          region_code: '',
          isRead: item.IsRead,
          readAt: item.ReadAt ? new Date(item.ReadAt) : undefined,
          createdAt: new Date(item.CreatedAt),
        }));

        useNotificationsStore.getState().loadNotifications(notifications);
      } catch (error) {
        console.error('Error loading notification history:', error);
      }
    };

    // Загрузить историю перед подключением SSE
    loadHistory();

    // Создать SSE клиента
    const client = new SSEClient({
      token,
      onMessage: (event: NotificationEvent) => {
        // Добавить в store
        addNotification(event);

        // Показать toast, если включено
        if (settings.showToasts) {
          showNotificationToast(event);
        }
      },
      onConnected: () => {
        console.log('[useNotifications] SSE Connected');
        setConnected(true);
      },
      onDisconnected: () => {
        console.log('[useNotifications] SSE Disconnected');
        setConnected(false);
      },
      onError: (error) => {
        console.error('[useNotifications] SSE Error:', error);
        setConnected(false);
      },
    });

    // Подключиться
    client.connect();
    sseClientRef.current = client;

    // Cleanup при размонтировании
    return () => {
      console.log('[useNotifications] Disconnecting SSE');
      client.disconnect();
      sseClientRef.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Подключаемся только один раз при монтировании

  // Функция показа toast уведомления
  const showNotificationToast = (event: NotificationEvent) => {
    const isSignificant = event.is_significant;
    const variant = isSignificant ? 'default' : 'default';

    toast({
      title: `${event.entity_type === 'company' ? 'Компания' : 'ИП'}: ${event.entity_name}`,
      description: `${getChangeTypeLabel(event.change_type)}: ${event.field_name}`,
      variant,
      duration: isSignificant ? 8000 : 5000,
    });
  };

  // Helper функция для получения человекочитаемого названия типа изменения
  const getChangeTypeLabel = (changeType: string): string => {
    const labels: Record<string, string> = {
      status: 'Изменение статуса',
      director: 'Смена руководителя',
      founders: 'Изменение учредителей',
      address: 'Смена адреса',
      capital: 'Изменение капитала',
      activities: 'Изменение видов деятельности',
    };

    return labels[changeType] || 'Изменение';
  };
}
