"use client";

import { useState } from "react";
import Link from "next/link";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  FileSpreadsheet,
  FileText,
  ExternalLink,
} from "lucide-react";
import type { SearchRow } from "@/hooks/use-search";
import type { SearchFiltersInput } from "@/lib/validations";
import { decodeHtmlEntities } from "@/lib/html-utils";

interface SearchResultsProps {
  rows: SearchRow[];
  total: number;
  isLoading: boolean;
  isFetching: boolean;
  error: Error | null;
  filters: SearchFiltersInput;
  onSortChange: (
    sortBy: "name" | "inn" | "ogrn" | "region" | "registrationDate",
    sortOrder: "asc" | "desc"
  ) => void;
  onExport?: (format: "csv" | "xlsx") => void;
}

type SortField = "name" | "inn" | "ogrn" | "registrationDate";

export function SearchResults({
  rows,
  total,
  isLoading,
  isFetching,
  error,
  filters,
  onSortChange,
  onExport,
}: SearchResultsProps) {
  const [selectedRows, setSelectedRows] = useState<Set<string>>(new Set());

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedRows(new Set(rows.map((row) => row.id)));
    } else {
      setSelectedRows(new Set());
    }
  };

  const handleSelectRow = (id: string, checked: boolean) => {
    const newSelected = new Set(selectedRows);
    if (checked) {
      newSelected.add(id);
    } else {
      newSelected.delete(id);
    }
    setSelectedRows(newSelected);
  };

  const getSortIcon = (field: SortField) => {
    if (filters.sortBy !== field) {
      return <ArrowUpDown className="h-4 w-4 ml-1" />;
    }
    return filters.sortOrder === "asc" ? (
      <ArrowUp className="h-4 w-4 ml-1" />
    ) : (
      <ArrowDown className="h-4 w-4 ml-1" />
    );
  };

  const handleSort = (field: SortField) => {
    const newOrder =
      filters.sortBy === field && filters.sortOrder === "asc" ? "desc" : "asc";
    onSortChange(field, newOrder);
  };

  const formatDate = (date: string | null | undefined) => {
    if (!date) return "-";
    try {
      return new Date(date).toLocaleDateString("ru-RU");
    } catch {
      return date;
    }
  };

  if (error) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="text-center py-8">
            <p className="text-destructive text-lg font-semibold mb-2">
              Ошибка загрузки результатов
            </p>
            <p className="text-muted-foreground">
              {error instanceof Error ? error.message : "Неизвестная ошибка"}
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Результаты поиска</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {[...Array(5)].map((_, i) => (
              <div
                key={i}
                className="h-16 bg-muted animate-pulse rounded-md"
              />
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (rows.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Результаты поиска</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-12">
            <p className="text-lg font-semibold mb-2">Ничего не найдено</p>
            <p className="text-muted-foreground">
              Попробуйте изменить параметры поиска
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>
            Результаты поиска
            {isFetching && (
              <span className="ml-2 text-sm font-normal text-muted-foreground">
                (обновление...)
              </span>
            )}
            <span className="ml-2 text-sm font-normal text-muted-foreground">
              ({total})
            </span>
          </CardTitle>
          {onExport && (
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => onExport("csv")}
                disabled={rows.length === 0}
              >
                <FileText className="h-4 w-4 mr-2" />
                CSV
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => onExport("xlsx")}
                disabled={rows.length === 0}
              >
                <FileSpreadsheet className="h-4 w-4 mr-2" />
                Excel
              </Button>
            </div>
          )}
        </div>
      </CardHeader>
      <CardContent>
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-12">
                  <Checkbox
                    checked={
                      rows.length > 0 && selectedRows.size === rows.length
                    }
                    onCheckedChange={handleSelectAll}
                  />
                </TableHead>
                <TableHead>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={() => handleSort("name")}
                  >
                    Наименование
                    {getSortIcon("name")}
                  </Button>
                </TableHead>
                <TableHead>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={() => handleSort("inn")}
                  >
                    ИНН
                    {getSortIcon("inn")}
                  </Button>
                </TableHead>
                <TableHead>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={() => handleSort("ogrn")}
                  >
                    ОГРН/ОГРНИП
                    {getSortIcon("ogrn")}
                  </Button>
                </TableHead>
                <TableHead>Регион</TableHead>
                <TableHead>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 -ml-3"
                    onClick={() => handleSort("registrationDate")}
                  >
                    Дата регистрации
                    {getSortIcon("registrationDate")}
                  </Button>
                </TableHead>
                <TableHead>Действия</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((row) => (
                <TableRow key={row.id}>
                  <TableCell>
                    <Checkbox
                      checked={selectedRows.has(row.id)}
                      onCheckedChange={(checked) =>
                        handleSelectRow(row.id, checked as boolean)
                      }
                    />
                  </TableCell>
                  <TableCell className="font-medium">{decodeHtmlEntities(row.name)}</TableCell>
                  <TableCell>{row.inn}</TableCell>
                  <TableCell>{row.ogrn ?? "-"}</TableCell>
                  <TableCell>{row.region ?? "-"}</TableCell>
                  <TableCell>{formatDate(row.registrationDate)}</TableCell>
                  <TableCell>
                    <Link
                      href={
                        row.type === "company"
                          ? `/company/${row.id}`
                          : `/entrepreneur/${row.id}`
                      }
                    >
                      <Button variant="ghost" size="sm">
                        <ExternalLink className="h-4 w-4" />
                      </Button>
                    </Link>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
        {selectedRows.size > 0 && (
          <div className="mt-4 p-3 bg-muted rounded-md flex items-center justify-between">
            <span className="text-sm">
              Выбрано: {selectedRows.size} из {rows.length}
            </span>
            <Button variant="outline" size="sm">
              Массовые операции
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

