"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Clock, Filter, ChevronDown, ChevronRight } from "lucide-react";
import { useCompanyHistoryQuery } from "@/lib/api/company-hooks";
import { formatDate } from "@/lib/format-utils";
import { decodeHtmlEntities } from "@/lib/html-utils";
import type { LegalEntity } from "@/lib/api";

interface CompanyHistoryProps {
  company: LegalEntity;
}

export function CompanyHistory({ company }: CompanyHistoryProps) {
  const [selectedFilter, setSelectedFilter] = useState<string>("all");
  const [expandedRecords, setExpandedRecords] = useState<Set<string>>(new Set());
  
  const { data: historyData, isLoading: historyLoading } = useCompanyHistoryQuery(company.ogrn);

  const historyRecords = historyData?.company?.history || company.history || [];

  // Фильтрация записей по типу
  const filteredRecords = historyRecords.filter(record => 
    selectedFilter === "all" || record.type === selectedFilter
  );

  // Получение уникальных типов изменений
  const changeTypes = Array.from(new Set(historyRecords.map(record => record.type)));

  const toggleRecordExpansion = (recordId: string) => {
    const newExpanded = new Set(expandedRecords);
    if (newExpanded.has(recordId)) {
      newExpanded.delete(recordId);
    } else {
      newExpanded.add(recordId);
    }
    setExpandedRecords(newExpanded);
  };

  const getChangeTypeColor = (type: string) => {
    switch (type.toLowerCase()) {
      case "регистрация":
        return "bg-green-100 text-green-800 border-green-200";
      case "изменение":
        return "bg-blue-100 text-blue-800 border-blue-200";
      case "ликвидация":
        return "bg-red-100 text-red-800 border-red-200";
      case "приостановление":
        return "bg-yellow-100 text-yellow-800 border-yellow-200";
      default:
        return "bg-gray-100 text-gray-800 border-gray-200";
    }
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Clock className="h-5 w-5" />
            История изменений
            {historyLoading && (
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
            )}
          </CardTitle>
          
          {changeTypes.length > 0 && (
            <div className="flex items-center gap-2">
              <Filter className="h-4 w-4 text-gray-500" />
              <Select value={selectedFilter} onValueChange={setSelectedFilter}>
                <SelectTrigger className="w-48">
                  <SelectValue placeholder="Фильтр по типу" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Все изменения</SelectItem>
                  {changeTypes.map((type) => (
                    <SelectItem key={type} value={type}>
                      {type}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}
        </div>
      </CardHeader>
      
      <CardContent>
        {historyLoading ? (
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="animate-pulse border rounded-lg p-4">
                <div className="flex items-center gap-3 mb-2">
                  <div className="h-6 w-20 bg-gray-200 rounded"></div>
                  <div className="h-4 w-24 bg-gray-200 rounded"></div>
                </div>
                <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                <div className="h-4 bg-gray-200 rounded w-1/2"></div>
              </div>
            ))}
          </div>
        ) : filteredRecords.length > 0 ? (
          <div className="space-y-4">
            {filteredRecords.map((record) => {
              const isExpanded = expandedRecords.has(record.id);
              
              return (
                <div
                  key={record.id}
                  className="border rounded-lg p-4 hover:bg-gray-50 transition-colors"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <Badge className={getChangeTypeColor(record.type)}>
                          {record.type}
                        </Badge>
                        <span className="text-sm text-gray-500">
                          {formatDate(record.date)}
                        </span>
                      </div>
                      
                      <p className="text-sm font-medium mb-2">
                        {decodeHtmlEntities(record.description)}
                      </p>
                      
                      {record.details && Object.keys(record.details).length > 0 && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => toggleRecordExpansion(record.id)}
                          className="flex items-center gap-1 p-0 h-auto text-blue-600 hover:text-blue-800"
                        >
                          {isExpanded ? (
                            <ChevronDown className="h-4 w-4" />
                          ) : (
                            <ChevronRight className="h-4 w-4" />
                          )}
                          {isExpanded ? "Скрыть детали" : "Показать детали"}
                        </Button>
                      )}
                    </div>
                  </div>
                  
                  {isExpanded && record.details && (
                    <div className="mt-3 pt-3 border-t border-gray-200">
                      <div className="grid gap-2">
                        {Object.entries(record.details).map(([key, value]) => (
                          <div key={key} className="flex gap-2">
                            <span className="text-sm text-gray-500 min-w-0 flex-shrink-0">
                              {key}:
                            </span>
                            <span className="text-sm font-medium">
                              {typeof value === "object" 
                                ? JSON.stringify(value, null, 2)
                                : decodeHtmlEntities(String(value))
                              }
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        ) : (
          <div className="text-center py-8">
            <Clock className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-500">
              {selectedFilter === "all" 
                ? "История изменений отсутствует"
                : `Изменения типа "${selectedFilter}" не найдены`
              }
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}