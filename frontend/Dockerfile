# Build stage
FROM node:22-alpine AS builder

WORKDIR /app

# Установка pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# Копируем файлы зависимостей (lock-файл, если он появится, тоже попадёт в образ)
COPY package.json pnpm-lock.yaml* ./

# Устанавливаем зависимости без жёсткой привязки к lock-файлу
# (если pnpm-lock.yaml появится, он будет использован, но его отсутствие не приведёт к падению сборки)
RUN pnpm install --no-frozen-lockfile

# Копируем исходный код
COPY . .

# Сборка приложения
RUN pnpm build

# Runtime stage
FROM node:22-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production

# Создаем непривилегированного пользователя
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Копируем необходимые файлы
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]

