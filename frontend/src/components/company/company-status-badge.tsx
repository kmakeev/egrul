"use client";

import { Badge } from "@/components/ui/badge";
import { unifiedStatusOptions } from "@/lib/statuses";
import type { LegalEntity } from "@/lib/api";

interface CompanyStatusBadgeProps {
  company: LegalEntity;
}

export function CompanyStatusBadge({ company }: CompanyStatusBadgeProps) {
  // Функция для получения короткого текста статуса
  const getShortStatusText = (code: string): string => {
    switch (code) {
      case "101":
        return "Ликвидируется";
      case "105":
      case "106":
      case "107":
        return "Исключается из реестра";
      case "111":
        return "Уменьшение капитала";
      case "112":
        return "Изменение адреса";
      case "113":
        return "Банкротство возбуждено";
      case "114":
        return "Наблюдение";
      case "115":
        return "Финансовое оздоровление";
      case "116":
        return "Внешнее управление";
      case "117":
        return "Конкурсное производство";
      case "121":
        return "Реорганизация (преобразование)";
      case "122":
        return "Реорганизация (слияние)";
      case "123":
        return "Реорганизация (разделение)";
      case "124":
        return "Реорганизация (присоединение)";
      case "129":
      case "139":
        return "Реорганизация (смешанная)";
      case "131":
        return "Реорганизация (выделение)";
      case "132":
        return "Реорганизация (присоединение к нему)";
      case "134":
        return "Реорганизация (присоединение + выделение)";
      case "136":
        return "Реорганизация (выделение + присоединение)";
      case "701":
      case "702":
        return "Регистрация недействительна";
      case "801":
        return "Запись ошибочна";
      default:
        // Для неизвестных кодов пытаемся найти в unifiedStatusOptions
        const statusOption = unifiedStatusOptions.find(opt => opt.code === code);
        if (statusOption) {
          // Берем первую часть до запятой или весь текст, если запятой нет
          const parts = statusOption.label.split(',');
          return parts[0].trim();
        }
        return "Действующая";
    }
  };

  // Определяем статус и его вариант отображения
  const getStatusInfo = () => {
    // Если есть дата прекращения деятельности, значит компания закрыта
    if (company.terminationDate) {
      return { text: "Прекратила деятельность", variant: "destructive" as const };
    }

    // Если есть код статуса, используем его
    if (company.statusCode) {
      const statusText = getShortStatusText(company.statusCode);
      return {
        text: statusText,
        variant: getVariantByCode(company.statusCode)
      };
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