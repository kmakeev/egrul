"use client";

import { Badge } from "@/components/ui/badge";
import { unifiedStatusOptions } from "@/lib/statuses";
import type { LegalEntity } from "@/lib/api";

interface CompanyStatusBadgeProps {
  company: LegalEntity;
}

export function CompanyStatusBadge({ company }: CompanyStatusBadgeProps) {
  // Определяем статус и его вариант отображения
  const getStatusInfo = () => {
    // Если есть дата прекращения деятельности, значит компания закрыта
    if (company.terminationDate) {
      return { text: "Прекратила деятельность", variant: "destructive" as const };
    }

    // Если есть код статуса, используем его
    if (company.statusCode) {
      const statusOption = unifiedStatusOptions.find(opt => opt.code === company.statusCode);
      if (statusOption) {
        // Для ЮЛ показываем только первую часть до запятой (статус ЮЛ)
        const statusText = statusOption.label.split(',')[0].trim();
        return {
          text: statusText,
          variant: getVariantByCode(company.statusCode)
        };
      }
    }

    // По умолчанию считаем действующей
    return { text: "Действующая", variant: "default" as const };
  };

  // Определяем вариант бейджа по коду статуса
  const getVariantByCode = (code: string): "default" | "secondary" | "destructive" | "outline" => {
    // Коды ликвидации
    if (code === "101") {
      return "destructive";
    }
    
    // Коды исключения из реестра (недействующие)
    if (code === "105" || code === "106" || code === "107") {
      return "destructive";
    }
    
    // Коды банкротства (113-117)
    if (code === "113" || code === "114" || code === "115" || code === "116" || code === "117") {
      return "destructive";
    }
    
    // Коды реорганизации (121-139)
    if (code.startsWith("12") || code.startsWith("13")) {
      return "secondary";
    }
    
    // Коды недействительности регистрации (701, 702, 801, 802)
    if (code.startsWith("70") || code.startsWith("80")) {
      return "destructive";
    }
    
    // Остальные коды (например, 111 - уменьшение капитала, 112 - изменение места нахождения)
    return "outline";
  };

  const statusInfo = getStatusInfo();

  return (
    <Badge variant={statusInfo.variant} className="text-xs">
      {statusInfo.text}
    </Badge>
  );
}