---
name: tech-blueprint
description: Генерирует технические контракты (Prisma-схема, GraphQL API, архитектура, план тестирования) на основе SPEC.md и UI-прототипа. Также проверяет уже созданный blueprint по запросу. Запускай когда SPEC.md утверждён и нужно перейти к техническому проектированию, или когда пользователь говорит «проверь ТЗ», «валидируй blueprint». НЕ запускай если SPEC.md ещё черновик или задача — рефакторинг существующего кода.
license: MIT
compatibility: opencode
metadata:
  level: multi
  output: папка 3_TECH_BLUEPRINT/
---

## Назначение

Скилл создаёт пять документов, полностью описывающих техническую реализацию продукта:
- **IMPLEMENTATION_GUIDE.md** — онбординг разработчика: стек, что уже готово, как запустить
- **DATABASE_MODEL.md** — Prisma-схема, отношения, индексы
- **API_CONTRACTS.md** — GraphQL-схема с типами, операциями и правилами доступа
- **ARCHITECTURE.md** — NestJS-модули, FSD-слайсы, управление состоянием
- **TESTING_PLAN.md** — Unit-тесты и E2E-сценарии Playwright

**Когда запускать:**
- SPEC.md имеет статус «на ревью» или «утверждён»
- UI-прототип существует в папке `prototype/`
- Команда готова начинать разработку

**Когда НЕ запускать:**
- SPEC.md ещё черновик
- Нет понимания бизнес-сущностей
- Задача — рефакторинг или фикс бага в существующем проекте

---

## Входные данные

Перед генерацией **обязательно прочитать и проанализировать**:

1. **`1_PRODUCT_VISION/VISION.md`** — цель, границы, метрики успеха
2. **`2_PRODUCT_SPEC/SPEC.md`** — сущности, операции, страницы, тестирование (**главный источник истины**)
3. **`prototype/`** — только для понимания скопа: какие страницы существуют, какие потоки между ними, какие формы. **Не копировать из прототипа технические решения.** Прототип сделан быстро и без учёта best practices — он показывает «что», но не «как». Всё архитектурное — по лучшим практикам стека, независимо от того, как это реализовано в прототипе.
4. **Существующие схемы** — если в проекте уже есть `schema.prisma` или `schema.graphql`: прочитать перед генерацией, это означает режим обновления

Если входных данных недостаточно — задать **не более 2 уточняющих вопросов**.

Перед генерацией составить внутренний список:
- бизнес-сущности (из SPEC.md §Сущности)
- ключевые операции (из SPEC.md §Ключевые операции)
- страницы (из SPEC.md §Страницы, прототип — для сверки скопа)
- роли пользователей (из SPEC.md §Глоссарий или §Сущности)

---

## Стартовые шаблоны: что уже реализовано

Проекты строятся на основе двух предготовленных стартовых шаблонов (boilerplate — готовая кодовая база с решёнными базовыми задачами): шаблон для бэкенда (NestJS) и шаблон для фронтенда (React). В них уже реализованы авторизация, работа с пользователями и тестовая инфраструктура. **Не проектировать заново** — только расширять.

### Backend

**Prisma-модель `Profile` (уже существует, не дублировать):**
```prisma
model Profile {
  id        Int           @id @default(autoincrement())
  createdAt DateTime      @default(now())
  updatedAt DateTime      @updatedAt
  oidcSub   String        @unique   // связь с OIDC-провайдером
  roles     ProfileRole[] @default([USER])
  avatarUrl String?
}

enum ProfileRole { ADMIN  USER }
```

**GraphQL-операции (уже реализованы, не включать в Blueprint):**
| Операция | Тип | Доступ |
|---------|-----|--------|
| `me` | Query | JwtAuthGuard |
| `updateProfile(input)` | Mutation | JwtAuthGuard |
| `profileUpdated` | Subscription | по userId |
| `debug` | Query | JwtOptionalAuthGuard |
| `echo` | Query + Mutation | публичный |

**Инфраструктура (настроена, просто использовать):**
- Auth guards: `JwtAuthGuard`, `JwtOptionalAuthGuard`, `RolesGuard`, `@CurrentUser()`, `@Roles()`
- GraphQL: **MercuriusDriver** (Fastify) — не Apollo Server
- Subscriptions: WebSocket + Redis pub/sub (mqemitter-redis)
- Error handling: `gqlErrorFormatter` — маппит `ZodValidationException` и `HttpException` в GraphQL-ошибки
- 404 backend: `@All()` → `NotFoundException`
- Health check: `GET /health`
- Queue: BullMQ + Redis
- i18n: nestjs-i18n (en/ru)

**Backend тестовая инфраструктура:**
- `E2EClient.loginAs(profile)` — auth через `x-mock-sub` header
- `clearDatabase()` + `clearRedis()` — обязательно в `beforeEach()`
- `OIDC_MOCK_ENABLED=true` — режим мок-авторизации в `.env.test`

### Frontend

**FSD-слайсы** (FSD — Feature-Sliced Design, методология структуры фронтенд-кода по бизнес-слоям) **уже реализованы, не включать в Blueprint:**
| Слой | Слайс | Что содержит |
|------|-------|-------------|
| `features` | `auth` | OIDC PKCE конфигурация, `AuthGuard`, `MockAuthProvider` |
| `widgets` | `Header` | Навигация + кнопки авторизации (состояние auth) |
| `pages` | `404` | Полностью реализована с кнопками «Назад» и «На главную» |
| `shared/api` | `graphql-client` | URQL client factory с Bearer-token |

**Роуты (уже реализованы):**
| Роут | Назначение |
|------|-----------|
| `/callback` | OIDC redirect handler (после логина) |
| `*` (любой несуществующий) | 404 page через `defaultNotFoundComponent` |

**Frontend тестовая инфраструктура:**
- `VITE_MOCK_AUTH=true` — `MockAuthProvider` для Playwright e2e
- `tests/setup.ts` — глобальный мок `react-oidc-context` для Vitest unit/component-тестов
- Coverage: lines / functions / statements ≥ 80%, branches ≥ 70%

---

## Структура вывода

Технические контракты хранятся **отдельно** от продуктовой документации (doc-gen), чтобы не смешивать бизнес-артефакты с техническими:

```
docs/<имя_проекта>/          ← продуктовая документация (doc-gen)
├── INDEX.md
├── 1_PRODUCT_VISION/
└── 2_PRODUCT_SPEC/

blueprint/<имя_проекта>/     ← технические контракты (этот скилл)
└── 3_TECH_BLUEPRINT/
    ├── IMPLEMENTATION_GUIDE.md   ← онбординг разработчика (генерировать первым)
    ├── DATABASE_MODEL.md
    ├── API_CONTRACTS.md
    ├── ARCHITECTURE.md
    └── TESTING_PLAN.md
```

Создать папку перед генерацией:
```bash
mkdir -p blueprint/<ИмяПроекта>/3_TECH_BLUEPRINT
```

---

## Правила генерации: DATABASE_MODEL.md

### Структура файла

```markdown
# Модель данных: <Название продукта>

**Статус:** черновик | **Дата:** YYYY-MM-DD

## Ссылки
- Спецификация: [SPEC.md](../2_PRODUCT_SPEC/SPEC.md)

## Контекст реализации
- **ORM:** Prisma (TypeScript ORM) — схема в `prisma/schema.prisma`, миграции через `prisma migrate dev`
- **БД:** PostgreSQL
- **`Profile` — модель текущего пользователя** (уже существует в стартовом шаблоне, не дублировать): содержит `id`, `oidcSub @unique` (идентификатор из OIDC-провайдера), `roles`, `avatarUrl`. Для связи бизнес-сущности с пользователем: `profileId Int` + `@relation(fields: [profileId], references: [id])`.
- **Авторизация через OIDC** (OpenID Connect — протокол входа через внешний сервис, например Logto или Keycloak): таблицы `User`, `Session`, `AuthToken` — **не создавать**. Авторизация целиком вынесена на сторону внешнего провайдера.
- **Soft delete:** `deletedAt DateTime?` — сущность не удаляется физически. Запросы к таким моделям всегда фильтруют: `WHERE deletedAt IS NULL`.

## Сущности и отношения
<для каждой модели: что хранит, ключевые поля, связи — язык бизнеса, не базы данных>

## Схема базы данных

```prisma
// schema.prisma — <Название продукта>
// Сгенерировано на основе SPEC.md

generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

// <модели и энумы>
```

## Энумы
<для каждого enum: название, значения, когда и почему используется>

## Индексы и ограничения
| Модель | Поле(я) | Тип | Причина |
|--------|---------|-----|---------|

## Каскадные операции
| Модель | Поле связи | onDelete | onUpdate | Обоснование |
|--------|-----------|----------|----------|-------------|
```

### Обязательные правила

**Трассируемость.** В комментарии к каждой модели добавлять ссылку:
```prisma
// Сущность «Заказ» — SPEC.md §Сущности
model Order {
```

**Технические поля.** Каждая модель (кроме join-таблиц, см. ниже) обязана содержать:
```prisma
createdAt DateTime @default(now())
updatedAt DateTime @updatedAt
```

**Мягкое удаление.** Для критичных бизнес-сущностей (Пользователи, Заказы, Проекты и аналоги по SPEC.md) добавлять `deletedAt DateTime?` если SPEC.md не запрещает. Все запросы к таким моделям учитывают `WHERE deletedAt IS NULL`.

**Энумы вместо строк.** Для перечислимых значений — только `enum`, никаких `String`:
```prisma
// Правильно:
enum OrderStatus { DRAFT CONFIRMED COMPLETED CANCELLED }

// Запрещено:
status String // вместо enum
```

**Связи.** Каждая связь — с явным `@relation`, `onDelete`, `onUpdate`.

**Уникальные индексы.** `@@unique` для бизнес-уникальных комбинаций полей.

**Join-таблицы.** Именовать со знаком `_` в названии (например, `_UserToRole`, `Post_Tag`). Технические поля для них опциональны.

**Profile уже существует.** Не добавлять модель `User`, `Account`, `Session`, `AuthToken` — авторизация полностью на стороне внешнего OIDC-провайдера. Для связи бизнес-сущностей с пользователем использовать `profileId`:
```prisma
// Правильно:
model Order {
  // ...
  profileId Int
  profile   Profile @relation(fields: [profileId], references: [id], onDelete: Cascade)
}

// Запрещено:
model User { /* дублирует Profile */ }
model Session { /* OIDC — внешний провайдер */ }
```

**Запрещено:**
- `String` для перечислимых значений (статусы, роли, типы)
- Внешние ключи без явного `onDelete`
- Модели без поля `id`
- Модели `User`, `Account`, `Session`, `AuthToken` (дублируют boilerplate)

---

## Правила генерации: API_CONTRACTS.md

### Структура файла

```markdown
# API-контракты: <Название продукта>

**Статус:** черновик | **Дата:** YYYY-MM-DD

## Ссылки
- Спецификация: [SPEC.md](../2_PRODUCT_SPEC/SPEC.md)
- Модель данных: [DATABASE_MODEL.md](./DATABASE_MODEL.md)

## Контекст реализации
- **GraphQL-движок:** MercuriusDriver — реализация GraphQL для веб-фреймворка Fastify (используется вместо Apollo Server; поведение идентично, отличается только конфигурацией).
- **Гарды авторизации** (NestJS guard — декоратор, проверяющий права доступа перед выполнением резолвера):
  - `@UseGuards(JwtAuthGuard)` — требует действующий JWT-токен
  - `@UseGuards(JwtOptionalAuthGuard)` — авторизация необязательна (работает и с токеном, и без)
  - `@UseGuards(JwtAuthGuard) @Roles(ProfileRole.ADMIN)` — только для роли ADMIN
  - `@CurrentUser()` — декоратор параметра резолвера, возвращает текущий `Profile`
- **Операции стартового шаблона (уже готовы, не включать в Blueprint):** `me`, `updateProfile`, `profileUpdated`, `debug`, `echo`
- **Обработка ошибок:** `gqlErrorFormatter` автоматически преобразует `ZodValidationException` и `HttpException` в корректные GraphQL-ошибки — не нужно обрабатывать вручную.
- **Subscriptions:** работают через WebSocket + Redis pub/sub — инфраструктура уже настроена, просто использовать.
- **OIDC-операции** (вход/выход/регистрация): `login`, `register`, `logout`, `refreshToken` — **не описывать**, они находятся в OIDC-провайдере (внешний сервис).

## Обзор API
<количество операций, основные домены, принципиальные решения>

## Кастомные ошибки
| Код | Описание | Когда возникает |
|-----|----------|----------------|

## Схема GraphQL

```graphql
# schema.graphql — <Название продукта>
# Сгенерировано на основе SPEC.md

# <директивы, scalars, типы, inputs, queries, mutations>
```

## Описание операций
<для нетривиальных операций: входные данные, результат, правило доступа, возможные ошибки>
```

### Обязательные правила

**Трассируемость.** В комментарии к каждой Query и Mutation:
```graphql
# Описано в SPEC.md §Ключевые операции — создание заказа
createOrder(input: CreateOrderInput!): CreateOrderResult!
```

**Директивы доступа.** Каждая Query и Mutation помечается комментарием:
```graphql
# @public
products(limit: Int! = 20, offset: Int! = 0): ProductsPage!

# @auth
myOrders(limit: Int! = 20, offset: Int! = 0): OrdersPage!

# @auth @hasRole(ADMIN)
deleteUser(id: ID!): DeleteUserResult!
```
Если операция публичная — писать `# @public`. Без этого комментария операция считается непомеченной — ошибка.

**Строгая пагинация.** Любое поле, возвращающее список, обязано иметь пагинацию. Использовать wrapper-тип:
```graphql
# Правильно:
type Query {
  orders(limit: Int! = 20, offset: Int! = 0, filter: OrderFilterInput): OrdersPage!
}

type OrdersPage {
  items: [Order!]!
  total: Int!
  hasMore: Boolean!
}

# Запрещено:
type Query {
  orders: [Order!]!  # голый массив без пагинации — ошибка
}
```

**Input-типы.** Каждая мутация принимает отдельный Input-тип, не набор скалярных аргументов:
```graphql
# Правильно:
createOrder(input: CreateOrderInput!): CreateOrderResult!

# Запрещено:
createOrder(userId: ID!, productId: ID!, quantity: Int!): Order!
```

**Кастомные ошибки.** Для операций с бизнес-ошибками — union-тип результата:
```graphql
union CreateOrderResult = Order | OrderError | InsufficientStockError
```

**GraphQL-движок: MercuriusDriver (Fastify).** Не описывать директивы и механизмы Apollo Server. Subscriptions — через WebSocket + Redis pub/sub (уже настроено в boilerplate).

**Синтаксис guards в комментариях.** Использовать NestJS-style для точности:
```graphql
# @UseGuards(JwtAuthGuard)                      — требует авторизации
# @UseGuards(JwtOptionalAuthGuard)              — авторизация опциональна
# @UseGuards(JwtAuthGuard) @Roles(ADMIN)        — только для роли ADMIN
# @public                                        — без авторизации
```

**Операции boilerplate уже реализованы.** Не включать в Blueprint:
`me`, `updateProfile(input)`, `profileUpdated`, `debug`, `echo`

**Запрещено описывать:** операции авторизации (login, register, logout, refreshToken) — это OIDC-провайдер. Загрузку файлов и управление сессиями — если не несут уникальной бизнес-логики.

**Запрещено:**
- Голые массивы `[Entity!]!` в Query/Mutation
- Query/Mutation без директивы доступа
- Мутации с набором скалярных аргументов вместо Input-типа
- Операции `me`, `updateProfile`, `profileUpdated` (уже в boilerplate)
- GraphQL-типы `LoginInput`, `RegisterInput`, `SessionType` (OIDC — внешний провайдер)

---

## Правила генерации: ARCHITECTURE.md

### Структура файла

```markdown
# Архитектура: <Название продукта>

**Статус:** черновик | **Дата:** YYYY-MM-DD

## Ссылки
- Спецификация: [SPEC.md](../2_PRODUCT_SPEC/SPEC.md)

## Контекст реализации

**Стек:**
- Бэкенд: NestJS + Fastify, Prisma ORM, PostgreSQL, GraphQL (MercuriusDriver), Redis (BullMQ — очереди задач; pub/sub — GraphQL Subscriptions)
- Фронтенд: React + Vite, TanStack Router (файловый роутинг), FSD (Feature-Sliced Design — методология структуры кода по бизнес-слоям), URQL (GraphQL-клиент с кэшированием), Zustand (хранилище UI-состояния), ParaglideJS (i18n-библиотека для Vite)

**Уже реализовано в стартовых шаблонах (не включать в Blueprint):**

| Слой | Что готово |
|----------|------------|
| Бэкенд | Auth: гарды `JwtAuthGuard`, `JwtOptionalAuthGuard`, `RolesGuard`; OIDC JWT через внешний провайдер |
| Бэкенд | Profile GraphQL: операции `me`, `updateProfile`, `profileUpdated`; модель `Profile` в БД |
| Бэкенд | Инфраструктура: Health check `GET /health`, BullMQ, Redis pub/sub, i18n (en/ru), gqlErrorFormatter |
| Фронтенд | `features/auth` — OIDC PKCE-авторизация, `AuthGuard` (блокирует неавторизованных), `MockAuthProvider` (для тестов) |
| Фронтенд | `widgets/Header` (навигация + кнопки входа/выхода), `pages/404`, `shared/api/graphql-client` (URQL с Bearer-токеном) |
| Фронтенд | Роуты: `/callback` (OIDC redirect после входа), `*` (404 через `defaultNotFoundComponent`) |

**Protected routes (TanStack Router):** паттерн `beforeLoad` + `redirect()`. Контент страницы оборачивается в `AuthGuard` — компонент, перенаправляющий неавторизованного пользователя.

## Backend: NestJS-модули
| Модуль | Ответственность | Зависимости |
|--------|----------------|-------------|

## Frontend: FSD-слайсы

### Entities (бизнес-сущности)
| Сущность | Что представляет | Модель данных |
|---------|-----------------|---------------|

### Features (действия пользователя)
| Фича | Что делает пользователь | Связанная операция API |
|------|------------------------|----------------------|

### Widgets (составные блоки UI)
| Виджет | Из каких фич состоит | Где используется |
|--------|---------------------|-----------------|

### Pages (страницы приложения)
<иерархия страниц из prototype/ без файловых путей>

## Frontend: управление состоянием

### Серверный стейт (URQL-кэш)
<какие данные кэшируются через URQL: сущности, списки, их инвалидация>

### UI-стейт (Zustand)
<какие состояния в Zustand: открытые модалки, активные фильтры, тема, локальные флаги>

## Frontend: локализация (i18n)
<пространства имён ParaglideJS на основе прототипа или «i18n не требуется»>

## Переменные окружения
| Переменная | Тип | Назначение | BE/FE |
|-----------|-----|-----------|-------|
```

### Обязательные правила

**СТРОГИЙ ЗАПРЕТ файловых путей.** Запрещены любые строки вида `src/features/auth`, `src/entities/user`, `app/pages/dashboard`, `components/Button`. Только логические названия: фича `auth`, сущность `User`, страница `Dashboard`.

**Разделение состояния обязательно.** Явно указать для каждой области данных:
- Данные, пришедшие с сервера и кэшированные → URQL
- Локальное UI-состояние (не нужно сохранять на сервер) → Zustand

Нельзя оставить раздел «Управление состоянием» без конкретного содержимого или написать «используется URQL» без указания конкретных сущностей.

**FSD без деталей реализации.** Описывать только бизнес-смысл слайса. Не упоминать структуру файлов, barrel-exports, модель папок.

**ENV — только бизнес-переменные.** Не включать `PORT`, `NODE_ENV`, `DATABASE_URL`, `REDIS_URL`, стандартные переменные фреймворков. Только специфичные для данного продукта.

**i18n.** Если в прототипе есть тексты (labels, кнопки, сообщения, уведомления) — перечислить namespace'ы ParaglideJS. Если прототип без текстов — явно писать: «i18n не требуется».

**Роутер: TanStack Router (file-based).** Не описывать Next.js или React Router паттерны. Защищённые страницы используют `beforeLoad` + `redirect()`:
```
// Паттерн protected route (beforeLoad):
если не аутентифицирован → redirect на главную
контент страницы оборачивается в AuthGuard
```

**Уже существующие FSD-слайсы из стартового шаблона — не включать в Blueprint:**
- `features/auth` — OIDC-авторизация, `AuthGuard`, `MockAuthProvider`
- `widgets/Header` — навигация с кнопками входа/выхода
- `pages/404` — страница «не найдено»
- `shared/api/graphql-client` — URQL (GraphQL-клиент) с Bearer-токеном

**Zustand: паттерн со store.** Все новые store используют обёртку с devtools:
```typescript
// Паттерн нового Zustand-стора:
create(devtools<MyStore>((set) => ({ ... })))
```

**Запрещено:**
- Файловые пути FSD (`src/`, `app/`, `components/`)
- Раздел «Управление состоянием» без разделения URQL/Zustand
- Технические паттерны реализации (DI-контейнеры, репозитории, паттерны)
- Упоминание библиотек вне утверждённого стека
- Включать в Blueprint уже существующие FSD-слайсы (`auth`, `Header`, `404`)

---

## Правила генерации: TESTING_PLAN.md

### Структура файла

```markdown
# План тестирования: <Название продукта>

**Статус:** черновик | **Дата:** YYYY-MM-DD

## Ссылки
- Спецификация: [SPEC.md](../2_PRODUCT_SPEC/SPEC.md) — источник сценариев

## Контекст тестирования

**Бэкенд (Vitest):**
- `E2EClient` — встроенная тест-утилита из `test/utils/`, выполняет GraphQL-запросы с авторизацией через специальный HTTP-заголовок `x-mock-sub` (без реального OIDC-провайдера).
- Использование: `await client.loginAs(profile)` — выполнить запросы от имени пользователя.
- Обязательно в `beforeEach()`: `clearDatabase()` (очистить таблицы БД) + `clearRedis()` (сбросить кэш Redis) — чтобы тесты не влияли друг на друга.
- В `.env.test`: `OIDC_MOCK_ENABLED=true` — включает мок-режим авторизации, JWT проверяется через `x-mock-sub`, не через реальный OIDC.

**Фронтенд (Vitest + React Testing Library + Playwright):**
- Unit/component тесты: `tests/setup.ts` глобально мокает `react-oidc-context` (библиотека OIDC-авторизации) — OIDC-состояние доступно в тестах без дополнительной настройки.
- E2E тесты (Playwright): запускать с `VITE_MOCK_AUTH=true` — активирует `MockAuthProvider`, который автоматически авторизует пользователя в браузере без реального OIDC-провайдера.

**Coverage targets:** lines / functions / statements ≥ 80%, branches ≥ 70%

## Unit-тесты: сложная бизнес-логика
| Что тестируем | Входные данные | Ожидаемый результат | Почему нетривиально |
|--------------|----------------|--------------------|--------------------|

## E2E-сценарии (Playwright)
| Сценарий | Шаги | Ожидаемый результат | Источник в SPEC.md |
|----------|------|--------------------|--------------------|

## Критические пути (из SPEC.md §Тестирование)
<полный перенос критических сценариев без сокращений>

## Негативные сценарии (из SPEC.md §Тестирование)
<полный перенос негативных сценариев без сокращений>
```

### Обязательные правила

- Unit-тесты — только для нетривиальной логики: расчёты, валидации, конечные автоматы состояний. Не тестировать простой CRUD.
- E2E-сценарии берутся из SPEC.md §Тестирование. Каждый сценарий ссылается на источник.
- Критические пути и негативные сценарии из SPEC.md переносятся полностью.

### Backend: конкретные паттерны

**E2E-тесты** используют готовую инфраструктуру из `test/utils/`:
```typescript
// Аутентификация в тесте:
await client.loginAs(profile)

// Запрос:
const result = await client.requestGraphQL<MyQuery>(query, variables)

// Очистка (обязательно в beforeEach):
await clearDatabase()
await clearRedis()
```

**Мок авторизации:** `OIDC_MOCK_ENABLED=true` + `x-mock-sub` header. Не писать собственный механизм мока.

**Unit-тесты:** Vitest + NestJS Testing Module. Моки сервисов — через фабрики из `test/utils/mocks.ts`.

### Frontend: конкретные паттерны

**Три уровня тестирования:**
| Уровень | Инструмент | Что тестировать |
|---------|-----------|----------------|
| Unit | Vitest | Store-логика, утилиты, трансформации данных |
| Component | Vitest + React Testing Library | Рендер компонентов, взаимодействие, состояния |
| E2E | Playwright | Пользовательские сценарии из SPEC.md |

**Auth в тестах:**
- Unit/component: OIDC автоматически замокан через `tests/setup.ts` (unauthenticated по умолчанию). Переопределять локально в конкретном тесте.
- E2E: собирать приложение с `VITE_MOCK_AUTH=true` — `MockAuthProvider` авторизует автоматически.

**Coverage targets (обязательно):**
- lines / functions / statements: ≥ 80%
- branches: ≥ 70%

---

## Правила генерации: IMPLEMENTATION_GUIDE.md

### Назначение

Онбординг-документ: позволяет новому разработчику взять Blueprint и начать работу **без знакомства с boilerplate**. Генерировать **последним** — после DATABASE_MODEL.md, API_CONTRACTS.md, ARCHITECTURE.md, TESTING_PLAN.md, так как суммирует их содержимое.

### Структура файла

```markdown
# Руководство по реализации: <Название продукта>

**Статус:** черновик | **Дата:** YYYY-MM-DD

## Стек

| Слой | Технология | Назначение |
|------|-----------|-----------|
| Бэкенд | NestJS + Fastify | Node.js-фреймворк для API-сервера |
| GraphQL | MercuriusDriver | GraphQL-сервер (адаптер Fastify) |
| ORM | Prisma | TypeScript ORM для работы с БД |
| БД | PostgreSQL | Реляционная база данных |
| Очереди | BullMQ + Redis | Фоновые задачи и очереди |
| Фронтенд | React + Vite | UI-библиотека + сборщик |
| Роутер | TanStack Router | Файловый роутинг для React |
| GraphQL-клиент | URQL | Запросы к API с кэшированием |
| UI-состояние | Zustand | Хранилище локального состояния |
| i18n | ParaglideJS | Локализация (Vite-плагин) |
| Тесты бэкенда | Vitest + E2EClient | Юнит и E2E для бэкенда |
| Тесты фронтенда | Vitest + RTL + Playwright | Юнит, компонент, E2E для фронтенда |

## Что уже реализовано

> Эти части **не нужно писать** — они готовы в стартовых шаблонах проекта.

### Бэкенд
- **Авторизация (OIDC JWT):** вход через внешний OIDC-провайдер (Logto/Keycloak). Гарды NestJS: `JwtAuthGuard` (требует токен), `JwtOptionalAuthGuard` (необязательно), `RolesGuard` (по роли). Декораторы резолвера: `@CurrentUser()` (получить текущий Profile), `@Roles(ProfileRole.ADMIN)` (ограничить по роли).
- **Profile — модель пользователя:** `Profile { id, oidcSub @unique, roles, avatarUrl }`. Уже реализованы GraphQL-операции: `me` (получить свой профиль), `updateProfile` (обновить), `profileUpdated` (подписка на изменения).
- **Инфраструктура:** Health check `GET /health`, BullMQ (очереди задач на Redis), Redis pub/sub (для GraphQL Subscriptions), nestjs-i18n en/ru, `gqlErrorFormatter` (автоматический маппинг ошибок в GraphQL-формат).

### Фронтенд
- **Авторизация (OIDC PKCE):** `react-oidc-context` — библиотека для OAuth2/OIDC в React. FSD-слайс `features/auth` содержит: `AuthGuard` (блокирует доступ неавторизованным), `MockAuthProvider` (автоматическая авторизация в тестах).
- **Готовые UI-компоненты:** `widgets/Header` (навигация с кнопками входа/выхода), `pages/404` (страница ошибки с кнопками «Назад» и «На главную»).
- **GraphQL API:** `shared/api/graphql-client` — настроенный URQL-клиент (GraphQL-клиент для React) с автоматической подстановкой Bearer-токена авторизации.
- **Роутинг (TanStack Router):** `/callback` — обработчик OIDC-редиректа после входа; `*` — 404-заглушка через `defaultNotFoundComponent`.

## Что нужно реализовать

> Всё описанное в данном Blueprint.

### Backend-модули
<список из ARCHITECTURE.md §Backend: NestJS-модули>

### Frontend-слайсы (новые)
<список из ARCHITECTURE.md §Frontend: FSD-слайсы — только те, которых нет в boilerplate>

## Локальный запуск

### Требования
- Node.js 20+, pnpm
- Docker (PostgreSQL + Redis)
- OIDC: для разработки используется mock-режим

### Backend
```bash
pnpm install
cp .env.example .env
# Заполнить: DATABASE_URL, REDIS_URL, JWT_SECRET, OIDC_ISSUER

docker compose up -d postgres redis
pnpm prisma migrate dev
pnpm start:dev
```

### Frontend
```bash
pnpm install
cp .env.example .env
# Заполнить: VITE_API_URL, VITE_OIDC_AUTHORITY, VITE_OIDC_CLIENT_ID

VITE_MOCK_AUTH=true pnpm dev   # разработка с мок-авторизацией
```

## Тестирование

### Backend
```bash
pnpm test         # unit (Vitest)
pnpm test:e2e     # e2e (требует Postgres + Redis)
```

### Frontend
```bash
pnpm test                              # unit + component (Vitest + RTL)
VITE_MOCK_AUTH=true pnpm test:e2e      # e2e (Playwright)
```

**Coverage:** lines / functions / statements ≥ 80%, branches ≥ 70%

## Ключевые решения

<Обоснование нетривиальных технических решений.
Пример: «Soft delete для Order — SPEC.md §Данные запрещает физическое удаление».
Если нетривиальных решений нет — написать «Особых нетривиальных решений нет».>

## Ссылки

- Спецификация: [SPEC.md](../../docs/<имя>/2_PRODUCT_SPEC/SPEC.md)
- [DATABASE_MODEL.md](./DATABASE_MODEL.md)
- [API_CONTRACTS.md](./API_CONTRACTS.md)
- [ARCHITECTURE.md](./ARCHITECTURE.md)
- [TESTING_PLAN.md](./TESTING_PLAN.md)
```

### Обязательные правила

- `## Стек` — полная таблица технологий без сокращений
- `## Что уже реализовано` — точный список boilerplate: разработчик видит, что **не нужно писать**
- `## Что нужно реализовать` — конкретный список модулей и слайсов из ARCHITECTURE.md, не общие фразы
- `## Локальный запуск` — реальные shell-команды с `.env` и docker; не «см. README»
- `## Тестирование` — команды запуска + coverage targets
- `## Ключевые решения` — не пустой раздел: либо обоснование решений, либо явная фраза «нет нетривиальных решений»

**Запрещено:**
- Пустые разделы
- `## Что нужно реализовать` без списка конкретных модулей/слайсов
- Общие фразы вместо команд в `## Локальный запуск`
- Дублирование полного содержимого других документов Blueprint

---

## Адаптивность: режим обновления проекта

Если в проекте уже существуют `schema.prisma` или `schema.graphql`:

1. **Не переписывать схему с нуля.** Прочитать существующие схемы, описать только дельту.
2. **DATABASE_MODEL.md обязан содержать раздел `## План миграции`:**
   ```markdown
   ## План миграции
   
   ### Новые таблицы
   - `Notification` — уведомления пользователей
   
   ### Изменения существующих таблиц
   - `User` — добавить поле `avatarUrl String?`
   - `Order` — добавить enum-значение `OrderStatus.REFUNDED`
   
   ### Опасные операции
   - Нет
   ```
3. **API_CONTRACTS.md** содержит разделы `## Новые операции` и `## Изменённые операции`.
4. Существующие схемы не ломать — только расширять.

---

## Проверка ТЗ по запросу

Когда пользователь говорит **«проверь ТЗ»**, **«валидируй blueprint»**, **«check blueprint»**, **«посмотри что я изменил»** — выполнить следующую последовательность:

### Шаг 1. Запустить скрипт-валидатор

```bash
# Стандартный режим:
python3 {BLUEPRINT} validate "ИмяПроекта"

# Если проект — обновление существующего:
python3 {BLUEPRINT} validate "ИмяПроекта" --update-mode
```

### Шаг 2. Прочитать все 5 файлов полностью

`IMPLEMENTATION_GUIDE.md`, `DATABASE_MODEL.md`, `API_CONTRACTS.md`, `ARCHITECTURE.md`, `TESTING_PLAN.md`

### Шаг 3. Ручная проверка противоречий с boilerplate

- [ ] Нет моделей `User`, `Session`, `AuthToken` — пользователь только через `Profile` (oidcSub)
- [ ] Нет GraphQL-операций `login`, `register`, `logout`, `refreshToken` — OIDC внешний
- [ ] Нет операций `me`, `updateProfile`, `profileUpdated` — уже в boilerplate
- [ ] В ARCHITECTURE.md нет `features/auth`, `widgets/Header`, `pages/404` — уже есть
- [ ] Protected routes описаны через `beforeLoad` паттерн
- [ ] Тесты упоминают `E2EClient.loginAs()` (BE) или `VITE_MOCK_AUTH=true` (FE)
- [ ] Нет несовместимых библиотек (Apollo вместо Mercurius, React Router вместо TanStack)

### Шаг 4. Проверка согласованности со SPEC.md

- [ ] Все сущности SPEC.md §Сущности покрыты в DATABASE_MODEL.md
- [ ] Все операции SPEC.md §Ключевые операции присутствуют в API_CONTRACTS.md
- [ ] Все страницы SPEC.md §Страницы отражены в ARCHITECTURE.md
- [ ] Сценарии SPEC.md §Тестирование перенесены в TESTING_PLAN.md

### Шаг 5. Выдать структурированный отчёт

```
## Проверка Blueprint: <ИмяПроекта>

### Скрипт-валидатор
✓ Прошёл без ошибок
— или —
✗ Ошибки: <список>
⚠ Предупреждения: <список>

### Противоречия с boilerplate
✓ Не обнаружены
— или —
✗ Обнаружены:
  - <проблема> → <как исправить>

### Согласованность со SPEC.md
✓ Все требования покрыты
— или —
⚠ Не покрыто:
  - <что отсутствует в Blueprint>

### Итог
Готов к разработке / Требует исправлений: <N проблем>
```

После исправлений — повторно запустить валидатор и создать git-коммит.

---

## Процесс работы

> **Скрипт:** `~/.config/opencode/skills/tech-blueprint/scripts/blueprint_validator.py`
> Скрипт НЕ копируется в проект — используется напрямую из папки скилла.
> Далее: `{BLUEPRINT}` = полный путь выше.

### Шаг 1. Анализ входных данных

1. Прочитать VISION.md и SPEC.md целиком — это **единственный источник истины**
2. Проверить `prototype/`: зафиксировать список страниц и переходов между ними. **Не переносить** реализацию — только скоп. Прототип — черновик, архитектурные решения принимаются независимо
3. Проверить наличие `schema.prisma`, `schema.graphql` → определить режим (новый / обновление)
4. Составить рабочий список: сущности, операции, страницы, роли

### Шаг 2. Генерация документов

Заполнять строго в порядке:
```
DATABASE_MODEL.md → API_CONTRACTS.md → ARCHITECTURE.md → TESTING_PLAN.md → IMPLEMENTATION_GUIDE.md
```
Каждый следующий документ опирается на предыдущий (API — на модели БД, архитектура — на API). IMPLEMENTATION_GUIDE.md — последним: он суммирует все четыре документа.

### Шаг 3. Самопроверка перед валидатором

- [ ] Каждая Prisma-модель (кроме join-таблиц с `_` в названии) имеет `createdAt`/`updatedAt`
- [ ] Для критичных сущностей добавлено `deletedAt DateTime?`
- [ ] Нет голых массивов `[Type!]!` в Query/Mutation без пагинации
- [ ] Каждая Query/Mutation имеет комментарий с директивой доступа
- [ ] В ARCHITECTURE.md нет ни одного файлового пути (`src/`, `app/`)
- [ ] В DATABASE_MODEL.md и API_CONTRACTS.md есть ссылки на SPEC.md
- [ ] Если режим обновления — раздел `## План миграции` существует
- [ ] IMPLEMENTATION_GUIDE.md содержит разделы `## Стек`, `## Что уже реализовано`, `## Локальный запуск`

### Шаг 4. Валидация (обязательно)

```bash
# Новый проект:
python3 {BLUEPRINT} validate "ИмяПроекта"

# Обновление существующего проекта:
python3 {BLUEPRINT} validate "ИмяПроекта" --update-mode

# Нестандартная папка вывода:
python3 {BLUEPRINT} validate "ИмяПроекта" --output /путь/к/blueprint
```

**Документация не считается готовой, пока валидация не пройдена без ошибок.**

Скрипт выполняет 8 или 9 проверок:
1. Наличие всех 5 файлов (включая IMPLEMENTATION_GUIDE.md)
2. Наличие блоков ` ```prisma ` и ` ```graphql `
3. Отсутствие FSD-путей в ARCHITECTURE.md
4. Кросс-чек: Prisma-модели покрыты в API_CONTRACTS.md (≥50%)
5. Трассируемость: ссылки на SPEC.md в DATABASE_MODEL.md и API_CONTRACTS.md
6. Пагинация: все Query/Mutation-поля с `[...]` имеют аргументы пагинации
7. Технические поля: `createdAt`/`updatedAt` в каждой Prisma-модели (кроме join-таблиц)
8. IMPLEMENTATION_GUIDE.md содержит обязательные разделы (`## Стек`, `## Что уже реализовано`, `## Локальный запуск`)
9. *(только с `--update-mode`)* Наличие раздела «План миграции» в DATABASE_MODEL.md

### Шаг 5. Git-коммит (обязательно)

```bash
git add blueprint/<ИмяПроекта>/3_TECH_BLUEPRINT/
git commit -m "feat(blueprint): технические контракты <ИмяПроекта>"
```

---

## Глоссарий

| Термин | Значение |
|--------|---------|
| Boilerplate | Стартовый шаблон проекта с готовыми базовыми решениями (auth, инфраструктура) |
| OIDC | OpenID Connect — протокол авторизации через внешний сервис (Logto, Keycloak, Google) |
| JWT | JSON Web Token — подписанный токен, подтверждающий личность пользователя |
| FSD | Feature-Sliced Design — методология структуры фронтенда по слоям: app, pages, widgets, features, entities, shared |
| MercuriusDriver | NestJS-адаптер GraphQL для веб-фреймворка Fastify (альтернатива Apollo Server) |
| URQL | GraphQL-клиент для React с встроенным кэшем и инвалидацией |
| Zustand | Минималистичная библиотека состояния для React (альтернатива Redux) |
| ParaglideJS | i18n-библиотека с типизированными ключами переводов, работает как Vite-плагин |
| BullMQ | Библиотека очередей задач для Node.js на основе Redis |
| E2EClient | Встроенная тест-утилита (`test/utils/`) для выполнения GraphQL-запросов с мок-авторизацией |
| Profile | Модель пользователя в БД; связана с OIDC через поле `oidcSub @unique` |
| TanStack Router | Файловый роутер для React: маршруты определяются структурой папок `src/routes/` |
| AuthGuard | FSD-слайс-компонент, перенаправляющий неавторизованного пользователя с защищённой страницы |
| beforeLoad | Хук TanStack Router для проверки условий перед загрузкой страницы (используется для защиты роутов) |

---

## Ключевые ограничения

- Язык документации — **русский**, код — латиница
- Стек: **NestJS, Prisma, PostgreSQL, GraphQL (MercuriusDriver), React, TanStack Router, FSD, URQL, Zustand, ParaglideJS**
- **Запрещены файловые пути FSD** в ARCHITECTURE.md — только логические названия
- **Все Query/Mutation с `[...]`** — обязательно с пагинацией (limit/offset или first/after)
- **Все Query/Mutation** — с комментарием-директивой доступа (`@public`, `@UseGuards(JwtAuthGuard)`, `@Roles(ADMIN)`)
- **Каждая Prisma-модель** (не join) — с `createdAt`/`updatedAt`
- **Enum вместо String** для всех перечислимых значений
- **Трассируемость** — ссылки на SPEC.md в комментариях к схемам
- **Режим обновления** — раздел «План миграции», не переписывание схемы
- **Не дублировать boilerplate**: Profile, me/updateProfile/profileUpdated, auth-слайс, Header, 404-страница
- **Тесты**: использовать `E2EClient.loginAs()` (BE) и `VITE_MOCK_AUTH=true` (FE), coverage ≥ 80%/70%
- После валидации — **git-коммит**
