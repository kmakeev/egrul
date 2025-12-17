"use client";

import { useState, useCallback } from "react";

interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
}

export function SearchBar({ value, onChange }: SearchBarProps) {
  const [isFocused, setIsFocused] = useState(false);

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
    },
    []
  );

  return (
    <form onSubmit={handleSubmit} className="w-full max-w-3xl mx-auto">
      <div
        className={`relative transition-all duration-300 ${
          isFocused ? "animate-pulse-glow" : ""
        }`}
      >
        <div className="absolute inset-0 bg-gradient-to-r from-indigo-500/20 to-purple-500/20 rounded-2xl blur-xl" />
        
        <div className="relative glass rounded-2xl p-2 flex items-center gap-2">
          {/* Иконка поиска */}
          <div className="pl-4 text-slate-400">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-6 w-6"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
          </div>

          {/* Поле ввода */}
          <input
            type="text"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            onFocus={() => setIsFocused(true)}
            onBlur={() => setIsFocused(false)}
            placeholder="Введите ИНН, ОГРН, название или ФИО..."
            className="flex-1 bg-transparent text-white placeholder-slate-500 text-lg py-4 px-2 focus:outline-none"
          />

          {/* Фильтры */}
          <div className="flex items-center gap-2 pr-2">
            <FilterButton active>Все</FilterButton>
            <FilterButton>ЮЛ</FilterButton>
            <FilterButton>ИП</FilterButton>
          </div>

          {/* Кнопка поиска */}
          <button
            type="submit"
            className="bg-gradient-to-r from-indigo-500 to-purple-600 text-white px-8 py-4 rounded-xl font-medium hover:opacity-90 transition-opacity"
          >
            Найти
          </button>
        </div>
      </div>

      {/* Подсказки */}
      <div className="mt-4 flex flex-wrap justify-center gap-2 text-sm text-slate-500">
        <span>Примеры:</span>
        <button
          type="button"
          onClick={() => onChange("7707083893")}
          className="text-indigo-400 hover:text-indigo-300 transition-colors"
        >
          ИНН: 7707083893
        </button>
        <span>•</span>
        <button
          type="button"
          onClick={() => onChange("1027700132195")}
          className="text-indigo-400 hover:text-indigo-300 transition-colors"
        >
          ОГРН: 1027700132195
        </button>
        <span>•</span>
        <button
          type="button"
          onClick={() => onChange("Газпром")}
          className="text-indigo-400 hover:text-indigo-300 transition-colors"
        >
          Название: Газпром
        </button>
      </div>
    </form>
  );
}

function FilterButton({
  children,
  active = false,
}: {
  children: React.ReactNode;
  active?: boolean;
}) {
  return (
    <button
      type="button"
      className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
        active
          ? "bg-indigo-500/20 text-indigo-300 border border-indigo-500/30"
          : "text-slate-400 hover:text-white hover:bg-slate-700/50"
      }`}
    >
      {children}
    </button>
  );
}

