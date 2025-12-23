"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ChevronDown, ChevronUp, Building2 } from "lucide-react";
import { useCompanyActivitiesQuery } from "@/lib/api/company-hooks";
import { decodeHtmlEntities } from "@/lib/html-utils";
import type { LegalEntity } from "@/lib/api";

interface CompanyActivitiesProps {
  company: LegalEntity;
}

export function CompanyActivities({ company }: CompanyActivitiesProps) {
  const [showAllActivities, setShowAllActivities] = useState(false);
  const { data: activitiesData, isLoading: activitiesLoading } = useCompanyActivitiesQuery(company.ogrn);

  const allActivities = activitiesData?.company?.activities || company.activities || [];
  const additionalActivities = allActivities.filter(
    (activity) => activity.code !== company.mainActivity?.code
  );

  const visibleActivities = showAllActivities 
    ? additionalActivities 
    : additionalActivities.slice(0, 5);

  return (
    <div className="space-y-6">
      {/* Основной вид деятельности */}
      {company.mainActivity && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Building2 className="h-5 w-5" />
              Основной вид деятельности
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-start gap-4 p-3 border rounded-lg">
              <Badge variant="default" className="bg-blue-600 dark:bg-blue-700">
                Основной
              </Badge>
              <div className="flex-1">
                <p className="font-mono text-sm font-semibold mb-1">
                  {company.mainActivity.code}
                </p>
                <p className="text-sm text-muted-foreground">
                  {decodeHtmlEntities(company.mainActivity.name)}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Дополнительные виды деятельности */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <span className="flex items-center gap-2">
              <Building2 className="h-5 w-5" />
              Дополнительные виды деятельности
              {activitiesLoading && (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
              )}
            </span>
            {additionalActivities.length > 0 && (
              <Badge variant="secondary">
                {additionalActivities.length}
              </Badge>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {activitiesLoading ? (
            <div className="space-y-3">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="animate-pulse flex items-start gap-4 p-3 border rounded-lg">
                  <div className="h-6 w-12 bg-gray-200 rounded"></div>
                  <div className="flex-1 space-y-2">
                    <div className="h-4 bg-gray-200 rounded w-24"></div>
                    <div className="h-4 bg-gray-200 rounded w-full"></div>
                  </div>
                </div>
              ))}
            </div>
          ) : additionalActivities.length > 0 ? (
            <div className="space-y-3">
              {visibleActivities.map((activity, index) => (
                <div
                  key={`${activity.code}-${index}`}
                  className="flex items-start gap-4 p-3 border rounded-lg transition-colors"
                >
                  <Badge variant="outline" className="text-xs">
                    Доп.
                  </Badge>
                  <div className="flex-1">
                    <p className="font-mono text-sm font-semibold mb-1">
                      {activity.code}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {decodeHtmlEntities(activity.name)}
                    </p>
                  </div>
                </div>
              ))}

              {additionalActivities.length > 5 && (
                <div className="flex justify-center pt-4">
                  <Button
                    variant="outline"
                    onClick={() => setShowAllActivities(!showAllActivities)}
                    className="flex items-center gap-2"
                  >
                    {showAllActivities ? (
                      <>
                        <ChevronUp className="h-4 w-4" />
                        Скрыть
                      </>
                    ) : (
                      <>
                        <ChevronDown className="h-4 w-4" />
                        Показать все ({additionalActivities.length - 5} еще)
                      </>
                    )}
                  </Button>
                </div>
              )}
            </div>
          ) : (
            <div className="text-center py-8">
              <Building2 className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-500">
                Дополнительные виды деятельности не указаны
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}