/**
 * Настройки уведомлений
 */

'use client';

import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { useNotificationsStore } from '@/store/notifications-store';

export function NotificationSettings() {
  const settings = useNotificationsStore((state) => state.settings);
  const updateSettings = useNotificationsStore((state) => state.updateSettings);

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-medium">Настройки уведомлений</h3>
        <p className="text-sm text-muted-foreground">
          Управление отображением уведомлений
        </p>
      </div>

      <div className="space-y-4">
        {/* Показывать Toast */}
        <div className="flex items-center justify-between space-x-4">
          <div className="flex-1 space-y-1">
            <Label htmlFor="show-toasts">
              Показывать всплывающие уведомления
            </Label>
            <p className="text-sm text-muted-foreground">
              Отображать Toast при получении новых уведомлений
            </p>
          </div>
          <Switch
            id="show-toasts"
            checked={settings.showToasts}
            onCheckedChange={(checked) =>
              updateSettings({ showToasts: checked })
            }
          />
        </div>

        {/* Только важные */}
        <div className="flex items-center justify-between space-x-4">
          <div className="flex-1 space-y-1">
            <Label htmlFor="significant-only">
              Только важные изменения
            </Label>
            <p className="text-sm text-muted-foreground">
              Показывать только уведомления о важных изменениях
            </p>
          </div>
          <Switch
            id="significant-only"
            checked={settings.showOnlySignificant}
            onCheckedChange={(checked) =>
              updateSettings({ showOnlySignificant: checked })
            }
          />
        </div>
      </div>
    </div>
  );
}
