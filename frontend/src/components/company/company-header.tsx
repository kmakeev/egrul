"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Heart, Download, Share2 } from "lucide-react";
import { decodeHtmlEntities } from "@/lib/html-utils";
import { formatDate } from "@/lib/format-utils";
import { CompanyStatusBadge } from "./company-status-badge";
import type { LegalEntity } from "@/lib/api";

interface CompanyHeaderProps {
  company: LegalEntity;
}

export function CompanyHeader({ company }: CompanyHeaderProps) {
  const handleAddToFavorites = () => {
    // TODO: Implement favorites functionality
    console.log("Add to favorites:", company.ogrn);
  };

  const handleDownloadExtract = () => {
    // TODO: Implement extract download
    console.log("Download extract:", company.ogrn);
  };

  const handleShare = () => {
    // TODO: Implement share functionality
    const decodedName = decodeHtmlEntities(company.fullName || company.shortName || "");
    if (navigator.share) {
      navigator.share({
        title: decodedName,
        text: `Информация о компании ${decodedName}`,
        url: window.location.href,
      });
    } else {
      navigator.clipboard.writeText(window.location.href);
    }
  };

  // Декодируем HTML-сущности в названиях
  const decodedFullName = company.fullName ? decodeHtmlEntities(company.fullName) : "";
  const decodedShortName = company.shortName ? decodeHtmlEntities(company.shortName) : null;

  // Определяем что показывать как основное название
  const primaryName = decodedFullName || decodedShortName || "Название не указано";
  const secondaryName = decodedFullName && decodedShortName && decodedFullName !== decodedShortName 
    ? decodedShortName 
    : null;

  return (
    <Card>
      <CardHeader className="pb-4">
        <div className="flex flex-col lg:flex-row lg:items-start lg:justify-between gap-4">
          <div className="flex-1 min-w-0">
            <div className="mb-2">
              <CompanyStatusBadge company={company} />
            </div>
            <div className="flex items-center gap-3 mb-2">
              <h1 className="text-2xl font-bold text-white break-words" style={{maxWidth: '100%', wordWrap: 'break-word'}}>
                {primaryName}
              </h1>
            </div>
            {secondaryName && (
              <p className="text-lg text-gray-200 mb-3 break-words">{secondaryName}</p>
            )}
            {company.registrationDate && (
              <p className="text-sm text-gray-300">
                Дата регистрации: {formatDate(company.registrationDate)}
              </p>
            )}
          </div>
          
          <div className="flex flex-col sm:flex-row gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleAddToFavorites}
              className="flex items-center justify-center gap-2 w-full sm:w-auto"
            >
              <Heart className="h-4 w-4" />
              В избранное
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleDownloadExtract}
              className="flex items-center justify-center gap-2 w-full sm:w-auto"
            >
              <Download className="h-4 w-4" />
              Скачать выписку
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleShare}
              className="flex items-center justify-center gap-2 w-full sm:w-auto"
            >
              <Share2 className="h-4 w-4" />
              Поделиться
            </Button>
          </div>
        </div>
      </CardHeader>
      
      <CardContent>
        <div className="grid grid-cols-3 gap-6">
          <div>
            <p className="text-sm text-gray-500 mb-1">ОГРН</p>
            <p className="font-mono text-lg font-semibold">{company.ogrn}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500 mb-1">ИНН</p>
            <p className="font-mono text-lg font-semibold">{company.inn}</p>
          </div>
          {company.kpp && (
            <div>
              <p className="text-sm text-gray-500 mb-1">КПП</p>
              <p className="font-mono text-lg font-semibold">{company.kpp}</p>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}