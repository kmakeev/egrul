# Инструкция по настройке проекта

## Быстрый старт

1. **Установка зависимостей:**
   ```bash
   npm install
   # или
   pnpm install
   ```

2. **Настройка переменных окружения:**
   ```bash
   cp .env.example .env.local
   ```
   Отредактируйте `.env.local` и укажите URL вашего API Gateway.

3. **Инициализация Husky (для Git hooks):**
   ```bash
   npm run prepare
   # или
   npx husky install
   ```

4. **Создание pre-commit hook:**
   ```bash
   npx husky add .husky/pre-commit "npx lint-staged"
   ```

5. **Запуск dev сервера:**
   ```bash
   npm run dev
   ```

   Приложение будет доступно по адресу: http://localhost:3000

## Проверка конфигурации

```bash
# Проверка типов TypeScript
npm run type-check

# Линтинг
npm run lint

# Форматирование
npm run format:check
```

## Структура маршрутов

- `/` - Главная страница
- `/login` - Страница входа
- `/register` - Страница регистрации
- `/search` - Поиск по ЕГРЮЛ/ЕГРИП
- `/company/[ogrn]` - Детальная информация о юридическом лице
- `/entrepreneur/[ogrnip]` - Детальная информация об ИП
- `/analytics` - Аналитика
- `/watchlist` - Избранное
- `/settings` - Настройки

## Компоненты

Все UI компоненты находятся в `src/components/ui/` и основаны на Shadcn/ui.

Для использования компонентов:

```tsx
import { Button, Card, Input } from "@/components/ui";
```

## API клиент

API клиент находится в `src/lib/api/client.ts`:

```tsx
import { apiClient } from "@/lib/api/client";

// Поиск
const results = await apiClient.globalSearch("ООО Рога и Копыта");

// Получение компании
const company = await apiClient.getLegalEntity("1234567890123");
```

## Состояние приложения

Используйте Zustand сторы:

```tsx
import { useAuthStore } from "@/store/auth-store";
import { useSearchStore } from "@/store/search-store";

// В компоненте
const { user, isAuthenticated, logout } = useAuthStore();
const { filters, setFilters } = useSearchStore();
```

## Валидация форм

Используйте Zod схемы с React Hook Form:

```tsx
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { loginSchema } from "@/lib/validations/auth";

const form = useForm({
  resolver: zodResolver(loginSchema),
});
```

## Troubleshooting

### Ошибки типов TypeScript

Убедитесь, что все зависимости установлены:
```bash
npm install
```

### Ошибки линтера

Запустите автоисправление:
```bash
npm run lint:fix
npm run format
```

### Проблемы с Husky

Если pre-commit hooks не работают:
```bash
npx husky install
npx husky add .husky/pre-commit "npx lint-staged"
chmod +x .husky/pre-commit
```

