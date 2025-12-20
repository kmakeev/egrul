"use client";

import React from "react";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

interface EntityTypeSelectProps {
  value: "all" | "company" | "entrepreneur";
  onChange: (value: "all" | "company" | "entrepreneur") => void;
  className?: string;
}

export function EntityTypeSelect({ value, onChange, className }: EntityTypeSelectProps) {
  const options = [
    { value: "all" as const, label: "Все" },
    { value: "company" as const, label: "ЮЛ" },
    { value: "entrepreneur" as const, label: "ИП" },
  ];

  return (
    <RadioGroup
      value={value}
      onValueChange={onChange}
      className={cn("flex flex-row gap-6", className)}
    >
      {options.map((option) => (
        <div key={option.value} className="flex items-center space-x-2">
          <RadioGroupItem value={option.value} id={option.value} />
          <Label
            htmlFor={option.value}
            className="text-sm font-normal cursor-pointer"
          >
            {option.label}
          </Label>
        </div>
      ))}
    </RadioGroup>
  );
}