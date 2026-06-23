---
name: litefront-prototype
description: >
  Building interactive prototypes on top of LiteFront (Vite, React 19, TanStack Router,
  URQL, FSD, Tailwind v4 + DaisyUI v5). Priority: prototyping speed
  and visual quality through the project's existing infrastructure.
---

# LiteFront Prototype Skill

> **PROTOTYPE — everything is simulated.** No real backend, no real API, no real authentication, no real file uploads. The goal is an interactive UI that works offline with no external dependencies.

## Prototype rules (strictly follow)

| What | Rule |
|-----|---------|
| **Authentication** | `VITE_MOCK_AUTH=true` → `MockAuthProvider`. **Do not connect** Logto / any OIDC server |
| **GraphQL API** | Only `createMockExchange` — no real HTTP requests to a backend |
| **File uploads** | Simulate via `mockUpload` — no real upload server |
| **External services** | Any external calls — only through in-memory mocks |
| **Type generation** | `npm run gen` is not needed — write types by hand |
| **Environment variables** | `VITE_GRAPHQL_API_URL` and other API variables are ignored when mocking |

---

## Directory and initialization

Create all files inside the **`prototype/`** folder.

If `prototype/` is empty or doesn't exist — initialize it via LiteFront:

```bash
npx degit uxname/litefront prototype
cd prototype
npm install
cp .env.example .env
```

Immediately after initialization, add to `.env`:
```
VITE_MOCK_AUTH=true
```

**After initialization (mandatory):**
1. **Read `prototype/AGENTS.md`** — the primary source of the project's conventions: file structure, architecture, commands, code patterns. Without this you can't generate code correctly.
2. Study `src/` — which FSD layers already exist, which components exist.
3. Check `package.json` — the list of available scripts and dependencies.
4. Make sure `.env` has `VITE_MOCK_AUTH=true` set.

## Technology stack

| Layer | Technology |
|------|-----------|
| **Build** | Vite 8 + React 19 + TypeScript 6 (strict) |
| **Routing** | TanStack Router v1 (file-based, auto code-splitting) |
| **Data** | URQL + `createMockExchange` (no real GraphQL API) |
| **State** | Zustand 5 (local), URQL with mocks (server) |
| **Styles** | Tailwind v4 + DaisyUI v5 (themes: cmyk / dark) |
| **Authentication** | `MockAuthProvider` via `VITE_MOCK_AUTH=true` (no real OIDC) |
| **i18n** | ParaglideJS (Inlang), `messages/{locale}.json` |
| **UI kit** | DaisyUI, Lucide-React (icons), Sonner (toasts) |
| **Architecture** | Feature-Sliced Design (FSD) |

## FSD architecture (Feature-Sliced Design)

```
src/
├── app/           # Initialization, providers, ErrorBoundary
├── entities/      # Business entities (counter, user, project…)
├── features/      # User scenarios (auth, createProject…)
├── pages/         # Page components
├── routes/        # TanStack Router route definitions
├── shared/        # Reusable: api, ui, config, lib
├── widgets/       # Composite blocks (Header, Sidebar…)
├── graphql/       # .graphql files (queries/, mutations/, fragments/)
└── generated/     # Auto-generated (routes, GraphQL types)
```

**FSD rules:**
- A layer may import only the layers below it: `pages` → `widgets` → `features` → `entities` → `shared`.
- A slice (a folder within a layer) does not import other slices of the same layer.
- A slice's public API is exposed through `index.ts` (barrel export).

**Path aliases** (already configured in `tsconfig.json`):
`@shared/*`, `@entities/*`, `@features/*`, `@widgets/*`, `@pages/*`, `@generated/*`, `@public/*`

---

## 1. Existing components (already in the project)

Don't recreate them — use the ready-made ones:

- **`@features/auth`** — `useAuth()`, `AuthGuard`, `MockAuthProvider`
- **`@entities/counter`** — `useCounterStore`, `Counter`
- **`@widgets/Header`** — `Header`
- **`@shared/ui/ErrorFallback`** — a ready-made error page with categories (AUTH, ACCESS, NETWORK, SERVER, UNKNOWN)
- **`@shared/ui/Toaster`** — `toast` from Sonner + custom styles
- **`@pages/404`** — a 404 page

## 2. Routing (TanStack Router)

Route files live in `src/routes/`. Routes separate the logic (`index.tsx`) from the lazy component (`index.lazy.tsx`).

**New route:**
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

**Page:**
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

**Auth guard on a route (simulated):**
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

With `VITE_MOCK_AUTH=true`, `context.auth.isAuthenticated` is always `true` — the redirect won't happen.

**Navigation in code:** `useNavigate()` from `@tanstack/react-router`.
**It is forbidden** to use `<a href>`.

**After creating a new route file — run `npm run gen:routes`**
to regenerate `src/generated/routeTree.gen.ts`.

## 3. Authentication (simulation only)

`VITE_MOCK_AUTH=true` in `.env` activates `MockAuthProvider` from `@features/auth`.
**Do not connect** a real OIDC server (Logto or any other) — the prototype works without it.

```tsx
// The mock user is available immediately, no login needed:
const auth = useAuth();
// auth.isAuthenticated === true
// auth.user?.profile?.email === 'mock@example.com'
```

`AuthGuard` and protected routes work through the mock — no extra setup needed.

## 4. Data (mocks only, no GraphQL API)

The prototype has **no real GraphQL API**. All data is kept in memory and returned via `createMockExchange`.
No HTTP requests to a backend. `npm run gen` is not needed.

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

### Wiring into the GraphQL client

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

### Mock data in `__root.tsx`

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

### Example component with mock data

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

### Mock data rules

1. All data is in memory, no external services.
2. The UI should look like it has real data: tables are non-empty, navigation works, toasts appear.
3. For lists — at least 2–3 items (not empty, not a single record).
4. The loading/empty/error states must be visually visible under the corresponding conditions.

### Types without codegen

`npm run gen` is not needed. Write the types by hand:

```ts
// src/shared/api/mock-types.ts
export interface Project {
  id: string;
  name: string;
  status: 'active' | 'draft' | 'archived';
}
```

Or use a local schema in `codegen.yml`:
```yaml
schema: "./src/generated/schema.graphql"
documents: "./src/graphql/**/*.graphql"
```

## 5. File uploads (simulation only)

In the prototype, files are **not uploaded to a server**. Simulate via `mockUpload`:

```ts
// src/shared/api/mock-upload.ts
export async function mockUpload(file: File): Promise<{ url: string; name: string }> {
  await new Promise((r) => setTimeout(r, 800)); // simulate a delay
  return {
    url: URL.createObjectURL(file), // temporary local URL for the preview
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
    toast.success(`File uploaded: ${result.name}`);
  };

  return (
    <div>
      <input type="file" onChange={handleFile} className="file-input file-input-bordered" />
      {preview && <img src={preview} alt="preview" className="mt-2 max-h-48 rounded-lg" />}
    </div>
  );
}
```

For non-image files — show the name and size, return a mock URL `/mock/uploaded/<filename>`.

## 6. Local state (Zustand)

For state that doesn't go through GraphQL:

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

## 7. Styles (Tailwind + DaisyUI)

- **DaisyUI:** `btn`, `card`, `input`, `badge`, `modal`, `table`, `table-zebra`, `loading loading-spinner`, `toggle`, `select`, `alert`, `avatar`, `tooltip`.
- **Modals:** use the `<dialog>` tag + `useRef`. Open: `.showModal()`.
  Close: `<form method="dialog">`.
- **Icons:** `lucide-react`.
- **Themes:** `cmyk` (light, default), `dark` (dark, prefers-color-scheme).
  DaisyUI themes are applied via the `data-theme` attribute on `<html>`.
- **DaisyUI colors:** `bg-base-100`, `text-base-content`, `text-primary`,
  `border-base-300`, `bg-primary`, `text-primary-content`, etc.
  Don't use hardcoded colors — only the theme's semantic tokens.

## 8. i18n (ParaglideJS)

Messages live in `messages/{locale}.json`. Use the existing ones or add new ones.

```json
{
  "projects_title": "Projects",
  "project_created": "Project created successfully"
}
```

Import in code:
```tsx
import * as m from '@generated/paraglide/messages';
// m.projects_title() → "Projects"
```

For a prototype, i18n can be skipped for now — use direct strings.

## 9. Steps when generating code

1. **Read `prototype/AGENTS.md`** — the project's conventions, structure, patterns.
2. Study the `src/` structure — which FSD slices already exist.
3. Determine which FSD layer the new feature belongs to.
4. **Set up the data mocks** — all operations through `createMockExchange` (no real API calls).
5. **Authentication — only via `VITE_MOCK_AUTH=true`**, don't connect Logto.
6. **File uploads — only via `mockUpload`**, no real server.
7. Add mock data for the new operations to `mockData` in `__root.tsx`.
8. Use the existing components where possible.
9. Follow the patterns from the existing pages.

## 10. Commands

```bash
npm run start:dev    # Dev server (localhost:3000)
npm run gen:routes   # Generate routeTree.gen.ts (needed after a new route)
npm run check        # Linting + typecheck + Knip
npm run lint:fix     # Biome with auto-fix
npm run build        # Production build
npm run test:dev     # Tests (Vitest, watch)
npm run test:e2e:dev # Playwright UI mode
```

`npm run gen` (GraphQL codegen) is **not needed** in the prototype — types are written by hand.

## 11. Response format

Respond with **only** code blocks. Before each block, give the path
(relative to `prototype/`). Output files in full.

```
**`src/pages/projects/ui/index.tsx`**
\`\`\`tsx
// full file code
\`\`\`
```
