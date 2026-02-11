"use client";

import { useState, useMemo, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { CalendarIcon, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

interface MonthRangePickerProps {
  dateFrom?: Date;
  dateTo?: Date;
  onChange: (dateFrom?: Date, dateTo?: Date) => void;
}

/**
 * Компонент выбора диапазона месяцев для фильтрации данных
 * Выбирается месяц и год, возвращаются даты начала и конца месяца
 */
export function MonthRangePicker({ dateFrom, dateTo, onChange }: MonthRangePickerProps) {
  const [open, setOpen] = useState(false);

  // Внутреннее состояние для выбора месяцев (инициализируется undefined чтобы избежать hydration mismatch)
  const [fromMonth, setFromMonth] = useState<number | undefined>(undefined);
  const [fromYear, setFromYear] = useState<number | undefined>(undefined);
  const [toMonth, setToMonth] = useState<number | undefined>(undefined);
  const [toYear, setToYear] = useState<number | undefined>(undefined);

  // Синхронизируем внутреннее состояние с props после монтирования (избегаем hydration error)
  useEffect(() => {
    setFromMonth(dateFrom ? dateFrom.getMonth() : undefined);
    setFromYear(dateFrom ? dateFrom.getFullYear() : undefined);
    setToMonth(dateTo ? dateTo.getMonth() : undefined);
    setToYear(dateTo ? dateTo.getFullYear() : undefined);
  }, [dateFrom, dateTo]);

  // Генерируем список годов (от 2000 до текущего года)
  const years = useMemo(() => {
    const currentYear = new Date().getFullYear();
    const yearsList = [];
    for (let year = 2000; year <= currentYear; year++) {
      yearsList.push(year);
    }
    return yearsList.reverse(); // Сначала последние годы
  }, []);

  // Названия месяцев
  const months = [
    "Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
    "Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"
  ];

  // Форматируем выбранный период для отображения
  const displayText = useMemo(() => {
    if (!dateFrom && !dateTo) return "Выберите период";

    const formatMonth = (date: Date) => {
      return date.toLocaleDateString("ru-RU", { month: "short", year: "numeric" });
    };

    if (dateFrom && dateTo) {
      return `${formatMonth(dateFrom)} — ${formatMonth(dateTo)}`;
    }
    if (dateFrom) {
      return `с ${formatMonth(dateFrom)}`;
    }
    if (dateTo) {
      return `до ${formatMonth(dateTo)}`;
    }
    return "Выберите период";
  }, [dateFrom, dateTo]);

  // Применить выбор
  const handleApply = () => {
    let newDateFrom: Date | undefined;
    let newDateTo: Date | undefined;

    // Создаем даты начала и конца выбранных месяцев
    if (fromYear !== undefined && fromMonth !== undefined) {
      newDateFrom = new Date(fromYear, fromMonth, 1); // Первый день месяца
    }

    if (toYear !== undefined && toMonth !== undefined) {
      newDateTo = new Date(toYear, toMonth + 1, 0, 23, 59, 59, 999); // Последний день месяца
    }

    onChange(newDateFrom, newDateTo);
    setOpen(false);
  };

  // Сбросить выбор
  const handleReset = () => {
    setFromMonth(undefined);
    setFromYear(undefined);
    setToMonth(undefined);
    setToYear(undefined);
    onChange(undefined, undefined);
    setOpen(false);
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className={cn(
            "w-[300px] justify-start text-left font-normal",
            !dateFrom && !dateTo && "text-muted-foreground"
          )}
        >
          <CalendarIcon className="mr-2 h-4 w-4" />
          {displayText}
          {(dateFrom || dateTo) && (
            <X
              className="ml-auto h-4 w-4 opacity-50 hover:opacity-100"
              onClick={(e) => {
                e.stopPropagation();
                handleReset();
              }}
            />
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-4" align="start">
        <div className="space-y-4">
          {/* Период "с" */}
          <div className="space-y-2">
            <label className="text-sm font-medium">С месяца</label>
            <div className="flex gap-2">
              <Select
                value={fromMonth?.toString()}
                onValueChange={(value) => setFromMonth(parseInt(value))}
              >
                <SelectTrigger className="w-[150px]">
                  <SelectValue placeholder="Месяц" />
                </SelectTrigger>
                <SelectContent>
                  {months.map((month, index) => (
                    <SelectItem key={index} value={index.toString()}>
                      {month}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Select
                value={fromYear?.toString()}
                onValueChange={(value) => setFromYear(parseInt(value))}
              >
                <SelectTrigger className="w-[100px]">
                  <SelectValue placeholder="Год" />
                </SelectTrigger>
                <SelectContent>
                  {years.map((year) => (
                    <SelectItem key={year} value={year.toString()}>
                      {year}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Период "по" */}
          <div className="space-y-2">
            <label className="text-sm font-medium">По месяц</label>
            <div className="flex gap-2">
              <Select
                value={toMonth?.toString()}
                onValueChange={(value) => setToMonth(parseInt(value))}
              >
                <SelectTrigger className="w-[150px]">
                  <SelectValue placeholder="Месяц" />
                </SelectTrigger>
                <SelectContent>
                  {months.map((month, index) => (
                    <SelectItem key={index} value={index.toString()}>
                      {month}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Select
                value={toYear?.toString()}
                onValueChange={(value) => setToYear(parseInt(value))}
              >
                <SelectTrigger className="w-[100px]">
                  <SelectValue placeholder="Год" />
                </SelectTrigger>
                <SelectContent>
                  {years.map((year) => (
                    <SelectItem key={year} value={year.toString()}>
                      {year}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Кнопки */}
          <div className="flex justify-between pt-2">
            <Button variant="ghost" size="sm" onClick={handleReset}>
              Сбросить
            </Button>
            <Button size="sm" onClick={handleApply}>
              Применить
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}
