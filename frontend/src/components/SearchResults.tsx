"use client";

interface SearchResultsProps {
  query: string;
}

export function SearchResults({ query }: SearchResultsProps) {
  // TODO: Подключить к API
  const isLoading = false;
  const results: SearchResult[] = [];

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500" />
      </div>
    );
  }

  if (results.length === 0) {
    return (
      <div className="glass rounded-2xl p-8 text-center">
        <div className="text-6xl mb-4">🔍</div>
        <h3 className="text-xl font-semibold text-white mb-2">
          Поиск по запросу: &quot;{query}&quot;
        </h3>
        <p className="text-slate-400">
          Результаты будут отображаться здесь после подключения к API
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-white">
          Результаты поиска{" "}
          <span className="text-slate-400 font-normal">
            ({results.length} найдено)
          </span>
        </h2>
        <div className="flex items-center gap-2">
          <SortButton>По релевантности</SortButton>
          <SortButton>По дате</SortButton>
        </div>
      </div>

      <div className="grid gap-4">
        {results.map((result) => (
          <ResultCard key={result.id} result={result} />
        ))}
      </div>
    </div>
  );
}

interface SearchResult {
  id: string;
  type: "legal_entity" | "entrepreneur";
  name: string;
  inn: string;
  ogrn?: string;
  ogrnip?: string;
  status: string;
  address?: string;
  registrationDate?: string;
}

function ResultCard({ result }: { result: SearchResult }) {
  const isLegalEntity = result.type === "legal_entity";

  return (
    <div className="glass rounded-xl p-6 hover:border-indigo-500/40 transition-all duration-300 cursor-pointer group">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-center gap-3 mb-2">
            <span
              className={`px-2 py-1 rounded-md text-xs font-medium ${
                isLegalEntity
                  ? "bg-blue-500/20 text-blue-300"
                  : "bg-emerald-500/20 text-emerald-300"
              }`}
            >
              {isLegalEntity ? "ЮЛ" : "ИП"}
            </span>
            <StatusBadge status={result.status} />
          </div>

          <h3 className="text-lg font-semibold text-white group-hover:text-indigo-300 transition-colors">
            {result.name}
          </h3>

          <div className="mt-3 flex flex-wrap gap-4 text-sm text-slate-400">
            <span>
              ИНН: <span className="text-slate-300">{result.inn}</span>
            </span>
            <span>
              {isLegalEntity ? "ОГРН" : "ОГРНИП"}:{" "}
              <span className="text-slate-300">
                {result.ogrn || result.ogrnip}
              </span>
            </span>
            {result.registrationDate && (
              <span>
                Дата регистрации:{" "}
                <span className="text-slate-300">{result.registrationDate}</span>
              </span>
            )}
          </div>

          {result.address && (
            <p className="mt-2 text-sm text-slate-500">{result.address}</p>
          )}
        </div>

        <button className="text-slate-400 hover:text-white transition-colors p-2">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 5l7 7-7 7"
            />
          </svg>
        </button>
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const statusConfig: Record<string, { label: string; className: string }> = {
    active: {
      label: "Действующее",
      className: "bg-green-500/20 text-green-300",
    },
    liquidated: {
      label: "Ликвидировано",
      className: "bg-red-500/20 text-red-300",
    },
    reorganized: {
      label: "Реорганизовано",
      className: "bg-yellow-500/20 text-yellow-300",
    },
    bankrupt: {
      label: "Банкрот",
      className: "bg-orange-500/20 text-orange-300",
    },
  };

  const config = statusConfig[status] || {
    label: "Неизвестно",
    className: "bg-slate-500/20 text-slate-300",
  };

  return (
    <span className={`px-2 py-1 rounded-md text-xs font-medium ${config.className}`}>
      {config.label}
    </span>
  );
}

function SortButton({ children }: { children: React.ReactNode }) {
  return (
    <button className="px-3 py-1.5 text-sm text-slate-400 hover:text-white hover:bg-slate-700/50 rounded-lg transition-colors">
      {children}
    </button>
  );
}

