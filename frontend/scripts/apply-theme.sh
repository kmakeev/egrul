#!/bin/bash

# Скрипт для применения цветовой темы
# Использование: ./scripts/apply-theme.sh <theme-name>
# Пример: ./scripts/apply-theme.sh blue

THEME_NAME=${1:-dark}
THEMES_FILE="src/styles/themes.ts"
GLOBALS_CSS="src/app/globals.css"

if [ ! -f "$THEMES_FILE" ]; then
  echo "❌ Файл $THEMES_FILE не найден!"
  exit 1
fi

if [ ! -f "$GLOBALS_CSS" ]; then
  echo "❌ Файл $GLOBALS_CSS не найден!"
  exit 1
fi

echo "🎨 Применение темы: $THEME_NAME"
echo ""
echo "⚠️  ВНИМАНИЕ: Этот скрипт требует ручной настройки."
echo "   Откройте файл src/styles/themes.ts и скопируйте значения"
echo "   из темы '$THEME_NAME' в src/app/globals.css"
echo ""
echo "📖 Подробные инструкции в файле THEMES.md"

