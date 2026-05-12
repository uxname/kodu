---
name: litefront-prototype
description: >
  Создание интерактивных прототипов на базе LiteFront (Vite, React 19, TanStack Router,
  URQL, FSD, Tailwind v4 + DaisyUI v5). Приоритет: скорость прототипирования
  и визуальное качество через существующую инфраструктуру проекта.
---

# LiteFront Prototype Skill

> **ПРОТОТИП — всё симулируется.** Нет реального бэкенда, нет реального API, нет реальной авторизации, нет реальной загрузки файлов. Цель — интерактивный UI, который работает офлайн без внешних зависимостей.

## Правила прототипа (строго соблюдать)

| Что | Правило |
|-----|---------|
| **Авторизация** | `VITE_MOCK_AUTH=true` → `MockAuthProvider`. **Не подключать** Logto / любой OIDC-сервер |
| **GraphQL API** | Только `createMockExchange` — никаких реальных HTTP-запросов к бэкенду |
| **Загрузка файлов** | Симулировать через `mockUpload` — никакого реального upload-сервера |
| **Внешние сервисы** | Любые внешние вызовы — только через моки в памяти |
| **Генерация типов** | `npm run gen` не нужен — типы писать вручную |
| **Переменные окружения** | `VITE_GRAPHQL_API_URL` и другие API-переменные игнорируются при мокинге |

---

## Директория и инициализация

Все файлы создавать внутри папки **`prototype/`**.

Если `prototype/` пуста или не существует — инициализируй через LiteFront:

```bash
npx degit uxname/litefront prototype
cd prototype
npm install
cp .env.example .env
```

Сразу после инициализации добавить в `.env`:
```
VITE_MOCK_AUTH=true
```

**После инициализации (обязательно):**
1. **Прочитай `prototype/AGENTS.md`** — главный источник конвенций проекта: структура файлов, архитектура, команды, паттерны кода. Без этого нельзя правильно генерировать код.
2. Изучи `src/` — какие слои FSD уже есть, какие компоненты существуют.
3. Проверь `package.json` — список доступных скриптов и зависимостей.
4. Убедись, что в `.env` установлено `VITE_MOCK_AUTH=true`.

## Технологический стек

| Слой | Технология |
|------|-----------|
| **Билд** | Vite 8 + React 19 + TypeScript 6 (strict) |
| **Роутинг** | TanStack Router v1 (file-based, авто-код-сплиттинг) |
| **Данные** | URQL + `createMockExchange` (никакого реального GraphQL API) |
| **Стейт** | Zustand 5 (локальный), URQL с моками (серверный) |
| **Стили** | Tailwind v4 + DaisyUI v5 (themes: cmyk / dark) |
| **Аутентификация** | `MockAuthProvider` через `VITE_MOCK_AUTH=true` (без реального OIDC) |
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

**Auth guard на роут (симулированный):**
```tsx
// src/routes/protected/index.tsx
import { createFileRoute, redirect } from '@tanstack/react-router';

export const Route = createFileRoute('/protected/')({
  beforeLoad: ({ context }) => {
    if (!context.auth.isAuthenticated) {
      throw redirect({ to: '/' });
    }
  },
});
```

С `VITE_MOCK_AUTH=true` `context.auth.isAuthenticated` всегда `true` — редирект не произойдёт.

**Навигация в коде:** `useNavigate()` из `@tanstack/react-router`.
**Запрещено** использовать `<a href>`.

**После создания нового файла роута — запусти `npm run gen:routes`**
для перегенерации `src/generated/routeTree.gen.ts`.

## 3. Авторизация (только симуляция)

`VITE_MOCK_AUTH=true` в `.env` активирует `MockAuthProvider` из `@features/auth`.
**Не подключать** реальный OIDC-сервер (Logto или любой другой) — прототип работает без него.

```tsx
// Мок-пользователь доступен сразу без логина:
const auth = useAuth();
// auth.isAuthenticated === true
// auth.user?.profile?.email === 'mock@example.com'
```

`AuthGuard` и защищённые роуты работают через мок — дополнительной настройки не нужно.

## 4. Данные (только моки, без GraphQL API)

В прототипе **нет реального GraphQL API**. Все данные хранятся в памяти и возвращаются через `createMockExchange`.
Никаких HTTP-запросов к бэкенду. `npm run gen` не нужен.

### Mock exchange

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

        return pipe(fromValue(operation), forward);
      }),
    );
};
```

### Подключение в GraphQL-клиент

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

### Мок-данные в `__root.tsx`

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

### Пример компонента с мок-данными

```tsx
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
          </tr>
        </thead>
        <tbody>
          {data.projects.map((project) => (
            <tr key={project.id} className="hover">
              <td>{project.name}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

### Правила мок-данных

1. Все данные — в памяти, никаких внешних сервисов.
2. UI должен выглядеть как с реальными данными: таблицы непустые, навигация работает, тосты показываются.
3. Для списков — минимум 2–3 элемента (не пусто, не единичная запись).
4. Состояния loading/empty/error должны быть визуально видны при соответствующих условиях.

### Типы без codegen

`npm run gen` не нужен. Типы писать вручную:

```ts
// src/shared/api/mock-types.ts
export interface Project {
  id: string;
  name: string;
  status: 'active' | 'draft' | 'archived';
}
```

Либо использовать локальную схему в `codegen.yml`:
```yaml
schema: "./src/generated/schema.graphql"
documents: "./src/graphql/**/*.graphql"
```

## 5. Загрузка файлов (только симуляция)

В прототипе файлы **не загружаются на сервер**. Симулировать через `mockUpload`:

```ts
// src/shared/api/mock-upload.ts
export async function mockUpload(file: File): Promise<{ url: string; name: string }> {
  await new Promise((r) => setTimeout(r, 800)); // имитация задержки
  return {
    url: URL.createObjectURL(file), // временный локальный URL для превью
    name: file.name,
  };
}
```

```tsx
import { mockUpload } from '@shared/api/mock-upload';
import { toast } from '@shared/ui/Toaster';

function FileUploader() {
  const [preview, setPreview] = useState<string | null>(null);

  const handleFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const result = await mockUpload(file);
    setPreview(result.url);
    toast.success(`Файл загружен: ${result.name}`);
  };

  return (
    <div>
      <input type="file" onChange={handleFile} className="file-input file-input-bordered" />
      {preview && <img src={preview} alt="preview" className="mt-2 max-h-48 rounded-lg" />}
    </div>
  );
}
```

Для файлов не-изображений — показывать имя и размер, возвращать мок-URL `/mock/uploaded/<filename>`.

## 6. Локальный стейт (Zustand)

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

## 7. Стили (Tailwind + DaisyUI)

- **DaisyUI:** `btn`, `card`, `input`, `badge`, `modal`, `table`, `table-zebra`, `loading loading-spinner`, `toggle`, `select`, `alert`, `avatar`, `tooltip`.
- **Модалки:** используй тег `<dialog>` + `useRef`. Открытие: `.showModal()`.
  Закрытие: `<form method="dialog">`.
- **Иконки:** `lucide-react`.
- **Темы:** `cmyk` (светлая, по умолчанию), `dark` (тёмная, prefers-color-scheme).
  DaisyUI-темы подключаются через атрибут `data-theme` на `<html>`.
- **Цвета DaisyUI:** `bg-base-100`, `text-base-content`, `text-primary`,
  `border-base-300`, `bg-primary`, `text-primary-content` и т.д.
  Не используй хардкодные цвета — только семантические токены темы.

## 8. i18n (ParaglideJS)

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
```

Для прототипа i18n можно временно пропустить — используй прямые строки.

## 9. Порядок действий при генерации кода

1. **Прочитай `prototype/AGENTS.md`** — конвенции проекта, структура, паттерны.
2. Изучи структуру `src/` — какие слайсы FSD уже есть.
3. Определи, в какой слой FSD ложится новая фича.
4. **Настрой моки данных** — все операции через `createMockExchange` (никаких реальных API-вызовов).
5. **Авторизация — только через `VITE_MOCK_AUTH=true`**, не подключать Logto.
6. **Загрузка файлов — только через `mockUpload`**, без реального сервера.
7. Добавь мок-данные для новых операций в `mockData` в `__root.tsx`.
8. Используй существующие компоненты где возможно.
9. Следуй паттернам из существующих страниц.

## 10. Команды

```bash
npm run start:dev    # Dev-сервер (localhost:3000)
npm run gen:routes   # Генерация routeTree.gen.ts (нужен после нового роута)
npm run check        # Линтинг + typecheck + Knip
npm run lint:fix     # Biome с автофиксом
npm run build        # Production-сборка
npm run test:dev     # Тесты (Vitest, watch)
npm run test:e2e:dev # Playwright UI mode
```

`npm run gen` (GraphQL codegen) в прототипе **не нужен** — типы пишутся вручную.

## 11. Формат ответа

Отвечай **только** блоками кода. Перед каждым блоком указывай путь
(относительно `prototype/`). Выводи файлы целиком.

```
**`src/pages/projects/ui/index.tsx`**
\`\`\`tsx
// полный код файла
\`\`\`
```
