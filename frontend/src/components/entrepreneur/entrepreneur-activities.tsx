"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ChevronDown, ChevronUp, Building2 } from "lucide-react";
import { useEntrepreneurActivitiesQuery } from "@/lib/api/entrepreneur-hooks";
import { decodeHtmlEntities } from "@/lib/html-utils";
import type { IndividualEntrepreneur } from "@/lib/api";

interface EntrepreneurActivitiesProps {
  entrepreneur: IndividualEntrepreneur;
}

export function EntrepreneurActivities({ entrepreneur }: EntrepreneurActivitiesProps) {
  const [showAllActivities, setShowAllActivities] = useState(false);
  const { data: activitiesData, isLoading: activitiesLoading } = useEntrepreneurActivitiesQuery(entrepreneur.ogrnip);

  const allActivities = activitiesData?.entrepreneur?.activities || entrepreneur.activities || [];
  const additionalActivities = allActivities.filter(
    (activity) => activity.code !== entrepreneur.mainActivity?.code
  );

  const visibleActivities = showAllActivities 
    ? additionalActivities 
    : additionalActivities.slice(0, 5);

  return (
    <div className="space-y-6">
      {/* Основной вид деятельности */}
      {entrepreneur.mainActivity && (
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
                  {entrepreneur.mainActivity.code}
                </p>
                <p className="text-sm text-muted-foreground">
                  {decodeHtmlEntities(entrepreneur.mainActivity.name)}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Дополнительные виды деятельности */}
      {additionalActivities.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Building2 className="h-5 w-5" />
                Дополнительные виды деятельности
              </div>
              <Badge variant="outline">
                {additionalActivities.length}
              </Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            {activitiesLoading ? (
              <div className="flex items-center justify-center py-8">
                <p className="text-sm text-muted-foreground">Загрузка видов деятельности...</p>
              </div>
            ) : (
              <div className="space-y-3">
                {visibleActivities.map((activity, index) => (
                  <div key={`${activity.code}-${index}`} className="flex items-start gap-4 p-3 border rounded-lg">
                    <Badge variant="secondary">
                      Дополнительный
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

                {/* Кнопка "Показать все" */}
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
                          Показать все ({additionalActivities.length})
                        </>
                      )}
                    </Button>
                  </div>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Если нет дополнительных видов деятельности */}
      {additionalActivities.length === 0 && !activitiesLoading && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Building2 className="h-5 w-5" />
              Дополнительные виды деятельности
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground text-center py-8">
              Дополнительные виды деятельности не зарегистрированы
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}