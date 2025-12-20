"use client";

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RegionSelect } from "./region-select";
import { OkvedSelect } from "./okved-select";
import { DatePicker } from "./date-picker";
import { EntityTypeSelect } from "./entity-type-select";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { unifiedStatusOptions } from "@/lib/statuses";
import { Button } from "@/components/ui/button";
import { Save, Search } from "lucide-react";
import { ChevronDown, ChevronUp } from "lucide-react";
import type { SearchFiltersInput } from "@/lib/validations";

// Временно отключаем логи фронтенда для просмотра логов бэкенда
const ENABLE_FRONTEND_LOGS = false;

interface SearchFiltersProps {
  filters: SearchFiltersInput;
  onFiltersChange: (filters: Partial<SearchFiltersInput>) => void;
  onSave?: () => void;
}

export function SearchFilters({
  filters,
  onFiltersChange,
  onSave,
}: SearchFiltersProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [resetKey] = useState(0);
  const [statusSearchQuery, setStatusSearchQuery] = useState("");

  const today = new Date();
  today.setHours(0, 0, 0, 0);

  const filteredStatusOptions = unifiedStatusOptions.filter((opt) => {
    if (!statusSearchQuery.trim()) return true;
    const q = statusSearchQuery.toLowerCase();
    return (
      opt.code.toLowerCase().includes(q) ||
      opt.label.toLowerCase().includes(q)
    );
  });

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Расширенный поиск</CardTitle>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setIsExpanded(!isExpanded)}
          >
            {isExpanded ? (
              <>
                <ChevronUp className="h-4 w-4 mr-1" />
                Свернуть
              </>
            ) : (
              <>
                <ChevronDown className="h-4 w-4 mr-1" />
                Развернуть
              </>
            )}
          </Button>
        </div>
      </CardHeader>
      {isExpanded && (
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="entityType">Тип организации</Label>
            <EntityTypeSelect
              value={filters.entityType ?? "all"}
              onChange={(value) => {
                // #region agent log: EntityTypeSelect onChange
                if (ENABLE_FRONTEND_LOGS) {
                  fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
                  method: "POST",
                  headers: { "Content-Type": "application/json" },
                  body: JSON.stringify({
                    sessionId: "debug-session",
                    runId: "run-filters",
                    hypothesisId: "H1",
                    location: "search-filters.tsx:entityType:onChange",
                    message: "Entity type filter changed in SearchFilters",
                    data: { 
                      currentEntityType: filters.entityType, 
                      newEntityType: value,
                      allFilters: filters,
                    },
                    timestamp: Date.now(),
                  }),
                }).catch(() => {});
                }
                // #endregion agent log: EntityTypeSelect onChange
                onFiltersChange({ entityType: value });
              }}
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="region">Регион</Label>
              <RegionSelect
                key={resetKey}
                value={filters.region}
                onChange={(value) => {
                  // #region agent log: SearchFilters region onChange
                  if (ENABLE_FRONTEND_LOGS) {
                    fetch("http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                      sessionId: "debug-session",
                      runId: "run-filters",
                      hypothesisId: "H1",
                      location: "search-filters.tsx:region:onChange",
                      message: "Region filter changed in SearchFilters",
                      data: { 
                        currentRegion: filters.region, 
                        newRegion: value,
                        allFilters: filters,
                      },
                      timestamp: Date.now(),
                    }),
                  }).catch(() => {});
                  }
                  // #endregion agent log: SearchFilters region onChange
                  onFiltersChange({ region: value });
                }}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="okved">ОКВЭД</Label>
              <OkvedSelect
                value={filters.okved}
                onChange={(value) => onFiltersChange({ okved: value })}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="status">Статус (ЮЛ и ИП, по коду)</Label>
              <Select
                value={filters.status ?? "all"}
                onValueChange={(value) => {
                  const nextValue = value === "all" ? undefined : value;

                  if (ENABLE_FRONTEND_LOGS) {
                    fetch(
                      "http://127.0.0.1:7242/ingest/d909b3ca-a27d-43bc-a00e-99361eba3af1",
                      {
                        method: "POST",
                        headers: { "Content-Type": "application/json" },
                        body: JSON.stringify({
                          sessionId: "debug-session",
                          runId: "run-filters",
                          hypothesisId: "H5",
                          location: "search-filters.tsx:status:select",
                          message: "Unified status filter changed",
                          data: {
                            currentStatus: filters.status,
                            newStatus: nextValue,
                          },
                          timestamp: Date.now(),
                        }),
                      }
                    ).catch(() => {});
                  }

                  onFiltersChange({ status: nextValue });
                }}
              >
                <SelectTrigger id="status">
                  <SelectValue placeholder="Все статусы" />
                </SelectTrigger>
                <SelectContent>
                  <div className="p-2 border-b">
                    <div className="relative">
                      <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                      <Input
                        placeholder="Поиск по коду или описанию..."
                        value={statusSearchQuery}
                        onChange={(e) => setStatusSearchQuery(e.target.value)}
                        className="pl-8 h-8"
                        onClick={(e) => e.stopPropagation()}
                        onMouseDown={(e) => e.stopPropagation()}
                      />
                    </div>
                  </div>
                  <div className="max-h-[300px] overflow-y-auto">
                    <SelectItem value="all">Все статусы</SelectItem>
                    {filteredStatusOptions.map((opt) => (
                      <SelectItem key={opt.code} value={opt.code}>
                        <div className="flex items-center gap-2">
                          <span className="font-mono">{opt.code}</span>
                          <span
                            className="truncate max-w-[30rem]"
                            title={opt.label}
                          >
                            {opt.label}
                          </span>
                        </div>
                      </SelectItem>
                    ))}
                  </div>
                </SelectContent>
              </Select>
            </div>

            {/* Поле ФИО учредителя показываем только для ЮЛ */}
            {(filters.entityType === "all" || filters.entityType === "company") && (
              <div className="space-y-2">
                <Label htmlFor="founderName">ФИО учредителя (ЮЛ)</Label>
                <Input
                  id="founderName"
                  type="text"
                  placeholder="Введите ФИО учредителя..."
                  value={filters.founderName ?? ""}
                  onChange={(e) =>
                    onFiltersChange({
                      founderName: e.target.value || undefined,
                    })
                  }
                />
              </div>
            )}

            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="dateFrom">Диапазон дат регистрации</Label>
              <div className="flex items-center gap-2">
                <DatePicker
                  value={filters.dateFrom}
                  onChange={(value) => {
                    let nextTo = filters.dateTo;

                    // Если from > to, двигаем верхнюю границу
                    if (value && nextTo && value > nextTo) {
                      nextTo = value;
                    }

                    onFiltersChange({ dateFrom: value, dateTo: nextTo });
                  }}
                  placeholder="От"
                  maxDate={today}
                />
                <span className="text-muted-foreground">—</span>
                <DatePicker
                  value={filters.dateTo}
                  onChange={(value) => {
                    let nextFrom = filters.dateFrom;

                    // Если to < from, двигаем нижнюю границу
                    if (nextFrom && value && nextFrom > value) {
                      nextFrom = value;
                    }

                    onFiltersChange({ dateFrom: nextFrom, dateTo: value });
                  }}
                  placeholder="До"
                  maxDate={today}
                  minDate={
                    filters.dateFrom
                      ? new Date(filters.dateFrom + "T00:00:00")
                      : undefined
                  }
                />
              </div>
            </div>
          </div>

          {onSave && (
            <div className="flex gap-2 pt-2">
              <Button onClick={onSave} variant="outline">
                <Save className="h-4 w-4 mr-2" />
                Сохранить фильтр
              </Button>
            </div>
          )}
        </CardContent>
      )}
    </Card>
  );
}

