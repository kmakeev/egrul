"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { User, Users } from "lucide-react";
import { useCompanyFoundersQuery } from "@/lib/api/company-hooks";
import { formatCurrency, formatPercentage } from "@/lib/format-utils";
import { decodeHtmlEntities } from "@/lib/html-utils";
import type { LegalEntity } from "@/lib/api";

interface CompanyManagementProps {
  company: LegalEntity;
}

export function CompanyManagement({ company }: CompanyManagementProps) {
  const { data: foundersData, isLoading: foundersLoading } = useCompanyFoundersQuery(company.ogrn);
  
  const founders = foundersData?.company?.founders || company.founders || [];

  return (
    <div className="space-y-6">
      {/* Руководитель */}
      {(company.director || company.head) && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <User className="h-5 w-5" />
              Руководитель
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div>
                <p className="font-semibold text-lg">
                  {`${(company.director || company.head)?.lastName || ""} ${(company.director || company.head)?.firstName || ""} ${(company.director || company.head)?.middleName || ""}`.trim()}
                </p>
                {(company.director || company.head)?.position && (
                  <p className="text-sm text-gray-600">{decodeHtmlEntities((company.director || company.head)!.position!)}</p>
                )}
              </div>
              
              {(company.director || company.head)?.inn && (
                <div>
                  <p className="text-sm text-gray-500">ИНН</p>
                  <p className="font-mono">{(company.director || company.head)!.inn}</p>
                </div>
              )}
              
              {/* TODO: Добавить дату назначения из истории изменений */}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Учредители */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="h-5 w-5" />
            Учредители
            {foundersLoading && (
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {founders.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Тип</TableHead>
                  <TableHead>Наименование/ФИО</TableHead>
                  <TableHead>ИНН</TableHead>
                  <TableHead className="text-right">Доля, %</TableHead>
                  <TableHead className="text-right">Сумма</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {founders.map((founder) => (
                  <TableRow key={founder.id}>
                    <TableCell>
                      <Badge variant={founder.type === "legal" ? "default" : "secondary"}>
                        {founder.type === "legal" ? "ЮЛ" : "ФЛ"}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-medium">
                      {decodeHtmlEntities(founder.name)}
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {founder.inn || "—"}
                    </TableCell>
                    <TableCell className="text-right">
                      {founder.share ? formatPercentage(founder.share) : "—"}
                    </TableCell>
                    <TableCell className="text-right">
                      {founder.amount 
                        ? formatCurrency(founder.amount, founder.currency) 
                        : "—"
                      }
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : foundersLoading ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div key={i} className="animate-pulse flex space-x-4">
                  <div className="h-4 bg-gray-200 rounded w-16"></div>
                  <div className="h-4 bg-gray-200 rounded flex-1"></div>
                  <div className="h-4 bg-gray-200 rounded w-24"></div>
                  <div className="h-4 bg-gray-200 rounded w-16"></div>
                  <div className="h-4 bg-gray-200 rounded w-20"></div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <Users className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-500">Информация об учредителях отсутствует</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}