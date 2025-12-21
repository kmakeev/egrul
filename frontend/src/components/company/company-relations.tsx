"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Network, ExternalLink, Building } from "lucide-react";
import Link from "next/link";
import { useCompanyRelationsQuery } from "@/lib/api/company-hooks";
import { decodeHtmlEntities } from "@/lib/html-utils";
import type { LegalEntity } from "@/lib/api";

interface CompanyRelationsProps {
  company: LegalEntity;
}

export function CompanyRelations({ company }: CompanyRelationsProps) {
  const { data: relationsData, isLoading: relationsLoading } = useCompanyRelationsQuery(company.ogrn);
  
  const relatedCompanies = relationsData?.company?.relatedCompanies || company.relatedCompanies || [];

  const getRelationshipColor = (type: string) => {
    switch (type.toLowerCase()) {
      case "дочерняя":
      case "subsidiary":
        return "bg-blue-100 text-blue-800 border-blue-200";
      case "материнская":
      case "parent":
        return "bg-purple-100 text-purple-800 border-purple-200";
      case "аффилированная":
      case "affiliated":
        return "bg-orange-100 text-orange-800 border-orange-200";
      case "связанная":
      case "related":
        return "bg-green-100 text-green-800 border-green-200";
      default:
        return "bg-gray-100 text-gray-800 border-gray-200";
    }
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case "действующая":
      case "active":
        return "bg-green-100 text-green-800 border-green-200";
      case "ликвидирована":
      case "liquidated":
        return "bg-red-100 text-red-800 border-red-200";
      default:
        return "bg-gray-100 text-gray-800 border-gray-200";
    }
  };

  return (
    <div className="space-y-6">
      {/* Граф связей */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Network className="h-5 w-5" />
            Граф связей
          </CardTitle>
        </CardHeader>
        <CardContent>
          {/* TODO: Реализовать визуализацию графа связей */}
          <div className="h-64 bg-gray-100 rounded-lg flex items-center justify-center">
            <div className="text-center">
              <Network className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-500 mb-2">Визуализация графа связей</p>
              <p className="text-sm text-gray-400">(в разработке)</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Список связанных компаний */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <span className="flex items-center gap-2">
              <Building className="h-5 w-5" />
              Связанные компании
              {relationsLoading && (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
              )}
            </span>
            {relatedCompanies.length > 0 && (
              <Badge variant="secondary">
                {relatedCompanies.length}
              </Badge>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {relationsLoading ? (
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <div key={i} className="animate-pulse border rounded-lg p-4">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="h-6 w-20 bg-gray-200 rounded"></div>
                    <div className="h-6 w-16 bg-gray-200 rounded"></div>
                  </div>
                  <div className="h-6 bg-gray-200 rounded w-3/4 mb-1"></div>
                  <div className="h-4 bg-gray-200 rounded w-1/2"></div>
                </div>
              ))}
            </div>
          ) : relatedCompanies.length > 0 ? (
            <div className="space-y-4">
              {relatedCompanies.map((relatedCompany) => (
                <div
                  key={relatedCompany.id}
                  className="border rounded-lg p-4 hover:bg-gray-50 transition-colors"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <Badge className={getRelationshipColor(relatedCompany.relationshipType)}>
                          {relatedCompany.relationshipType}
                        </Badge>
                        <Badge className={getStatusColor(relatedCompany.status)}>
                          {relatedCompany.status}
                        </Badge>
                      </div>
                      
                      <h4 className="font-semibold text-lg mb-1">
                        {decodeHtmlEntities(relatedCompany.name)}
                      </h4>
                      
                      <p className="text-sm text-gray-600 font-mono">
                        ОГРН: {relatedCompany.ogrn}
                      </p>
                    </div>
                    
                    <Link href={`/company/${relatedCompany.ogrn}`}>
                      <Button variant="outline" size="sm" className="flex items-center gap-2">
                        <ExternalLink className="h-4 w-4" />
                        Открыть
                      </Button>
                    </Link>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <Building className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-500">Связанные компании не найдены</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}