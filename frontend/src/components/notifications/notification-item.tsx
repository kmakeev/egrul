/**
 * Компонент для отображения одного уведомления
 */

'use client';

import { formatDistanceToNow } from 'date-fns';
import { ru } from 'date-fns/locale';
import { Building2, User, Circle } from 'lucide-react';
import type { Notification } from '@/store/notifications-store';
import { cn } from '@/lib/utils';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

interface NotificationItemProps {
  notification: Notification;
  onClick?: () => void;
  compact?: boolean;
}

export function NotificationItem({ notification, onClick, compact = false }: NotificationItemProps) {
  const EntityIcon = notification.entity_type === 'company' ? Building2 : User;

  const changeTypeLabels: Record<string, string> = {
    status: 'Статус',
    director: 'Руководитель',
    founders: 'Учредители',
    address: 'Адрес',
    capital: 'Капитал',
    activities: 'Виды деятельности',
  };

  const changeTypeLabel = changeTypeLabels[notification.change_type] || notification.change_type;

  // Ограничение длины названия компании
  const maxLength = compact ? 40 : 60;
  const entityName = notification.entity_name;
  const truncatedName = entityName.length > maxLength
    ? `${entityName.substring(0, maxLength)}...`
    : entityName;
  const shouldShowTooltip = entityName.length > maxLength;

  return (
    <TooltipProvider>
      <div
        onClick={onClick}
        className={cn(
          'flex gap-3 p-3 rounded-lg border transition-colors cursor-pointer',
          !notification.isRead && 'bg-muted/50',
          'hover:bg-accent',
          compact && 'p-2'
        )}
      >
        {/* Иконка */}
        <div className={cn(
          'flex-shrink-0',
          compact ? 'mt-0.5' : 'mt-1'
        )}>
          <div className={cn(
            'flex items-center justify-center rounded-full',
            notification.is_significant ? 'bg-red-100 text-red-600' : 'bg-blue-100 text-blue-600',
            compact ? 'h-8 w-8' : 'h-10 w-10'
          )}>
            <EntityIcon className={cn(compact ? 'h-4 w-4' : 'h-5 w-5')} />
          </div>
        </div>

        {/* Контент */}
        <div className="flex-1 min-w-0 overflow-hidden">
          <div className="flex items-start justify-between gap-2">
            <div className="flex-1 min-w-0 overflow-hidden">
              {shouldShowTooltip ? (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <p className={cn(
                      'font-medium truncate overflow-hidden text-ellipsis',
                      compact ? 'text-sm' : 'text-base'
                    )}>
                      {truncatedName}
                    </p>
                  </TooltipTrigger>
                  <TooltipContent side="top" className="max-w-md">
                    <p className="text-xs break-words">{entityName}</p>
                  </TooltipContent>
                </Tooltip>
              ) : (
                <p className={cn(
                  'font-medium truncate overflow-hidden text-ellipsis',
                  compact ? 'text-sm' : 'text-base'
                )}>
                  {entityName}
                </p>
              )}

              <p className={cn(
                'text-muted-foreground truncate overflow-hidden text-ellipsis',
                compact ? 'text-xs' : 'text-sm'
              )}>
                {changeTypeLabel}: {notification.field_name}
              </p>
            </div>

            {/* Непрочитано индикатор */}
            {!notification.isRead && (
              <Circle className="h-2 w-2 fill-blue-500 text-blue-500 flex-shrink-0 mt-2" />
            )}
          </div>

          {!compact && (
            <div className="mt-1 text-xs text-muted-foreground flex items-center gap-2 flex-wrap">
              <span>
                {formatDistanceToNow(notification.createdAt, {
                  addSuffix: true,
                  locale: ru,
                })}
              </span>
              {notification.is_significant && (
                <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-red-100 text-red-700">
                  Важное
                </span>
              )}
            </div>
          )}
        </div>
      </div>
    </TooltipProvider>
  );
}
