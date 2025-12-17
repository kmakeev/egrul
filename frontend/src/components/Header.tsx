"use client";

import Link from "next/link";

export function Header() {
  return (
    <header className="border-b border-slate-800/50 glass sticky top-0 z-50">
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center">
            <span className="text-white font-bold text-lg">Р</span>
          </div>
          <span className="text-xl font-semibold text-white">Реестры</span>
        </Link>

        <nav className="hidden md:flex items-center gap-8">
          <NavLink href="/">Поиск</NavLink>
          <NavLink href="/legal-entities">Юр. лица</NavLink>
          <NavLink href="/entrepreneurs">ИП</NavLink>
          <NavLink href="/analytics">Аналитика</NavLink>
        </nav>

        <div className="flex items-center gap-4">
          <button className="px-4 py-2 text-sm text-slate-400 hover:text-white transition-colors">
            API
          </button>
          <button className="px-4 py-2 text-sm bg-gradient-to-r from-indigo-500 to-purple-600 text-white rounded-lg hover:opacity-90 transition-opacity">
            Войти
          </button>
        </div>
      </div>
    </header>
  );
}

function NavLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <Link
      href={href}
      className="text-slate-400 hover:text-white transition-colors relative group"
    >
      {children}
      <span className="absolute -bottom-1 left-0 w-0 h-0.5 bg-gradient-to-r from-indigo-500 to-purple-600 group-hover:w-full transition-all duration-300" />
    </Link>
  );
}

