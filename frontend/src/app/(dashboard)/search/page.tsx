"use client";

import { Suspense } from "react";
import { useSearch } from "@/hooks/use-search";
import { SearchForm } from "@/components/search/search-form";
import { SearchResults } from "@/components/search/search-results";
import { SearchPagination } from "@/components/search/search-pagination";
import { useToast } from "@/components/ui/use-toast";

export default function SearchPage() {
  return (
    <Suspense
      fallback={
        <div className="container mx-auto p-6">
          <div className="space-y-2">
            <h1 className="text-3xl font-bold tracking-tight">
              Поиск по ЕГРЮЛ/ЕГРИП
            </h1>
            <p className="text-muted-foreground">
              Гибкий поиск по юридическим лицам и индивидуальным предпринимателям
            </p>
          </div>
          <div className="mt-6 space-y-3">
            <div className="h-24 bg-muted animate-pulse rounded-md" />
            <div className="h-64 bg-muted animate-pulse rounded-md" />
          </div>
        </div>
      }
    >
      <SearchPageContent />
    </Suspense>
  );
}

function SearchPageContent() {
  const {
    filters,
      enabled,
    isLoading,
    isFetching,
    error,
    total,
    rows,
    page,
    pageSize,
    updateFilters,
    resetFilters,
    applyFilters,
  } = useSearch();

  const { toast } = useToast();

  const handleExport = (format: "csv" | "xlsx") => {
    // TODO: Реализовать экспорт данных
    // В реальном проекте здесь будет вызов API для экспорта
    toast({
      title: "Экспорт данных",
      description: `Экспорт в формате ${format.toUpperCase()} будет реализован в следующей версии`,
    });
  };

  const handleSaveFilter = () => {
    // TODO: Реализовать сохранение фильтров
    toast({
      title: "Сохранение фильтра",
      description:
        "Функция сохранения фильтров будет реализована в следующей версии",
    });
  };

  return (
    <div className="container mx-auto p-6 space-y-6">
                    <div className="space-y-2">
        <h1 className="text-3xl font-bold tracking-tight">Поиск по ЕГРЮЛ/ЕГРИП</h1>
        <p className="text-muted-foreground">
          Гибкий поиск по юридическим лицам и индивидуальным предпринимателям
        </p>
                    </div>

      <SearchForm
        filters={filters}
        onFiltersChange={updateFilters}
        onReset={resetFilters}
        onSave={handleSaveFilter}
        onApply={applyFilters}
        isLoading={isLoading || isFetching}
      />

      {enabled && (
        <>
          <SearchResults
            rows={rows}
            total={total}
            isLoading={isLoading}
            isFetching={isFetching}
            error={error}
            filters={filters}
            onSortChange={(sortBy, sortOrder) =>
              updateFilters({ sortBy, sortOrder })
            }
            onExport={handleExport}
          />

          {total > 0 && (
            <SearchPagination
              page={page}
              pageSize={pageSize}
              total={total}
              onPageChange={(newPage) => updateFilters({ page: newPage })}
              onPageSizeChange={(newPageSize) =>
                updateFilters({ pageSize: newPageSize, page: 1 })
              }
            />
          )}
        </>
                )}

      {!enabled && !isLoading && (
        <div className="text-center py-12 text-muted-foreground">
          <p>Задайте параметры поиска и нажмите &quot;Найти&quot; для получения результатов</p>
        </div>
      )}
    </div>
  );
}
