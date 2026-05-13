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

Скилл создаёт четыре документа, полностью описывающих техническую реализацию продукта:
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

## Boilerplate: уже реализовано

Оба шаблона (liteend / litefront) содержат готовый код. **Не проектировать заново** то, что уже есть — только расширять.

### Backend (liteend)

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

### Frontend (litefront)

**FSD-слайсы (уже реализованы, не включать в Blueprint):**
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

**Уже существующие FSD-слайсы boilerplate — не включать в Blueprint:**
- `features/auth` — OIDC, `AuthGuard`, `MockAuthProvider`
- `widgets/Header` — навигация с auth-кнопками
- `pages/404` — страница не найдено
- `shared/api/graphql-client` — URQL с Bearer-token

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

### Backend: конкретные паттерны (liteend)

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

### Frontend: конкретные паттерны (litefront)

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

### Шаг 2. Прочитать все 4 файла полностью

`DATABASE_MODEL.md`, `API_CONTRACTS.md`, `ARCHITECTURE.md`, `TESTING_PLAN.md`

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
DATABASE_MODEL.md → API_CONTRACTS.md → ARCHITECTURE.md → TESTING_PLAN.md
```
Каждый следующий документ опирается на предыдущий (API — на модели БД, архитектура — на API).

### Шаг 3. Самопроверка перед валидатором

- [ ] Каждая Prisma-модель (кроме join-таблиц с `_` в названии) имеет `createdAt`/`updatedAt`
- [ ] Для критичных сущностей добавлено `deletedAt DateTime?`
- [ ] Нет голых массивов `[Type!]!` в Query/Mutation без пагинации
- [ ] Каждая Query/Mutation имеет комментарий с директивой доступа
- [ ] В ARCHITECTURE.md нет ни одного файлового пути (`src/`, `app/`)
- [ ] В DATABASE_MODEL.md и API_CONTRACTS.md есть ссылки на SPEC.md
- [ ] Если режим обновления — раздел `## План миграции` существует

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

Скрипт выполняет 7 или 8 проверок:
1. Наличие всех 4 файлов
2. Наличие блоков ` ```prisma ` и ` ```graphql `
3. Отсутствие FSD-путей в ARCHITECTURE.md
4. Кросс-чек: Prisma-модели покрыты в API_CONTRACTS.md (≥50%)
5. Трассируемость: ссылки на SPEC.md в DATABASE_MODEL.md и API_CONTRACTS.md
6. Пагинация: все Query/Mutation-поля с `[...]` имеют аргументы пагинации
7. Технические поля: `createdAt`/`updatedAt` в каждой Prisma-модели (кроме join-таблиц)
8. *(только с `--update-mode`)* Наличие раздела «План миграции» в DATABASE_MODEL.md

### Шаг 5. Git-коммит (обязательно)

```bash
git add blueprint/<ИмяПроекта>/3_TECH_BLUEPRINT/
git commit -m "feat(blueprint): технические контракты <ИмяПроекта>"
```

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
