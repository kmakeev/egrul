"use client";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { CompanyInfo } from "./company-info";
import { CompanyManagement } from "./company-management";
import { CompanyActivities } from "./company-activities";
import { CompanyHistory } from "./company-history";
import { CompanyRelations } from "./company-relations";
import type { LegalEntity } from "@/lib/api";

interface CompanyTabsProps {
  company: LegalEntity;
}

export function CompanyTabs({ company }: CompanyTabsProps) {
  return (
    <Tabs defaultValue="info" className="w-full">
      <TabsList className="grid w-full grid-cols-5">
        <TabsTrigger value="info">Основная информация</TabsTrigger>
        <TabsTrigger value="management">Руководство и учредители</TabsTrigger>
        <TabsTrigger value="activities">Виды деятельности</TabsTrigger>
        <TabsTrigger value="history">История изменений</TabsTrigger>
        <TabsTrigger value="relations">Связанные компании</TabsTrigger>
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