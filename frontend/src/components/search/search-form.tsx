"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search, X } from "lucide-react";
import { SearchFilters } from "./search-filters";
import type { SearchFiltersInput } from "@/lib/validations";

interface SearchFormProps {
  filters: SearchFiltersInput;
  onFiltersChange: (filters: Partial<SearchFiltersInput>) => void;
  onReset: () => void;
  onSave?: () => void;
  onApply?: () => void;
  isLoading?: boolean;
}

export function SearchForm({
  filters,
  onFiltersChange,
  onReset,
  onSave,
  onApply,
  isLoading,
}: SearchFormProps) {
  const [localQuery, setLocalQuery] = useState(filters.q ?? "");

  // Синхронизируем локальное состояние с фильтрами из URL
  useEffect(() => {
    setLocalQuery(filters.q ?? "");
  }, [filters.q]);

  // Обработчик сброса - также сбрасываем локальное состояние
  const handleReset = () => {
    setLocalQuery("");
    onReset();
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setLocalQuery(value);
    // Обновляем URL только если значение изменилось
    if (value !== filters.q) {
      onFiltersChange({ q: value || undefined });
    }
  };

  const hasQuickSearch = localQuery.trim().length >= 2;
  const hasAdvancedFilters =
    !!filters.innOgrn ||
    !!filters.region ||
    !!filters.okved ||
    (!!filters.status && filters.status !== "all") ||
    !!filters.founderName ||
    !!filters.dateFrom ||
    !!filters.dateTo;

  const canSearch = hasQuickSearch || hasAdvancedFilters;

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Быстрый поиск</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                type="text"
                placeholder="Введите название, ИНН, ОГРН или ОГРНИП..."
                value={localQuery}
                onChange={handleInputChange}
                className="pl-10"
                onKeyDown={(e) => {
                  if (e.key === "Enter" && !isLoading) {
                    e.preventDefault();
                  }
                }}
              />
            </div>
            <Button 
              disabled={isLoading || !canSearch}
              onClick={(e) => {
                e.preventDefault();
                if (onApply && canSearch) {
                  onApply();
                }
              }}
            >
              {isLoading ? "Поиск..." : "Найти"}
            </Button>
          </div>
          <p className="text-xs text-muted-foreground mt-2">
            Введите минимум 2 символа или укажите дополнительные фильтры
          </p>
          <div className="mt-2">
            <Button
              variant="outline"
              onClick={handleReset}
              disabled={isLoading}
              size="sm"
            >
              <X className="h-4 w-4 mr-2" />
              Сбросить
            </Button>
          </div>
        </CardContent>
      </Card>

      <SearchFilters
        filters={filters}
        onFiltersChange={onFiltersChange}
        onSave={onSave}
      />
    </div>
  );
}

