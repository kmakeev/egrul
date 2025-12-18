"use client";

import * as React from "react";
import { format } from "date-fns";
import { ru } from "date-fns/locale";
import { Calendar as CalendarIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Input } from "@/components/ui/input";

interface DatePickerProps {
  value?: string; // ISO date string (YYYY-MM-DD)
  onChange: (value: string | undefined) => void;
  placeholder?: string;
  maxDate?: Date;
  minDate?: Date;
  disabled?: boolean;
  className?: string;
}

export function DatePicker({
  value,
  onChange,
  placeholder = "Выберите дату",
  maxDate,
  minDate,
  disabled = false,
  className,
}: DatePickerProps) {
  const [open, setOpen] = React.useState(false);
  const [inputValue, setInputValue] = React.useState("");

  // Преобразуем строку ISO в Date для календаря
  const date = React.useMemo(() => {
    if (!value) return undefined;
    const d = new Date(value + "T00:00:00");
    return isNaN(d.getTime()) ? undefined : d;
  }, [value]);

  // Форматируем дату для отображения в поле ввода
  React.useEffect(() => {
    if (date) {
      setInputValue(format(date, "dd.MM.yyyy", { locale: ru }));
    } else {
      setInputValue("");
    }
  }, [date]);

  // Обработка ввода с клавиатуры
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const rawValue = e.target.value.trim();
    setInputValue(rawValue);

    if (!rawValue) {
      onChange(undefined);
      return;
    }

    // Поддерживаем форматы: DD.MM.YYYY, YYYY-MM-DD
    let parsedDate: Date | undefined;

    // Формат DD.MM.YYYY
    const ddmmyyyy = rawValue.match(/^(\d{2})\.(\d{2})\.(\d{4})$/);
    if (ddmmyyyy) {
      const [, dd, mm, yyyy] = ddmmyyyy;
      parsedDate = new Date(`${yyyy}-${mm}-${dd}T00:00:00`);
    }

    // Формат YYYY-MM-DD
    const yyyymmdd = rawValue.match(/^(\d{4})-(\d{2})-(\d{2})$/);
    if (yyyymmdd) {
      parsedDate = new Date(rawValue + "T00:00:00");
    }

    if (parsedDate && !isNaN(parsedDate.getTime())) {
      // Проверяем ограничения
      if (minDate && parsedDate < minDate) {
        parsedDate = minDate;
      }
      if (maxDate && parsedDate > maxDate) {
        parsedDate = maxDate;
      }

      const isoDate = parsedDate.toISOString().slice(0, 10);
      onChange(isoDate);
      setInputValue(format(parsedDate, "dd.MM.yyyy", { locale: ru }));
    }
  };

  // Обработка выбора из календаря
  const handleSelect = (selectedDate: Date | undefined) => {
    if (selectedDate) {
      const isoDate = selectedDate.toISOString().slice(0, 10);
      onChange(isoDate);
      setOpen(false);
    } else {
      onChange(undefined);
    }
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          disabled={disabled}
          className={cn(
            "w-full justify-start text-left font-normal",
            !date && "text-muted-foreground",
            className
          )}
        >
          <CalendarIcon className="mr-2 h-4 w-4" />
          {date ? (
            format(date, "dd.MM.yyyy", { locale: ru })
          ) : (
            <span>{placeholder}</span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <div className="p-3 space-y-2">
          <Input
            placeholder="ДД.ММ.ГГГГ"
            value={inputValue}
            onChange={handleInputChange}
            className="w-full"
            onFocus={() => setOpen(true)}
          />
          <Calendar
            mode="single"
            selected={date}
            onSelect={handleSelect}
            disabled={disabled}
            toDate={maxDate}
            fromDate={minDate}
            initialFocus
          />
        </div>
      </PopoverContent>
    </Popover>
  );
}

