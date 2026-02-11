"use client";

import * as React from "react";
import { format, subMonths, startOfMonth, endOfMonth, startOfYear } from "date-fns";
import { ru } from "date-fns/locale";
import { CalendarIcon } from "lucide-react";
import { DateRange } from "react-day-picker";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";

interface DateRangePickerProps {
  dateFrom?: Date;
  dateTo?: Date;
  onChange: (dateFrom?: Date, dateTo?: Date) => void;
  className?: string;
}

// Быстрые пресеты для выбора периода
const PRESETS = [
  {
    label: "Последний месяц",
    getValue: () => ({
      from: startOfMonth(subMonths(new Date(), 1)),
      to: endOfMonth(subMonths(new Date(), 1)),
    }),
  },
  {
    label: "Последние 3 месяца",
    getValue: () => ({
      from: startOfMonth(subMonths(new Date(), 3)),
      to: new Date(),
    }),
  },
  {
    label: "Последние 6 месяцев",
    getValue: () => ({
      from: startOfMonth(subMonths(new Date(), 6)),
      to: new Date(),
    }),
  },
  {
    label: "Этот год",
    getValue: () => ({
      from: startOfYear(new Date()),
      to: new Date(),
    }),
  },
  {
    label: "Все время",
    getValue: () => ({
      from: undefined,
      to: undefined,
    }),
  },
];

/**
 * Компонент выбора диапазона дат для фильтров с быстрыми пресетами
 */
export function DateRangePicker({
  dateFrom,
  dateTo,
  onChange,
  className,
}: DateRangePickerProps) {
  const [date, setDate] = React.useState<DateRange | undefined>({
    from: dateFrom,
    to: dateTo,
  });

  const handleSelect = (range: DateRange | undefined) => {
    setDate(range);
    onChange(range?.from, range?.to);
  };

  const handlePresetClick = (preset: typeof PRESETS[0]) => {
    const range = preset.getValue();
    setDate(range);
    onChange(range.from, range.to);
  };

  return (
    <div className={cn("grid gap-2", className)}>
      <Popover>
        <PopoverTrigger asChild>
          <Button
            id="date"
            variant="outline"
            className={cn(
              "w-[280px] justify-start text-left font-normal",
              !date?.from && "text-muted-foreground"
            )}
          >
            <CalendarIcon className="mr-2 h-4 w-4" />
            {date?.from ? (
              date.to ? (
                <>
                  {format(date.from, "dd MMM yyyy", { locale: ru })} -{" "}
                  {format(date.to, "dd MMM yyyy", { locale: ru })}
                </>
              ) : (
                format(date.from, "dd MMM yyyy", { locale: ru })
              )
            ) : (
              <span>Выберите период</span>
            )}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <div className="flex">
            {/* Пресеты слева */}
            <div className="flex flex-col gap-1 p-3 border-r">
              <div className="text-xs font-semibold text-muted-foreground mb-1">
                Быстрый выбор
              </div>
              {PRESETS.map((preset) => (
                <Button
                  key={preset.label}
                  variant="ghost"
                  size="sm"
                  className="justify-start text-xs"
                  onClick={() => handlePresetClick(preset)}
                >
                  {preset.label}
                </Button>
              ))}
            </div>
            {/* Календарь справа */}
            <Calendar
              initialFocus
              mode="range"
              defaultMonth={date?.from}
              selected={date}
              onSelect={handleSelect}
              numberOfMonths={2}
              locale={ru}
            />
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}
