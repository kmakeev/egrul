#!/usr/bin/env node

/**
 * Скрипт для проверки конфигурации карт
 * Использование: node scripts/check-maps-config.js
 */

const fs = require('fs');
const path = require('path');

const ENV_FILES = ['.env.local', '.env'];
const REQUIRED_VARS = ['NEXT_PUBLIC_YANDEX_MAPS_API_KEY'];

function checkEnvFile(filePath) {
  if (!fs.existsSync(filePath)) {
    return null;
  }
  
  const content = fs.readFileSync(filePath, 'utf8');
  const vars = {};
  
  content.split('\n').forEach(line => {
    const match = line.match(/^([^#][^=]+)=(.*)$/);
    if (match) {
      vars[match[1].trim()] = match[2].trim();
    }
  });
  
  return vars;
}

function main() {
  console.log('🗺️  Проверка конфигурации карт\n');
  
  let foundConfig = false;
  let hasValidKey = false;
  
  for (const envFile of ENV_FILES) {
    const filePath = path.join(process.cwd(), envFile);
    const vars = checkEnvFile(filePath);
    
    if (vars) {
      console.log(`📄 Найден файл: ${envFile}`);
      foundConfig = true;
      
      for (const varName of REQUIRED_VARS) {
        const value = vars[varName];
        
        if (!value) {
          console.log(`   ❌ ${varName}: не задана`);
        } else if (value === 'your_yandex_maps_api_key_here') {
          console.log(`   ⚠️  ${varName}: содержит пример значения`);
        } else if (value.length < 10) {
          console.log(`   ⚠️  ${varName}: слишком короткий ключ`);
        } else {
          console.log(`   ✅ ${varName}: настроена`);
          hasValidKey = true;
        }
      }
      console.log();
    }
  }
  
  if (!foundConfig) {
    console.log('❌ Файлы конфигурации не найдены');
    console.log('💡 Создайте .env.local из .env.example:');
    console.log('   cp .env.example .env.local\n');
  }
  
  if (!hasValidKey) {
    console.log('🔧 Для настройки карт:');
    console.log('1. Получите API ключ: https://developer.tech.yandex.ru/');
    console.log('2. Добавьте в .env.local:');
    console.log('   NEXT_PUBLIC_YANDEX_MAPS_API_KEY=ваш_ключ');
    console.log('3. Перезапустите сервер разработки\n');
  } else {
    console.log('🎉 Конфигурация карт настроена корректно!');
  }
}

main();