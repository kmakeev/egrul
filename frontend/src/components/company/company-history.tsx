"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Clock, ChevronDown, ChevronRight, FileText, Building2 } from "lucide-react";
import { useCompanyHistoryQuery } from "@/lib/api/company-hooks";
import { formatDate } from "@/lib/format-utils";
import { decodeHtmlEntities } from "@/lib/html-utils";
import type { LegalEntity } from "@/lib/api";

interface CompanyHistoryProps {
  company: LegalEntity;
}

export function CompanyHistory({ company }: CompanyHistoryProps) {
  const [expandedRecords, setExpandedRecords] = useState<Set<string>>(new Set());
  
  const { data: historyData, isLoading: historyLoading } = useCompanyHistoryQuery(company.ogrn);

  const historyRecords = historyData?.company?.history || [];

  const toggleRecordExpansion = (recordId: string) => {
    const newExpanded = new Set(expandedRecords);
    if (newExpanded.has(recordId)) {
      newExpanded.delete(recordId);
    } else {
      newExpanded.add(recordId);
    }
    setExpandedRecords(newExpanded);
  };

  const getReasonBadgeColor = (reasonCode?: string | null) => {
    if (!reasonCode) return "bg-gray-100 text-gray-800 border-gray-200";
    
    // Коды причин: 11xxx - создание, 12xxx - изменение, 13xxx - прекращение
    const code = reasonCode.substring(0, 2);
    switch (code) {
      case "11":
        return "bg-green-100 text-green-800 border-green-200";
      case "12":
        return "bg-blue-100 text-blue-800 border-blue-200";
      case "13":
        return "bg-red-100 text-red-800 border-red-200";
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
          
          {historyRecords.length > 0 && (
            <Badge variant="outline" className="text-sm">
              Всего записей: {historyRecords.length}
            </Badge>
          )}
        </div>
      </CardHeader>
      
      <CardContent>
        {historyLoading ? (
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="animate-pulse border rounded-lg p-4">
                <div className="flex items-center gap-3 mb-2">
                  <div className="h-6 w-32 bg-gray-200 rounded"></div>
                  <div className="h-4 w-24 bg-gray-200 rounded"></div>
                </div>
                <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                <div className="h-4 bg-gray-200 rounded w-1/2"></div>
              </div>
            ))}
          </div>
        ) : historyRecords.length > 0 ? (
          <div className="space-y-4">
            {historyRecords.map((record) => {
              const isExpanded = expandedRecords.has(record.id);
              const hasDetails = record.authority || record.certificateNumber || 
                                record.snapshotFullName || record.snapshotStatus || 
                                record.snapshotAddress;
              
              return (
                <div
                  key={record.id}
                  className="border rounded-lg p-4 transition-colors"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2 flex-wrap">
                        <Badge className={getReasonBadgeColor(record.reasonCode)}>
                          ГРН: {record.grn}
                        </Badge>
                        {record.date && (
                          <span className="text-sm text-gray-500 flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            {formatDate(record.date)}
                          </span>
                        )}
                      </div>
                      
                      {record.reasonDescription && (
                        <p className="text-sm font-medium mb-2">
                          {decodeHtmlEntities(record.reasonDescription)}
                        </p>
                      )}
                      
                      {record.reasonCode && (
                        <p className="text-xs text-gray-500 mb-2">
                          Код причины: {record.reasonCode}
                        </p>
                      )}
                      
                      {hasDetails && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => toggleRecordExpansion(record.id)}
                          className="flex items-center gap-1 p-0 h-auto text-blue-600 hover:text-blue-800 mt-2"
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
                  
                  {isExpanded && hasDetails && (
                    <div className="mt-3 pt-3 border-t border-gray-200 space-y-3">
                      {record.authority && (record.authority.code || record.authority.name) && (
                        <div className="flex gap-2">
                          <Building2 className="h-4 w-4 text-gray-400 mt-0.5 flex-shrink-0" />
                          <div className="flex-1">
                            <p className="text-xs text-gray-500 mb-1">Регистрирующий орган:</p>
                            <p className="text-sm font-medium">
                              {record.authority.name && decodeHtmlEntities(record.authority.name)}
                              {record.authority.code && (
                                <span className="text-xs text-gray-500 ml-2">
                                  (код: {record.authority.code})
                                </span>
                              )}
                            </p>
                          </div>
                        </div>
                      )}
                      
                      {(record.certificateSeries || record.certificateNumber || record.certificateDate) && (
                        <div className="flex gap-2">
                          <FileText className="h-4 w-4 text-gray-400 mt-0.5 flex-shrink-0" />
                          <div className="flex-1">
                            <p className="text-xs text-gray-500 mb-1">Свидетельство:</p>
                            <p className="text-sm font-medium">
                              {record.certificateSeries && `Серия: ${record.certificateSeries} `}
                              {record.certificateNumber && `№ ${record.certificateNumber}`}
                              {record.certificateDate && (
                                <span className="text-xs text-gray-500 ml-2">
                                  от {formatDate(record.certificateDate)}
                                </span>
                              )}
                            </p>
                          </div>
                        </div>
                      )}
                      
                      {record.snapshotFullName && (
                        <div className="bg-gray-50 rounded p-3">
                          <p className="text-xs text-gray-500 mb-1">Наименование на момент изменения:</p>
                          <p className="text-sm">{decodeHtmlEntities(record.snapshotFullName)}</p>
                        </div>
                      )}
                      
                      {record.snapshotStatus && (
                        <div className="bg-gray-50 rounded p-3">
                          <p className="text-xs text-gray-500 mb-1">Статус на момент изменения:</p>
                          <p className="text-sm">{record.snapshotStatus}</p>
                        </div>
                      )}
                      
                      {record.snapshotAddress && (
                        <div className="bg-gray-50 rounded p-3">
                          <p className="text-xs text-gray-500 mb-1">Адрес на момент изменения:</p>
                          <p className="text-sm">{decodeHtmlEntities(record.snapshotAddress)}</p>
                        </div>
                      )}
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
              История изменений отсутствует
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}