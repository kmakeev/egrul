"use client";

import React, { useState, useMemo, useEffect } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Search } from "lucide-react";
import { okvedOptions } from "@/lib/okved";

interface OkvedSelectProps {
  value?: string;
  onChange: (value: string | undefined) => void;
}

export function OkvedSelect({ value, onChange }: OkvedSelectProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [visibleCount, setVisibleCount] = useState(100);

  const filteredOptions = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return okvedOptions;
    return okvedOptions.filter((opt) => {
      const code = opt.code.toLowerCase();
      const title = opt.title.toLowerCase();
      return code.includes(q) || title.includes(q);
    });
  }, [searchQuery]);

  useEffect(() => {
    // При смене строки поиска начинаем показ с первой «страницы»
    setVisibleCount(100);
  }, [searchQuery]);

  const visibleOptions = filteredOptions.slice(0, visibleCount);

  const handleScroll: React.UIEventHandler<HTMLDivElement> = (e) => {
    const target = e.currentTarget;
    const threshold = 40; // px до низа, когда подгружаем ещё

    if (
      target.scrollTop + target.clientHeight >=
        target.scrollHeight - threshold &&
      visibleCount < filteredOptions.length
    ) {
      setVisibleCount((prev: number) =>
        Math.min(prev + 100, filteredOptions.length)
      );
    }
  };

  return (
    <Select
      value={value ?? "all"}
      onValueChange={(val: string) =>
        onChange(val === "all" ? undefined : val)
      }
    >
      <SelectTrigger>
        <SelectValue placeholder="Любой ОКВЭД" />
      </SelectTrigger>
      <SelectContent>
        <div className="p-2 border-b">
          <div className="relative">
            <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Поиск по коду или названию ОКВЭД..."
              value={searchQuery}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                setSearchQuery(e.target.value)
              }
              className="pl-8 h-8"
              onClick={(e: React.MouseEvent<HTMLInputElement>) =>
                e.stopPropagation()
              }
              onMouseDown={(e: React.MouseEvent<HTMLInputElement>) =>
                e.stopPropagation()
              }
            />
          </div>
        </div>
        <div
          className="max-h-[320px] overflow-y-auto"
          onScroll={handleScroll}
        >
          <SelectItem value="all">Любой ОКВЭД</SelectItem>
          {visibleOptions.map((opt: (typeof okvedOptions)[number]) => (
            <SelectItem key={opt.code} value={opt.code} title={opt.title}>
              <div className="flex items-center gap-2">
                <span className="font-mono">{opt.code}</span>
                <span className="truncate max-w-[24rem]">
                  {opt.title}
                </span>
              </div>
            </SelectItem>
          ))}
          {filteredOptions.length > visibleOptions.length && (
            <div className="px-3 py-2 text-xs text-muted-foreground">
              Показано {visibleOptions.length} из {filteredOptions.length} кодов ОКВЭД
            </div>
          )}
        </div>
      </SelectContent>
    </Select>
  );
}

