"use client";

import { Badge } from "@/components/ui/badge";
import { unifiedStatusOptions } from "@/lib/statuses";
import type { IndividualEntrepreneur } from "@/lib/api";

interface EntrepreneurStatusBadgeProps {
  entrepreneur: IndividualEntrepreneur;
}

export function EntrepreneurStatusBadge({ entrepreneur }: EntrepreneurStatusBadgeProps) {
  // Функция для получения варианта бейджа на основе статуса
  const getBadgeVariant = () => {
    // Если есть дата прекращения деятельности, значит ИП закрыт
    if (entrepreneur.terminationDate) {
      return "destructive" as const;
    }

    // Если есть код статуса, используем его
    if (entrepreneur.statusCode) {
      const statusOption = unifiedStatusOptions.find(opt => opt.code === entrepreneur.statusCode);
      if (statusOption) {
        // Определяем вариант на основе кода статуса
        if (statusOption.code === "1") return "default" as const; // Действующий
        if (statusOption.code === "2") return "destructive" as const; // Прекратил деятельность
        if (statusOption.code === "3") return "secondary" as const; // Приостановил деятельность
        return "outline" as const;
      }
    }

    // По умолчанию считаем действующим
    return "default" as const;
  };

  // Функция для получения текста статуса
  const getStatusText = () => {
    // Если есть дата прекращения деятельности, значит ИП закрыт
    if (entrepreneur.terminationDate) {
      return "Прекратил деятельность";
    }

    // Если есть код статуса, используем его
    if (entrepreneur.statusCode) {
      const statusOption = unifiedStatusOptions.find(opt => opt.code === entrepreneur.statusCode);
      if (statusOption) {
        // Для ИП показываем вторую часть после запятой (статус ИП)
        const parts = statusOption.label.split(',');
        return parts.length > 1 ? parts[1].trim() : parts[0].trim();
      }
    }

    // По умолчанию считаем действующим
    return "Действующий";
  };

  return (
    <Badge variant={getBadgeVariant()} className="text-sm">
      {getStatusText()}
    </Badge>
  );
}