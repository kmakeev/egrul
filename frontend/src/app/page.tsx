"use client";

import { useState } from "react";
import { SearchBar } from "@/components/SearchBar";
import { SearchResults } from "@/components/SearchResults";
import { Header } from "@/components/Header";

export default function HomePage() {
  const [searchQuery, setSearchQuery] = useState("");

  return (
    <div className="min-h-screen flex flex-col">
      <Header />

      <main className="flex-1 container mx-auto px-4 py-8">
        {/* Hero секция */}
        <section className="text-center py-16 animate-fade-in">
          <h1 className="text-5xl md:text-7xl font-bold mb-6">
            <span className="gradient-text">ЕГРЮЛ / ЕГРИП</span>
          </h1>
          <p className="text-xl md:text-2xl text-slate-400 mb-12 max-w-2xl mx-auto">
            Поиск организаций и индивидуальных предпринимателей в реестрах
            Федеральной налоговой службы
          </p>

          <SearchBar value={searchQuery} onChange={setSearchQuery} />
        </section>

        {/* Результаты поиска */}
        {searchQuery && (
          <section className="mt-12 animate-fade-in" style={{ animationDelay: "0.2s" }}>
            <SearchResults query={searchQuery} />
          </section>
        )}

        {/* Статистика */}
        {!searchQuery && (
          <section className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-16">
            <StatCard
              title="Юридических лиц"
              value="—"
              description="зарегистрировано в системе"
              delay="0.1s"
            />
            <StatCard
              title="Индивидуальных предпринимателей"
              value="—"
              description="в базе данных"
              delay="0.2s"
            />
            <StatCard
              title="Обновлений"
              value="—"
              description="последнее обновление"
              delay="0.3s"
            />
          </section>
        )}
      </main>

      {/* Футер */}
      <footer className="border-t border-slate-800/50 py-6">
        <div className="container mx-auto px-4 text-center text-slate-500 text-sm">
          <p>© {new Date().getFullYear()} ЕГРЮЛ/ЕГРИП Система. Данные предоставлены ФНС России.</p>
        </div>
      </footer>
    </div>
  );
}

function StatCard({
  title,
  value,
  description,
  delay,
}: {
  title: string;
  value: string;
  description: string;
  delay: string;
}) {
  return (
    <div
      className="glass rounded-2xl p-6 animate-fade-in hover:scale-105 transition-transform duration-300"
      style={{ animationDelay: delay }}
    >
      <h3 className="text-slate-400 text-sm font-medium mb-2">{title}</h3>
      <p className="text-4xl font-bold gradient-text mb-2">{value}</p>
      <p className="text-slate-500 text-sm">{description}</p>
    </div>
  );
}

