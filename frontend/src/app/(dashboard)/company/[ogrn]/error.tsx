"use client";

import { useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { AlertTriangle, RefreshCw } from "lucide-react";

export default function CompanyError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Логирование ошибки
    console.error("Company page error:", error);
  }, [error]);

  return (
    <div className="container mx-auto p-6">
      <Card>
        <CardContent className="pt-6">
          <div className="text-center">
            <div className="text-red-600 mb-4">
              <AlertTriangle className="h-12 w-12 mx-auto" />
            </div>
            
            <h2 className="text-xl font-semibold text-gray-900 mb-2">
              Ошибка загрузки компании
            </h2>
            
            <p className="text-gray-600 mb-6">
              {error.message || "Произошла неизвестная ошибка при загрузке данных компании"}
            </p>
            
            <div className="flex gap-4 justify-center">
              <Button onClick={reset} className="flex items-center gap-2">
                <RefreshCw className="h-4 w-4" />
                Попробовать снова
              </Button>
              
              <Button 
                variant="outline" 
                onClick={() => window.history.back()}
              >
                Вернуться назад
              </Button>
            </div>
            
            {process.env.NODE_ENV === "development" && (
              <details className="mt-6 text-left">
                <summary className="cursor-pointer text-sm text-gray-500 hover:text-gray-700">
                  Детали ошибки (только в режиме разработки)
                </summary>
                <pre className="mt-2 p-4 bg-gray-100 rounded text-xs overflow-auto">
                  {error.stack}
                </pre>
              </details>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}