"use client";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { EntrepreneurInfo } from "./entrepreneur-info";
import { EntrepreneurActivities } from "./entrepreneur-activities";
import { EntrepreneurHistory } from "./entrepreneur-history";
import { Info, Building2, Clock } from "lucide-react";
import type { IndividualEntrepreneur } from "@/lib/api";

interface EntrepreneurTabsProps {
  entrepreneur: IndividualEntrepreneur;
}

export function EntrepreneurTabs({ entrepreneur }: EntrepreneurTabsProps) {
  return (
    <Tabs defaultValue="info" className="w-full">
      <TabsList className="grid w-full grid-cols-3 h-auto p-1">
        <TabsTrigger value="info" className="flex sm:flex-row items-center justify-center gap-1 sm:gap-2 p-1.5 sm:p-3 text-xs sm:text-sm min-h-[2.5rem]">
          <Info className="h-4 w-4 flex-shrink-0" />
          <span className="hidden lg:inline">Основная информация</span>
          <span className="hidden sm:inline lg:hidden">Основная</span>
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
      </TabsList>
      
      <TabsContent value="info" className="mt-6">
        <EntrepreneurInfo entrepreneur={entrepreneur} />
      </TabsContent>
      
      <TabsContent value="activities" className="mt-6">
        <EntrepreneurActivities entrepreneur={entrepreneur} />
      </TabsContent>
      
      <TabsContent value="history" className="mt-6">
        <EntrepreneurHistory entrepreneur={entrepreneur} />
      </TabsContent>
    </Tabs>
  );
}