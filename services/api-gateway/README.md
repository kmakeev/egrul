# EGRUL API Gateway

GraphQL API Gateway для доступа к данным ЕГРЮЛ/ЕГРИП.

## Возможности

- **GraphQL API** с полной схемой для запросов данных ЕГРЮЛ/ЕГРИП
- **REST API** для совместимости с существующими клиентами
- **Подключение к ClickHouse** для быстрых аналитических запросов
- **GraphQL Playground** для интерактивного тестирования запросов
- **Фильтрация и пагинация** данных
- **Текстовый поиск** по компаниям и ИП
- **Статистика** по регионам и видам деятельности

## Технологии

- Go 1.22+
- [gqlgen](https://gqlgen.com/) для GraphQL
- [chi](https://go-chi.io/) для HTTP маршрутизации
- [clickhouse-go](https://github.com/ClickHouse/clickhouse-go) для работы с ClickHouse
- [zap](https://github.com/uber-go/zap) для логирования
- [viper](https://github.com/spf13/viper) для конфигурации

## Быстрый старт

### Локальная разработка

```bash
# Установка зависимостей
go mod download

# Генерация GraphQL кода (опционально, для полной поддержки)
go run github.com/99designs/gqlgen generate

# Запуск
go run ./cmd/server
```

### Docker

```bash
# Из корня проекта
docker-compose up api-gateway
```

## Конфигурация

Конфигурация через переменные окружения или файл `config.yaml`:

| Переменная | Описание | По умолчанию |
|-----------|----------|--------------|
| `PORT` | Порт сервера | `8080` |
| `CLICKHOUSE_HOST` | Хост ClickHouse | `localhost` |
| `CLICKHOUSE_PORT` | Порт ClickHouse (native) | `9000` |
| `CLICKHOUSE_DATABASE` | База данных | `egrul` |
| `CLICKHOUSE_USER` | Пользователь | `default` |
| `CLICKHOUSE_PASSWORD` | Пароль | `` |
| `LOG_LEVEL` | Уровень логирования | `info` |
| `LOG_FORMAT` | Формат логов (json/text) | `json` |

## API Endpoints

### GraphQL

- `POST /graphql` - GraphQL endpoint
- `GET /playground` - GraphQL Playground

### REST (для совместимости)

- `GET /api/v1/companies/{ogrn}` - Получить компанию по ОГРН
- `GET /api/v1/entrepreneurs/{ogrnip}` - Получить ИП по ОГРНИП
- `GET /api/v1/search?q={query}` - Поиск

### Health Check

- `GET /health` - Проверка работоспособности
- `GET /ready` - Проверка готовности (включая ClickHouse)

## GraphQL Schema

### Основные запросы

```graphql
type Query {
  # Получение компании по ОГРН
  company(ogrn: ID!): Company
  
  # Получение компании по ИНН
  companyByInn(inn: String!): Company
  
  # Список компаний с фильтрацией
  companies(filter: CompanyFilter, pagination: Pagination, sort: CompanySort): CompanyConnection!
  
  # Поиск компаний
  searchCompanies(query: String!, limit: Int, offset: Int): [Company!]!
  
  # ИП
  entrepreneur(ogrnip: ID!): Entrepreneur
  entrepreneurs(filter: EntrepreneurFilter, pagination: Pagination): EntrepreneurConnection!
  
  # Универсальный поиск
  search(query: String!, limit: Int): SearchResult!
  
  # Статистика
  statistics(filter: StatsFilter): Statistics!
}
```

### Примеры запросов

#### Получение компании

```graphql
query GetCompany {
  company(ogrn: "1027700132195") {
    ogrn
    inn
    fullName
    status
    address {
      region
      city
      fullAddress
    }
    capital {
      amount
      currency
    }
    director {
      lastName
      firstName
      position
    }
    mainActivity {
      code
      name
    }
  }
}
```

#### Поиск компаний

```graphql
query SearchCompanies {
  searchCompanies(query: "Газпром", limit: 10) {
    ogrn
    inn
    fullName
    status
    registrationDate
  }
}
```

#### Список с фильтрацией

```graphql
query ListCompanies {
  companies(
    filter: {
      regionCode: "77"
      status: ACTIVE
      capitalMin: 1000000
    }
    pagination: { limit: 20, offset: 0 }
    sort: { field: CAPITAL_AMOUNT, order: DESC }
  ) {
    edges {
      node {
        ogrn
        fullName
        capital {
          amount
        }
      }
    }
    pageInfo {
      hasNextPage
      totalCount
    }
  }
}
```

#### Статистика

```graphql
query GetStatistics {
  statistics {
    totalCompanies
    activeCompanies
    liquidatedCompanies
    totalEntrepreneurs
    registeredThisYear
    byRegion {
      regionCode
      regionName
      companiesCount
    }
  }
}
```

## Разработка

### Структура проекта

```
services/api-gateway/
├── cmd/server/          # Точка входа
├── internal/
│   ├── config/          # Конфигурация
│   ├── graph/           # GraphQL схема и резолверы
│   │   ├── model/       # Модели данных
│   │   └── generated/   # Сгенерированный код
│   ├── middleware/      # HTTP middleware
│   ├── repository/      # Репозитории
│   │   └── clickhouse/  # ClickHouse репозитории
│   └── service/         # Бизнес-логика
├── config.yaml          # Конфигурация по умолчанию
├── gqlgen.yml           # Конфигурация gqlgen
├── Dockerfile
├── Makefile
└── README.md
```

### Команды Makefile

```bash
make build          # Сборка
make run            # Запуск
make test           # Тесты
make generate       # Генерация GraphQL кода
make lint           # Линтинг
make docker-build   # Сборка Docker образа
```

## Лицензия

MIT

