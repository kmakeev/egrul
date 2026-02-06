/**
 * Badge с количеством непрочитанных уведомлений
 */

'use client';

import { Bell } from 'lucide-react';
import { useNotificationsStore } from '@/store/notifications-store';
import { cn } from '@/lib/utils';

interface NotificationBadgeProps {
  className?: string;
  onClick?: () => void;
}

export function NotificationBadge({ className, onClick }: NotificationBadgeProps) {
  const unreadCount = useNotificationsStore((state) => state.unreadCount);
  const isConnected = useNotificationsStore((state) => state.isConnected);

  return (
    <button
      onClick={onClick}
      className={cn(
        'relative p-2 rounded-lg hover:bg-accent transition-colors',
        className
      )}
      aria-label={`Уведомления (${unreadCount} непрочитанных)`}
    >
      <Bell className={cn(
        'h-5 w-5',
        isConnected ? 'text-foreground' : 'text-muted-foreground'
      )} />

      {unreadCount > 0 && (
        <span className="absolute top-0 right-0 flex h-5 w-5 items-center justify-center rounded-full bg-red-500 text-[10px] font-bold text-white">
          {unreadCount > 99 ? '99+' : unreadCount}
        </span>
      )}

      {/* Индикатор подключения */}
      {isConnected && (
        <span className="absolute bottom-1 right-1 h-2 w-2 rounded-full bg-green-500 ring-2 ring-background" />
      )}
    </button>
  );
}
