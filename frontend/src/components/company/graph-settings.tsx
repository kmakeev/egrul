"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Settings, Filter } from "lucide-react";

interface GraphSettingsProps {
  availableTypes: string[];
  selectedTypes: string[];
  onTypesChange: (types: string[]) => void;
}

const RELATIONSHIP_LABELS = {
  FOUNDER_COMPANY: "Учредители",
  SUBSIDIARY_COMPANY: "Дочерние",
  COMMON_FOUNDERS: "Общие учредители",
  COMMON_DIRECTORS: "Общие руководители",
  COMMON_ADDRESS: "Общий адрес",
  FOUNDER_TO_DIRECTOR: "Перекрестные связи",
  DIRECTOR_TO_FOUNDER: "Перекрестные связи",
  RELATED_BY_PERSON: "Связанные",
} as const;

const RELATIONSHIP_COLORS = {
  FOUNDER_COMPANY: "#3b82f6",
  SUBSIDIARY_COMPANY: "#10b981",
  COMMON_FOUNDERS: "#f59e0b",
  COMMON_DIRECTORS: "#8b5cf6",
  COMMON_ADDRESS: "#06b6d4",
  FOUNDER_TO_DIRECTOR: "#ef4444",
  DIRECTOR_TO_FOUNDER: "#ef4444",
  RELATED_BY_PERSON: "#6b7280",
} as const;

export function GraphSettings({ availableTypes, selectedTypes, onTypesChange }: GraphSettingsProps) {
  const [isOpen, setIsOpen] = useState(false);

  const handleTypeToggle = (type: string, checked: boolean) => {
    try {
      if (checked) {
        onTypesChange([...selectedTypes, type]);
      } else {
        // Не позволяем снять все фильтры
        const newTypes = selectedTypes.filter(t => t !== type);
        if (newTypes.length > 0) {
          onTypesChange(newTypes);
        }
      }
    } catch (error) {
      console.warn('Error toggling filter type:', error);
    }
  };

  const handleSelectAll = () => {
    try {
      onTypesChange(availableTypes);
    } catch (error) {
      console.warn('Error selecting all types:', error);
    }
  };

  const handleSelectNone = () => {
    try {
      // Оставляем хотя бы один тип выбранным
      if (availableTypes.length > 0) {
        onTypesChange([availableTypes[0]]);
      }
    } catch (error) {
      console.warn('Error clearing types:', error);
    }
  };

  const filteredCount = availableTypes.length - selectedTypes.length;

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm" className="flex items-center gap-2">
          <Settings className="h-4 w-4" />
          Фильтры
          {filteredCount > 0 && (
            <Badge variant="secondary" className="ml-1">
              -{filteredCount}
            </Badge>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80" align="end">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h4 className="font-medium">Типы связей</h4>
            <div className="flex gap-2">
              <Button
                variant="ghost"
                size="sm"
                onClick={handleSelectAll}
                className="h-6 px-2 text-xs"
              >
                Все
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={handleSelectNone}
                className="h-6 px-2 text-xs"
                title="Оставить только первый тип"
              >
                Сбросить
              </Button>
            </div>
          </div>
          
          <div className="space-y-3">
            {availableTypes.map((type) => {
              const isSelected = selectedTypes.includes(type);
              const label = RELATIONSHIP_LABELS[type as keyof typeof RELATIONSHIP_LABELS] || type;
              const color = RELATIONSHIP_COLORS[type as keyof typeof RELATIONSHIP_COLORS] || "#6b7280";
              
              return (
                <div key={type} className="flex items-center space-x-3">
                  <Checkbox
                    id={type}
                    checked={isSelected}
                    onCheckedChange={(checked) => handleTypeToggle(type, checked as boolean)}
                  />
                  <div className="flex items-center gap-2 flex-1">
                    <div 
                      className="w-3 h-3 rounded"
                      style={{ backgroundColor: color }}
                    />
                    <Label 
                      htmlFor={type}
                      className="text-sm font-normal cursor-pointer flex-1"
                    >
                      {label}
                    </Label>
                  </div>
                </div>
              );
            })}
          </div>
          
          {filteredCount > 0 && (
            <div className="pt-2 border-t">
              <div className="flex items-center gap-2 text-sm text-gray-600">
                <Filter className="h-4 w-4" />
                <span>Скрыто типов связей: {filteredCount}</span>
              </div>
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}