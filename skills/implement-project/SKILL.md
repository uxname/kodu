---
name: implement-project
description: Реализует полный проект по готовым документации, ТЗ и прототипу. Инициализирует бэкенд (liteend-init) и фронтенд (litefront-init), реализует все сущности/операции/страницы, покрывает тестами, верифицирует соответствие документам. Запускай когда VISION.md + SPEC.md готовы и tech-blueprint утверждён. НЕ запускай если ТЗ ещё черновик или проект уже частично реализован.
license: MIT
compatibility: opencode
metadata:
  level: multi
  output: папка projects/<ИмяПроекта>/
---

## Назначение

Скилл берёт готовые артефакты проектирования и **реализует проект от нуля до работающего состояния с тестами**:
- Инициализирует бэкенд и фронтенд из стартовых шаблонов
- Имплементирует всю бизнес-логику строго по ТЗ
- Покрывает тестами согласно TESTING_PLAN.md
- Верифицирует соответствие документации и запускает полный прогон проверок

**Когда запускать:**
- `docs/<name>/` и `blueprint/<name>/3_TECH_BLUEPRINT/` готовы и утверждены
- ТЗ прошло валидацию `blueprint_validator.py` без ошибок
- Команда готова к разработке

**Когда НЕ запускать:**
- ТЗ в черновике или не прошло валидацию
- Проект уже частично реализован (это не скилл для рефакторинга)
- Задача — добавить одну фичу, а не построить проект с нуля

---

## Входные данные

Перед началом **обязательно прочитать** все документы:

```
docs/<ИмяПроекта>/
├── 1_PRODUCT_VISION/VISION.md           ← бизнес-цели, границы, роли
└── 2_PRODUCT_SPEC/SPEC.md               ← сущности, операции, страницы, бизнес-правила

blueprint/<ИмяПроекта>/3_TECH_BLUEPRINT/
├── IMPLEMENTATION_GUIDE.md              ← стек, что уже готово, команды запуска
├── DATABASE_MODEL.md                    ← Prisma-схема всех моделей
├── API_CONTRACTS.md                     ← GraphQL-схема, guards, пагинация
├── ARCHITECTURE.md                      ← NestJS-модули, FSD-слайсы, состояние
└── TESTING_PLAN.md                      ← unit-тесты, E2E-сценарии, coverage

prototype/                               ← UI-прототип (только для понимания UX-потоков)
```

Если любой из первых шести документов отсутствует — **остановиться** и сообщить пользователю какого файла не хватает. Прототип опционален.

---

## Структура вывода

```
projects/<ИмяПроекта>/
├── backend/     ← NestJS + Fastify API (инициализируется через liteend-init)
└── frontend/    ← React SPA (инициализируется через litefront-init)
```

Создать корневую папку до инициализации:
```bash
mkdir -p projects/<ИмяПроекта>
```

---

## Процесс реализации

### Шаг 0. Анализ документации

1. Прочитать все документы целиком
2. Составить внутренний рабочий список:
   - Prisma-модели из DATABASE_MODEL.md (без Profile/ProfileRole — они уже в шаблоне)
   - GraphQL-операции по доменам из API_CONTRACTS.md (без `me`, `updateProfile`, `profileUpdated`, `debug`, `echo`)
   - NestJS-модули из ARCHITECTURE.md §Backend
   - FSD-слайсы для реализации из ARCHITECTURE.md §Frontend (без `features/auth`, `widgets/Header`, `pages/404`, `shared/api/graphql-client`)
   - Тест-сценарии из TESTING_PLAN.md
3. Зафиксировать: всё из стартовых шаблонов — **не трогать и не дублировать**

---

### Шаг 1. Инициализация бэкенда

Запустить скилл **liteend-init** (`~/.config/opencode/skills/liteend-init/SKILL.md`):

```
project_name: projects/<ИмяПроекта>/backend
use_docker:   true
install_deps: true
```

После завершения:
- Прочитать `projects/<ИмяПроекта>/backend/AGENTS.md` — понять доступные команды проекта
- Убедиться что `GET /health` отвечает 200 (бэкенд запустился)

---

### Шаг 2. Реализация бэкенда

**2.1. Схема базы данных**

Открыть `backend/prisma/schema.prisma` и **добавить** новые модели из DATABASE_MODEL.md к существующим:
- Существующие модели `Profile`, `ProfileRole` — не изменять
- Для связи новых сущностей с пользователем: `profileId Int` + `@relation(fields: [profileId], references: [id], onDelete: Cascade)`
- Все связи — с явным `onDelete`; все перечислимые значения — только через `enum`

Применить миграцию:
```bash
# внутри backend/
npm run db:migrations:apply
```

**2.2. NestJS-модули**

Для каждого модуля из ARCHITECTURE.md §Backend создать структуру:
```
src/modules/<domain>/
├── <domain>.module.ts       ← @Module({ imports, providers, exports })
├── <domain>.resolver.ts     ← @Resolver() с GraphQL-операциями
├── <domain>.service.ts      ← бизнес-логика, Prisma-запросы
└── dto/
    └── <entity>.input.ts    ← ZodDto или class-validator Input для мутаций
```

Правила реализации резолверов:
- Guards соответствуют директивам доступа из API_CONTRACTS.md:
  ```typescript
  @UseGuards(JwtAuthGuard)         // # @auth в API_CONTRACTS.md
  @UseGuards(JwtOptionalAuthGuard) // # @auth? в API_CONTRACTS.md
  @Roles(ProfileRole.ADMIN)        // # @auth @hasRole(ADMIN)
  // без guard                     // # @public
  ```
- Текущий пользователь: `@CurrentUser() profile: Profile`
- Ошибки: выбрасывать `HttpException` или `ZodValidationException` — `gqlErrorFormatter` сам преобразует
- Soft delete: при `deletedAt DateTime?` в модели → фильтровать `{ deletedAt: null }` во **всех** запросах Prisma
- Пагинация: все списочные операции принимают `limit`/`offset` и возвращают `{ items, total, hasMore }`

**2.3. Subscriptions**

Если в API_CONTRACTS.md есть Subscription-операции — реализовывать через Redis pub/sub (уже настроен). Не добавлять новых transport-зависимостей.

**2.4. Коммит бэкенда**
```bash
git add .
git commit -m "feat(backend): реализация бизнес-логики <ИмяПроекта>"
```

---

### Шаг 3. Инициализация фронтенда

Запустить скилл **litefront-init** (`~/.config/opencode/skills/litefront-init/SKILL.md`):

```
project_name: projects/<ИмяПроекта>/frontend
install_deps: true
setup_env:    true
run_dev:      false
```

Настроить `.env` фронтенда:
```
VITE_GRAPHQL_API_URL=http://localhost:<PORT>/graphql
VITE_BASE_URL=http://localhost:<FRONT_PORT>
VITE_OIDC_AUTHORITY=<из IMPLEMENTATION_GUIDE.md>
VITE_OIDC_CLIENT_ID=<из IMPLEMENTATION_GUIDE.md>
```

Запустить бэкенд в фоне и сгенерировать GraphQL-типы:
```bash
# в backend/:
npm run start:dev &

# в frontend/:
npm run gen
```

Прочитать `projects/<ИмяПроекта>/frontend/AGENTS.md` — понять доступные команды.

---

### Шаг 4. Реализация фронтенда

**4.1. FSD-слайсы (только новые)**

Для каждого слайса из ARCHITECTURE.md §Frontend: FSD-слайсы создать структуру:
```
src/
├── entities/<name>/
│   ├── api/        ← URQL useQuery / useMutation с типами из npm run gen
│   └── model/      ← TypeScript-типы, трансформации данных
├── features/<name>/
│   ├── ui/         ← React-компоненты
│   └── model/      ← Zustand-стор или локальное состояние
└── widgets/<name>/
    └── ui/
```

Правила:
- GraphQL-запросы: только через URQL, типы из `npm run gen` — без `any`
- Zustand-стор: `create(devtools<MyStore>((set, get) => ({ ... })))` — только UI-состояние, не дублировать серверные данные
- Компоненты: из прототипа брать только UX-логику (потоки, формы), не копировать код

**4.2. Страницы и роутинг**

Для каждой страницы из ARCHITECTURE.md §Pages создать файл в `src/routes/`:

```typescript
// Защищённая страница (паттерн beforeLoad):
export const Route = createFileRoute('/my-page')({
  beforeLoad: ({ context: { auth } }) => {
    if (!auth.isAuthenticated) throw redirect({ to: '/' })
  },
  component: () => <AuthGuard><MyPage /></AuthGuard>,
})

// Публичная страница:
export const Route = createFileRoute('/public-page')({
  component: MyPublicPage,
})
```

**4.3. Переменные окружения**

Добавить в `.env` переменные из ARCHITECTURE.md §Переменные окружения (только бизнес-специфичные).

**4.4. i18n**

Если ARCHITECTURE.md §Frontend: локализация содержит namespace'ы — создать файлы переводов в `messages/`.

**4.5. Коммит фронтенда**
```bash
git add .
git commit -m "feat(frontend): реализация <ИмяПроекта>"
```

---

### Шаг 5. Тесты

Реализовать все сценарии из TESTING_PLAN.md. **Не пропускать сценарии** — каждый описан в документе.

**Backend E2E (Vitest + E2EClient):**
```typescript
describe('<Domain>Resolver', () => {
  let client: E2EClient

  beforeEach(async () => {
    await clearDatabase()
    await clearRedis()
    client = await createTestClient()
  })

  it('успешный сценарий', async () => {
    const profile = await createTestProfile()
    await client.loginAs(profile)
    const result = await client.requestGraphQL<QueryType>(QUERY, vars)
    expect(result.data).toBeDefined()
  })

  it('негативный сценарий — нет прав', async () => {
    const result = await client.requestGraphQL<QueryType>(QUERY, vars)
    expect(result.errors?.[0].message).toContain('Unauthorized')
  })
})
```

**Backend unit (Vitest + NestJS Testing Module):**
```typescript
describe('<Domain>Service.complexMethod', () => {
  it('корректно рассчитывает ...', () => {
    const result = service.complexMethod(input)
    expect(result).toEqual(expected)
  })
})
```

**Frontend component (Vitest + React Testing Library):**
```typescript
describe('<ComponentName>', () => {
  it('отображает состояние загрузки', () => {
    render(<MyComponent loading />)
    expect(screen.getByRole('progressbar')).toBeInTheDocument()
  })
})
// OIDC автоматически замокан через tests/setup.ts — дополнительная настройка не нужна
```

**Frontend E2E (Playwright):**
```typescript
// Запускать с VITE_MOCK_AUTH=true — MockAuthProvider авторизует автоматически
test('критический путь: <название>', async ({ page }) => {
  await page.goto('/target-page')
  await page.getByRole('button', { name: 'Действие' }).click()
  await expect(page.getByText('Успех')).toBeVisible()
})
```

**Коммит тестов:**
```bash
git add .
git commit -m "test(<ИмяПроекта>): покрытие по TESTING_PLAN.md"
```

---

### Шаг 6. Верификация (обязательна — не пропускать)

**6.1. Автоматические проверки**

Запустить в обоих проектах:
```bash
# Бэкенд:
cd projects/<ИмяПроекта>/backend && npm run check && npm run test:all

# Фронтенд:
cd projects/<ИмяПроекта>/frontend && npm run check && npm run test:all
```

**Если хотя бы одна команда завершается с ошибкой — устранить причину и запустить повторно.** Проект не считается реализованным, пока все проверки не зелёные.

**6.2. Ручной чеклист по документам**

Сверить реализацию с каждым документом:

**DATABASE_MODEL.md:**
- [ ] Каждая Prisma-модель из документа существует в `schema.prisma`
- [ ] Все миграции применены — `prisma migrate status` без pending
- [ ] Индексы `@@index`, уникальные ограничения `@@unique` соответствуют документу
- [ ] Каскадные операции `onDelete`/`onUpdate` соответствуют документу

**API_CONTRACTS.md:**
- [ ] Каждая Query/Mutation/Subscription реализована в соответствующем резолвере
- [ ] Guards соответствуют директивам (`# @auth` → `JwtAuthGuard`, `# @public` → без guard)
- [ ] Типы возвращаемых значений совпадают (включая wrapper-типы пагинации)
- [ ] Input-типы мутаций соответствуют схеме

**ARCHITECTURE.md:**
- [ ] Все NestJS-модули из §Backend реализованы и зарегистрированы в AppModule
- [ ] Все FSD-слайсы из §Frontend реализованы (кроме boilerplate-слайсов)
- [ ] Все страницы из §Pages существуют и маршрутизируются корректно
- [ ] Серверный стейт управляется через URQL, UI-стейт — через Zustand

**TESTING_PLAN.md:**
- [ ] Все сценарии из §Unit-тесты реализованы
- [ ] Все сценарии из §E2E-сценарии реализованы
- [ ] Все сценарии из §Критические пути покрыты
- [ ] Все сценарии из §Негативные сценарии покрыты
- [ ] Coverage: statements / functions / lines ≥ 80%, branches ≥ 70%

**SPEC.md:**
- [ ] Все бизнес-сущности из §Сущности присутствуют в коде
- [ ] Все бизнес-правила из §Ключевые операции реализованы и соблюдены
- [ ] Все страницы из §Страницы реализованы

**Прототип:**
- [ ] Ключевые UX-потоки прототипа воспроизводятся в готовом продукте
- [ ] Формы, действия и обратная связь пользователю соответствуют прототипу

---

### Шаг 7. Финальный коммит

```bash
git add .
git commit -m "feat(<ИмяПроекта>): полная реализация — все тесты зелёные"
```

---

## Ключевые ограничения

- **Не дублировать стартовый шаблон:** модель `Profile`, операции `me`/`updateProfile`/`profileUpdated`, слайсы `features/auth`, `widgets/Header`, `pages/404`, `shared/api/graphql-client` — уже готовы, не трогать
- **Только типизированный код:** TypeScript без `any`; GraphQL-типы — только из `npm run gen`
- **Soft delete:** при `deletedAt DateTime?` → фильтровать `{ deletedAt: null }` во всех Prisma-запросах
- **Пагинация:** все списочные операции возвращают `{ items: [...], total: Int, hasMore: Boolean }`
- **Тесты обязательны** для каждого сценария из TESTING_PLAN.md — не пропускать, не заглушать
- **Читать AGENTS.md** обоих проектов — там могут быть специфичные команды и соглашения
- **Порядок имплементации:** бэкенд → фронтенд (генерация GraphQL-типов требует работающего бэкенда)
- **Верификация блокирующая:** `npm run check && npm run test:all` должны быть зелёными перед финальным коммитом
