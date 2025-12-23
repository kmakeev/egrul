"use client";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { CompanyInfo } from "./company-info";
import { CompanyManagement } from "./company-management";
import { CompanyActivities } from "./company-activities";
import { CompanyHistory } from "./company-history";
import { CompanyRelations } from "./company-relations";
import { Info, Users, Building2, Clock, Network } from "lucide-react";
import type { LegalEntity } from "@/lib/api";

interface CompanyTabsProps {
  company: LegalEntity;
}

export function CompanyTabs({ company }: CompanyTabsProps) {
  return (
    <Tabs defaultValue="info" className="w-full">
      <TabsList className="grid w-full grid-cols-5 h-auto p-1">
        <TabsTrigger value="info" className="flex sm:flex-row items-center justify-center gap-1 sm:gap-2 p-1.5 sm:p-3 text-xs sm:text-sm min-h-[2.5rem]">
          <Info className="h-4 w-4 flex-shrink-0" />
          <span className="hidden lg:inline">Основная информация</span>
          <span className="hidden sm:inline lg:hidden">Основная</span>
        </TabsTrigger>
        <TabsTrigger value="management" className="flex sm:flex-row items-center justify-center gap-1 sm:gap-2 p-1.5 sm:p-3 text-xs sm:text-sm min-h-[2.5rem]">
          <Users className="h-4 w-4 flex-shrink-0" />
          <span className="hidden lg:inline">Руководство и учредители</span>
          <span className="hidden sm:inline lg:hidden">Руководство</span>
        </TabsTrigger>
        <TabsTrigger value="activities" className="flex sm:flex-row items-center justify-center gap-1 sm:gap-2 p-1.5 sm:p-3 text-xs sm:text-sm min-h-[2.5rem]">
          <Building2 className="h-4 w-4 flex-shrink-0" />
          <span className="hidden lg:inline">Виды деятельности</span>
          <span className="hidden sm:inline lg:hidden">Деятельность</span>
        </TabsTrigger>
        <TabsTrigger value="history" className="flex sm:flex-row items-center justify-center gap-1 sm:gap-2 p-1.5 sm:p-3 text-xs sm:text-sm min-h-[2.5rem]">
          <Clock className="h-4 w-4 flex-shrink-0" />
          <span className="hidden lg:inline">История изменений</span>
          <span className="hidden sm:inline lg:hidden">История</span>
        </TabsTrigger>
        <TabsTrigger value="relations" className="flex sm:flex-row items-center justify-center gap-1 sm:gap-2 p-1.5 sm:p-3 text-xs sm:text-sm min-h-[2.5rem]">
          <Network className="h-4 w-4 flex-shrink-0" />
          <span className="hidden lg:inline">Связанные компании</span>
          <span className="hidden sm:inline lg:hidden">Связи</span>
        </TabsTrigger>
      </TabsList>
      
      <TabsContent value="info" className="mt-6">
        <CompanyInfo company={company} />
      </TabsContent>
      
      <TabsContent value="management" className="mt-6">
        <CompanyManagement company={company} />
      </TabsContent>
      
      <TabsContent value="activities" className="mt-6">
        <CompanyActivities company={company} />
      </TabsContent>
      
      <TabsContent value="history" className="mt-6">
        <CompanyHistory company={company} />
      </TabsContent>
      
      <TabsContent value="relations" className="mt-6">
        <CompanyRelations company={company} />
      </TabsContent>
    </Tabs>
  );
}