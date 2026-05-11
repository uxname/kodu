---
name: litefront-prototype
description: >
  Создание интерактивных прототипов на базе LiteFront (Vite, React 19, TanStack Router,
  URQL, OIDC, FSD, Tailwind v4 + DaisyUI v5). Приоритет: скорость прототипирования
  и визуальное качество через существующую инфраструктуру проекта.
---

# LiteFront Prototype Skill

Твоя задача — генерировать код интерактивного MVP внутри существующего
LiteFront-проекта. Весь функционал работает против GraphQL API (LiteEnd-бэкенд
или заглушки).

## Директория и инициализация

Все файлы создавать внутри папки **`prototype/`**.

Если `prototype/` пуста или не существует — инициализируй через LiteFront:

```bash
npx degit uxname/litefront prototype
cd prototype
npm install
cp .env.example .env
```

**После инициализации (обязательно):**
1. **Прочитай `prototype/AGENTS.md`** — это главный источник конвенций
   проекта: структура файлов, архитектура, команды, паттерны кода.
   Без этого нельзя правильно генерировать код.
2. Изучи `src/` — какие слои FSD уже есть, какие компоненты существуют.
3. Проверь `package.json` — список доступных скриптов и зависимостей.
4. Запусти `npm run gen` — только если есть реальный GraphQL-эндпоинт
   (иначе см. раздел «Прототипирование без бэкенда» ниже).

## Технологический стек

| Слой | Технология |
|------|-----------|
| **Билд** | Vite 8 + React 19 + TypeScript 6 (strict) |
| **Роутинг** | TanStack Router v1 (file-based, авто-код-сплиттинг) |
| **Данные** | GraphQL + URQL (codegen → `@generated/graphql`) |
| **Стейт** | Zustand 5 (локальный), URQL (серверный) |
| **Стили** | Tailwind v4 + DaisyUI v5 (themes: cmyk / dark) |
| **Аутентификация** | OIDC через `react-oidc-context` + `oidc-client-ts` |
| **i18n** | ParaglideJS (Inlang), `messages/{locale}.json` |
| **UI-кит** | DaisyUI, Lucide-React (иконки), Sonner (тосты) |
| **Архитектура** | Feature-Sliced Design (FSD) |

## Архитектура FSD (Feature-Sliced Design)

```
src/
├── app/           # Инициализация, провайдеры, ErrorBoundary
├── entities/      # Бизнес-сущности (counter, user, project…)
├── features/      # Пользовательские сценарии (auth, createProject…)
├── pages/         # Компоненты страниц
├── routes/        # Определения роутов TanStack Router
├── shared/        # Переиспользуемое: api, ui, config, lib
├── widgets/       # Композиционные блоки (Header, Sidebar…)
├── graphql/       # .graphql-файлы (queries/, mutations/, fragments/)
└── generated/     # Авто-генерация (роуты, GraphQL-типы)
```

**Правила FSD:**
- Слой может импортировать только нижележащие слои: `pages` → `widgets` → `features` → `entities` → `shared`.
- Слайс (папка внутри слоя) не импортирует другие слайсы того же слоя.
- Публичное API слайса — через `index.ts` (barrel export).

**Path aliases** (уже настроены в `tsconfig.json`):
`@shared/*`, `@entities/*`, `@features/*`, `@widgets/*`, `@pages/*`, `@generated/*`, `@public/*`

---

## 1. Существующие компоненты (уже есть в проекте)

Не создавай их заново — используй готовые:

- **`@features/auth`** — `useAuth()`, `AuthGuard`, `MockAuthProvider`
- **`@entities/counter`** — `useCounterStore`, `Counter`
- **`@widgets/Header`** — `Header`
- **`@shared/ui/ErrorFallback`** — готовая страница ошибки с категориями (AUTH, ACCESS, NETWORK, SERVER, UNKNOWN)
- **`@shared/ui/Toaster`** — `toast` из Sonner + кастомные стили
- **`@pages/404`** — страница 404

## 2. Роутинг (TanStack Router)

Файлы роутов — в `src/routes/`. Роуты разделяют логику (`index.tsx`) и ленивый компонент (`index.lazy.tsx`).

**Новый роут:**
```tsx
// src/routes/projects.tsx
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/projects')({
  head: () => ({
    meta: [
      { title: 'Projects | LiteFront' },
      { name: 'description', content: 'Manage projects' },
    ],
  }),
});
```

```tsx
// src/routes/projects.lazy.tsx
import { ProjectsPage } from '@pages/projects';
import { createLazyFileRoute } from '@tanstack/react-router';

export const Route = createLazyFileRoute('/projects')({
  component: ProjectsPage,
});
```

**Страница:**
```tsx
// src/pages/projects/ui/index.tsx
import { useAuth } from '@features/auth';
import { Link, useNavigate } from '@tanstack/react-router';
import { Plus } from 'lucide-react';

export function ProjectsPage() {
  const auth = useAuth();
  const navigate = useNavigate();

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Projects</h1>
        <button onClick={() => navigate({ to: '/projects/new' })}
                className="btn btn-primary">
          <Plus className="size-4" /> New
        </button>
      </div>
      {auth.isAuthenticated && (
        <p className="text-sm text-base-content/60 mb-4">
          Signed in as {auth.user?.profile?.email}
        </p>
      )}
      <Link to="/projects/$id" params={{ id: '123' }}>Project</Link>
    </div>
  );
}
```

**Auth guard на роут:**
```tsx
// src/routes/protected/index.tsx
import { createFileRoute, redirect } from '@tanstack/react-router';

export const Route = createFileRoute('/protected/')({
  beforeLoad: ({ context }) => {
    if (!context.auth.isAuthenticated) {
      throw redirect({ to: '/' });
    }
  },
  // ...
});
```

**Навигация в коде:** `useNavigate()` из `@tanstack/react-router`.
**Запрещено** использовать `<a href>`.

**После создания нового файла роута — запусти `npm run gen:routes`**
для перегенерации `src/generated/routeTree.gen.ts`.

## 3. GraphQL + данные

### Новый запрос

```graphql
# src/graphql/queries/get-projects.graphql
query GetProjects {
  projects {
    id
    name
    status
  }
}
```

После добавления — запусти `npm run gen`. Импортируй хук из `@generated/graphql`:

```tsx
import { Link } from '@tanstack/react-router';
import { Edit } from 'lucide-react';
import { useGetProjectsQuery } from '@generated/graphql';

function ProjectsList() {
  const [{ data, fetching, error }] = useGetProjectsQuery();

  if (fetching) return <span className="loading loading-spinner loading-lg" />;
  if (error) return <p className="text-error">{error.message}</p>;
  if (!data?.projects.length)
    return (
      <div className="text-center py-20 bg-base-100 rounded-xl border-2 border-dashed border-base-300">
        <p className="text-base-content/50">No projects yet</p>
      </div>
    );

  return (
    <div className="overflow-x-auto bg-base-100 rounded-xl shadow-sm border border-base-300">
      <table className="table table-zebra">
        <thead>
          <tr>
            <th>Name</th>
            <th className="text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          {data.projects.map((project) => (
            <tr key={project.id} className="hover">
              <td>{project.name}</td>
              <td className="flex justify-end gap-2">
                <Link to="/projects/$id" params={{ id: project.id }}
                      className="btn btn-sm btn-ghost">
                  <Edit className="size-4" />
                </Link>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

### Мутация

```graphql
# src/graphql/mutations/create-project.graphql
mutation CreateProject($input: CreateProjectInput!) {
  createProject(input: $input) {
    id
    name
  }
}
```

```tsx
import { useCreateProjectMutation } from '@generated/graphql';
import { toast } from '@shared/ui/Toaster';

function CreateProjectForm() {
  const [, createProject] = useCreateProjectMutation();

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = new FormData(e.currentTarget);
    const { error } = await createProject({
      input: { name: form.get('name') as string },
    });
    if (error) { toast.error(error.message); return; }
    toast.success('Project created');
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <input name="name" className="input input-bordered w-full"
             placeholder="Project name" required />
      <button type="submit" className="btn btn-primary">Create</button>
    </form>
  );
}
```

### Auth + GraphQL

Токен авторизации автоматически прокидывается в GraphQL-клиент
в `__root.tsx` через `auth.user?.id_token`. Ничего дополнительно делать
не нужно.

## 4. Локальный стейт (Zustand)

Для стейта, который не идёт через GraphQL:

```tsx
import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

interface UIState {
  sidebarOpen: boolean;
  toggleSidebar: () => void;
}

export const useUIStore = create(
  devtools<UIState>((set) => ({
    sidebarOpen: false,
    toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
  })),
);
```

## 5. Стили (Tailwind + DaisyUI)

- **DaisyUI:** `btn`, `card`, `input`, `badge`, `modal`, `table`, `table-zebra`, `loading loading-spinner`, `toggle`, `select`, `alert`, `avatar`, `tooltip`.
- **Модалки:** используй тег `<dialog>` + `useRef`. Открытие: `.showModal()`.
  Закрытие: `<form method="dialog">`.
- **Иконки:** `lucide-react`.
- **Темы:** `cmyk` (светлая, по умолчанию), `dark` (тёмная, prefers-color-scheme).
  DaisyUI-темы подключаются через атрибут `data-theme` на `<html>`.
- **Цвета DaisyUI:** `bg-base-100`, `text-base-content`, `text-primary`,
  `border-base-300`, `bg-primary`, `text-primary-content` и т.д.
  Не используй хардкодные цвета — только семантические токены темы.

## 6. i18n (ParaglideJS)

Сообщения — в `messages/{locale}.json`. Используй готовые или добавляй новые.

```json
{
  "projects_title": "Projects",
  "project_created": "Project created successfully"
}
```

Импорт в коде:
```tsx
import * as m from '@generated/paraglide/messages';
// m.projects_title() → "Projects"
// Параметры: m.project_created({ name: 'My Project' })
```

Для нового прототипа i18n можно временно пропустить — используй
прямые строки. Добавишь переводы потом.

## 7. Прототипирование без бэкенда (моки)

В прототипе **нет реального бэкенда** — все данные должны быть замоканы.
Моки должны работать из коробки, без запуска внешних сервисов.

### Auth (уже встроено)

В `.env` установи `VITE_MOCK_AUTH=true` — включится `MockAuthProvider`,
который не требует OIDC-сервера. Пользователь считается аутентифицированным.

### GraphQL — правильный mock exchange

Создай кастомный URQL exchange, который перехватывает операции
и возвращает мок-данные, **не делая запрос в сеть**:

```ts
// src/shared/api/mock-exchange.ts
import { Exchange, makeResult } from 'urql';
import { pipe, mergeMap, fromValue } from 'wonka';

type Mocks = Record<
  string,
  (variables?: Record<string, unknown>) => Record<string, unknown>
>;

export const createMockExchange = (mocks: Mocks): Exchange => {
  return ({ forward }) => (ops$) =>
    pipe(
      ops$,
      mergeMap((operation) => {
        const queryName = operation.query.definitions[0]?.name?.value;
        const handler = queryName ? mocks[queryName] : undefined;

        if (handler) {
          const data = handler(operation.variables as Record<string, unknown>);
          return fromValue(makeResult(operation, { data }));
        }

        // нет мока — пропускаем дальше по цепочке (до fetchExchange)
        return pipe(fromValue(operation), forward);
      }),
    );
};
```

**Как подключить в существующий GraphQL-клиент:**
добавь exchange в список **перед** `fetchExchange`:

```ts
// src/shared/api/graphql-client.ts
import { createMockExchange } from './mock-exchange';
import type { Mocks } from './mock-exchange';

export const createGraphQLClient = (
  accessToken?: string,
  mocks?: Mocks,
): Client => {
  const exchanges = [
    cacheExchange,
    errorExchange({ onError: (error) => { /* ... */ } }),
  ];

  if (mocks) exchanges.push(createMockExchange(mocks));

  exchanges.push(fetchExchange);

  return new Client({
    url: import.meta.env.VITE_GRAPHQL_API_URL,
    exchanges,
    fetchOptions: {
      headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : {},
    },
    requestPolicy: 'cache-and-network',
  });
};
```

**Где передать моки — в `__root.tsx`:**

```tsx
// src/routes/__root.tsx
import { GraphQLProvider, createGraphQLClient } from '@shared/api';

const mockData = {
  GetProjects: () => ({
    projects: [
      { id: 'p-1', name: 'Project Alpha', status: 'active' },
      { id: 'p-2', name: 'Project Beta', status: 'draft' },
    ],
  }),
  CreateProject: (vars) => ({
    createProject: { id: 'p-new', name: vars?.input?.name || 'New' },
  }),
};

function RootComponent() {
  const auth = useAuth();
  const client = useMemo(
    () => createGraphQLClient(auth.user?.id_token, mockData),
    [auth.user?.id_token],
  );
  // ...
}
```

**Итоговая схема для прототипа:**

```
VITE_MOCK_AUTH=true         → MockAuthProvider (уже есть)
mocks в createGraphQLClient → URQL не ходит в сеть
```

### Генерация типов без бэкенда

`npm run gen` не нужен — типы пишутся вручную, либо используй
локальную схему:

```yaml
# codegen.yml (прототип)
schema: "./src/generated/schema.graphql"
documents: "./src/graphql/**/*.graphql"
generates:
  src/generated/graphql.tsx:
    plugins:
      - "typescript"
      - "typescript-operations"
      - "typescript-urql"
```

Схему скопируй из любого GraphQL-проекта или сгенерируй,
указав временный endpoint.

### Правила мокинга

1. Все данные — в памяти, никаких внешних сервисов.
2. UI должен выглядеть как с реальными данными: таблицы непустые,
   навигация работает, тосты показываются.
3. Для списков — минимум 2-3 элемента, чтобы было видно состояние
   «с данными» (не пусто и не единичная запись).
4. Для ошибок — опционально, но loading/empty/error states должны быть
   видны при соответствующих условиях.

## 8. Порядок действий при генерации кода

1. **Прочитай `prototype/AGENTS.md`** — конвенции проекта, структура, паттерны.
2. Изучи существующую структуру `src/` — какие слайсы FSD уже есть.
3. Определи, в какой слой FSD ложится новая фича.
4. Настрой моки (см. секцию 7) — в прототипе всё должно работать офлайн.
5. Если нужны новые GraphQL-операции — создай `.graphql`-файл, запусти `npm run gen`
   (или напиши типы вручную, если бэкенда нет).
6. Добавь мок-данные для новых операций в `mockData`.
7. Импортируй сгенерированные хуки из `@generated/graphql`.
8. Используй существующие компоненты где возможно.
9. Следуй паттернам, которые уже есть в проекте (смотри на существующие страницы).

## 9. Команды

```bash
npm run start:dev   # Dev-сервер (localhost:3000)
npm run gen         # Генерация GraphQL-типов
npm run gen:routes  # Генерация routeTree.gen.ts
npm run check       # Линтинг + typecheck + Knip
npm run lint:fix    # Biome с автофиксом
npm run build       # Production-сборка
npm run test:dev    # Тесты (Vitest, watch)
npm run test:e2e:dev # Playwright UI mode
```

## 10. Формат ответа

Отвечай **только** блоками кода. Перед каждым блоком указывай путь
(относительно `prototype/`). Выводи файлы целиком.

```
**`src/graphql/queries/get-projects.graphql`**
\`\`\`graphql
query GetProjects { ... }
\`\`\`
```
