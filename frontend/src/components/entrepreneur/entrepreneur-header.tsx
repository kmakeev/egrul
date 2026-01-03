"use client";

import { Button } from "@/components/ui/button";
import { Card, CardHeader } from "@/components/ui/card";
import { Heart, Download, Share2 } from "lucide-react";
import { formatDate } from "@/lib/format-utils";
import { EntrepreneurStatusBadge } from "./entrepreneur-status-badge";
import type { IndividualEntrepreneur } from "@/lib/api";

interface EntrepreneurHeaderProps {
  entrepreneur: IndividualEntrepreneur;
}

export function EntrepreneurHeader({ entrepreneur }: EntrepreneurHeaderProps) {
  const handleAddToFavorites = () => {
    // TODO: Implement favorites functionality
    console.log("Add to favorites:", entrepreneur.ogrnip);
  };

  const handleDownloadExtract = () => {
    // TODO: Implement extract download
    console.log("Download extract:", entrepreneur.ogrnip);
  };

  const handleShare = () => {
    // TODO: Implement share functionality
    const fullName = `${entrepreneur.lastName} ${entrepreneur.firstName} ${entrepreneur.middleName || ""}`.trim();
    if (navigator.share) {
      navigator.share({
        title: fullName,
        text: `Информация об ИП ${fullName}`,
        url: window.location.href,
      });
    } else {
      navigator.clipboard.writeText(window.location.href);
    }
  };

  // Формируем полное имя ИП
  const fullName = `${entrepreneur.lastName} ${entrepreneur.firstName} ${entrepreneur.middleName || ""}`.trim();

  return (
    <Card>
      <CardHeader className="pb-4">
        <div className="flex flex-col lg:flex-row lg:items-start lg:justify-between gap-4">
          <div className="flex-1 min-w-0">
            <div className="mb-2">
              <EntrepreneurStatusBadge entrepreneur={entrepreneur} />
            </div>
            <div className="flex items-center gap-3 mb-2">
              <h1 className="text-2xl font-bold text-white break-words" style={{maxWidth: '100%', wordWrap: 'break-word'}}>
                {fullName}
              </h1>
            </div>
            <div className="space-y-1">
              <p className="text-sm text-gray-300">
                ОГРНИП: {entrepreneur.ogrnip}
              </p>
              <p className="text-sm text-gray-300">
                ИНН: {entrepreneur.inn}
              </p>
              {entrepreneur.registrationDate && (
                <p className="text-sm text-gray-300">
                  Дата регистрации: {formatDate(entrepreneur.registrationDate)}
                </p>
              )}
            </div>
          </div>
          
          <div className="flex flex-col sm:flex-row gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleAddToFavorites}
              className="flex items-center justify-center gap-2 w-full sm:w-auto"
            >
              <Heart className="h-4 w-4" />
              <span className="hidden sm:inline">В избранное</span>
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleDownloadExtract}
              className="flex items-center justify-center gap-2 w-full sm:w-auto"
            >
              <Download className="h-4 w-4" />
              <span className="hidden sm:inline">Выписка</span>
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleShare}
              className="flex items-center justify-center gap-2 w-full sm:w-auto"
            >
              <Share2 className="h-4 w-4" />
              <span className="hidden sm:inline">Поделиться</span>
            </Button>
          </div>
        </div>
      </CardHeader>
    </Card>
  );
}