---
name: implement-project
description: Implements a complete project from ready documentation, specs, and a prototype. Initializes the backend (liteend-init) and frontend (litefront-init), implements all entities/operations/pages, covers them with tests, and verifies conformance to the documents. Run when VISION.md + SPEC.md are ready and the tech-blueprint is approved. Do NOT run if the spec is still a draft or the project is already partially implemented.
license: MIT
compatibility: opencode
metadata:
  level: multi
  output: folder projects/<ProjectName>/
---

## Purpose

The skill takes the ready design artifacts and **implements the project from zero to a working, tested state**:
- Initializes the backend and frontend from the starter templates
- Implements all the business logic strictly per the spec
- Covers it with tests per TESTING_PLAN.md
- Verifies conformance to the documentation and runs the full suite of checks

**When to run:**
- `docs/<name>/` and `blueprint/<name>/3_TECH_BLUEPRINT/` are ready and approved
- The spec passed `blueprint_validator.py` validation without errors
- The team is ready to develop

**When NOT to run:**
- The spec is a draft or hasn't passed validation
- The project is already partially implemented (this is not a refactoring skill)
- The task is to add a single feature, not to build a project from scratch

---

## Inputs

Before starting, you **must read** all the documents:

```
docs/<ProjectName>/
├── 1_PRODUCT_VISION/VISION.md           ← business goals, scope, roles
└── 2_PRODUCT_SPEC/SPEC.md               ← entities, operations, pages, business rules

blueprint/<ProjectName>/3_TECH_BLUEPRINT/
├── IMPLEMENTATION_GUIDE.md              ← stack, what's already done, run commands
├── DATABASE_MODEL.md                    ← Prisma schema of all models
├── API_CONTRACTS.md                     ← GraphQL schema, guards, pagination
├── ARCHITECTURE.md                      ← NestJS modules, FSD slices, state
└── TESTING_PLAN.md                      ← unit tests, E2E scenarios, coverage

prototype/                               ← UI prototype (only for understanding UX flows)
```

If any of the first six documents is missing — **stop** and tell the user which file is missing. The prototype is optional.

---

## Output structure

```
projects/<ProjectName>/
├── backend/     ← NestJS + Fastify API (initialized via liteend-init)
└── frontend/    ← React SPA (initialized via litefront-init)
```

Create the root folder before initialization:
```bash
mkdir -p projects/<ProjectName>
```

---

## Implementation process

### Step 0. Documentation analysis

1. Read all documents in full
2. Build an internal working list:
   - Prisma models from DATABASE_MODEL.md (without Profile/ProfileRole — they're already in the template)
   - GraphQL operations by domain from API_CONTRACTS.md (without `me`, `updateProfile`, `profileUpdated`, `debug`, `echo`)
   - NestJS modules from ARCHITECTURE.md §Backend
   - FSD slices to implement from ARCHITECTURE.md §Frontend (without `features/auth`, `widgets/Header`, `pages/404`, `shared/api/graphql-client`)
   - Test scenarios from TESTING_PLAN.md
3. Note: everything from the starter templates is **not to be touched or duplicated**

---

### Step 1. Backend initialization

Run the **liteend-init** skill (`~/.config/opencode/skills/liteend-init/SKILL.md`):

```
project_name: projects/<ProjectName>/backend
use_docker:   true
install_deps: true
```

After it finishes:
- Read `projects/<ProjectName>/backend/AGENTS.md` — understand the project's available commands
- Make sure `GET /health` responds with 200 (the backend started)

---

### Step 2. Backend implementation

**2.1. Database schema**

Open `backend/prisma/schema.prisma` and **add** the new models from DATABASE_MODEL.md to the existing ones:
- Existing models `Profile`, `ProfileRole` — do not change
- To link new entities to the user: `profileId Int` + `@relation(fields: [profileId], references: [id], onDelete: Cascade)`
- All relations — with an explicit `onDelete`; all enumerable values — only via `enum`

Apply the migration:
```bash
# inside backend/
npm run db:migrations:apply
```

**2.2. NestJS modules**

For each module from ARCHITECTURE.md §Backend, create the structure:
```
src/modules/<domain>/
├── <domain>.module.ts       ← @Module({ imports, providers, exports })
├── <domain>.resolver.ts     ← @Resolver() with the GraphQL operations
├── <domain>.service.ts      ← business logic, Prisma queries
└── dto/
    └── <entity>.input.ts    ← ZodDto or class-validator Input for mutations
```

Resolver implementation rules:
- Guards correspond to the access directives from API_CONTRACTS.md:
  ```typescript
  @UseGuards(JwtAuthGuard)         // # @auth in API_CONTRACTS.md
  @UseGuards(JwtOptionalAuthGuard) // # @auth? in API_CONTRACTS.md
  @Roles(ProfileRole.ADMIN)        // # @auth @hasRole(ADMIN)
  // no guard                      // # @public
  ```
- Current user: `@CurrentUser() profile: Profile`
- Errors: throw `HttpException` or `ZodValidationException` — `gqlErrorFormatter` converts them itself
- Soft delete: when the model has `deletedAt DateTime?` → filter `{ deletedAt: null }` in **all** Prisma queries
- Pagination: all list operations accept `limit`/`offset` and return `{ items, total, hasMore }`

**2.3. Subscriptions**

If API_CONTRACTS.md has Subscription operations — implement them via Redis pub/sub (already configured). Don't add new transport dependencies.

**2.4. Backend commit**
```bash
git add .
git commit -m "feat(backend): implement business logic for <ProjectName>"
```

---

### Step 3. Frontend initialization

Run the **litefront-init** skill (`~/.config/opencode/skills/litefront-init/SKILL.md`):

```
project_name: projects/<ProjectName>/frontend
install_deps: true
setup_env:    true
run_dev:      false
```

Configure the frontend `.env`:
```
VITE_GRAPHQL_API_URL=http://localhost:<PORT>/graphql
VITE_BASE_URL=http://localhost:<FRONT_PORT>
VITE_OIDC_AUTHORITY=<from IMPLEMENTATION_GUIDE.md>
VITE_OIDC_CLIENT_ID=<from IMPLEMENTATION_GUIDE.md>
```

Start the backend in the background and generate the GraphQL types:
```bash
# in backend/:
npm run start:dev &

# in frontend/:
npm run gen
```

Read `projects/<ProjectName>/frontend/AGENTS.md` — understand the available commands.

**3.5. The project's root AGENTS.md**

Create `projects/<ProjectName>/AGENTS.md`:

```markdown
# <ProjectName>

## Project structure

- `backend/` — NestJS + Fastify API
- `frontend/` — React SPA

## Important

When working with the backend, **you must read** `backend/AGENTS.md`.
When working with the frontend, **you must read** `frontend/AGENTS.md`.

Each subproject contains specific commands, conventions, and environment quirks
that you need to know before making changes.
```

---

### Step 4. Frontend implementation

**4.1. FSD slices (new ones only)**

For each slice from ARCHITECTURE.md §Frontend, create the structure:
```
src/
├── entities/<name>/
│   ├── api/        ← URQL useQuery / useMutation with types from npm run gen
│   └── model/      ← TypeScript types, data transformations
├── features/<name>/
│   ├── ui/         ← React components
│   └── model/      ← Zustand store or local state
└── widgets/<name>/
    └── ui/
```

Rules:
- GraphQL requests: only via URQL, types from `npm run gen` — no `any`
- Zustand store: `create(devtools<MyStore>((set, get) => ({ ... })))` — only UI state, don't duplicate server data
- Components: take only the UX logic from the prototype (flows, forms), don't copy the code
- **Theme and style:** take them from the prototype as the baseline — color scheme, typography, spacing, visual language. The real components should look recognizable compared to the prototype
- **No mocked data in production code:** all data is fetched via GraphQL requests to the backend. Hardcoded data and `mockData` constants are allowed only in tests (`*.test.ts`, `*.spec.ts`, `tests/`)

**4.2. Pages and routing**

For each page from ARCHITECTURE.md §Pages, create a file in `src/routes/`:

```typescript
// Protected page (beforeLoad pattern):
export const Route = createFileRoute('/my-page')({
  beforeLoad: ({ context: { auth } }) => {
    if (!auth.isAuthenticated) throw redirect({ to: '/' })
  },
  component: () => <AuthGuard><MyPage /></AuthGuard>,
})

// Public page:
export const Route = createFileRoute('/public-page')({
  component: MyPublicPage,
})
```

**4.3. Environment variables**

Add the variables from ARCHITECTURE.md §Environment variables to `.env` (only the business-specific ones).

**4.4. i18n**

If ARCHITECTURE.md §Frontend: localization contains namespaces — create the translation files in `messages/`.

**4.5. Frontend commit**
```bash
git add .
git commit -m "feat(frontend): implement <ProjectName>"
```

---

### Step 5. Tests

Implement all scenarios from TESTING_PLAN.md. **Don't skip any scenarios** — each one is described in the document.

**Backend E2E (Vitest + E2EClient):**
```typescript
describe('<Domain>Resolver', () => {
  let client: E2EClient

  beforeEach(async () => {
    await clearDatabase()
    await clearRedis()
    client = await createTestClient()
  })

  it('success scenario', async () => {
    const profile = await createTestProfile()
    await client.loginAs(profile)
    const result = await client.requestGraphQL<QueryType>(QUERY, vars)
    expect(result.data).toBeDefined()
  })

  it('negative scenario — no permissions', async () => {
    const result = await client.requestGraphQL<QueryType>(QUERY, vars)
    expect(result.errors?.[0].message).toContain('Unauthorized')
  })
})
```

**Backend unit (Vitest + NestJS Testing Module):**
```typescript
describe('<Domain>Service.complexMethod', () => {
  it('correctly calculates ...', () => {
    const result = service.complexMethod(input)
    expect(result).toEqual(expected)
  })
})
```

**Frontend component (Vitest + React Testing Library):**
```typescript
describe('<ComponentName>', () => {
  it('renders the loading state', () => {
    render(<MyComponent loading />)
    expect(screen.getByRole('progressbar')).toBeInTheDocument()
  })
})
// OIDC is mocked automatically via tests/setup.ts — no extra setup needed
```

**Frontend E2E (Playwright):**
```typescript
// Run with VITE_MOCK_AUTH=true — MockAuthProvider authenticates automatically
test('critical path: <name>', async ({ page }) => {
  await page.goto('/target-page')
  await page.getByRole('button', { name: 'Action' }).click()
  await expect(page.getByText('Success')).toBeVisible()
})
```

**Test commit:**
```bash
git add .
git commit -m "test(<ProjectName>): coverage per TESTING_PLAN.md"
```

---

### Step 6. Verification (mandatory — do not skip)

**6.1. Automated checks**

Run in both projects:
```bash
# Backend:
cd projects/<ProjectName>/backend && npm run check && npm run test:all

# Frontend:
cd projects/<ProjectName>/frontend && npm run check && npm run test:all
```

**If even one command finishes with an error — fix the cause and run it again.** The project is not considered implemented until all checks are green.

**6.2. Manual checklist against the documents**

Verify the implementation against each document:

**DATABASE_MODEL.md:**
- [ ] Every Prisma model from the document exists in `schema.prisma`
- [ ] All migrations are applied — `prisma migrate status` shows no pending
- [ ] The `@@index` indexes and `@@unique` constraints match the document
- [ ] The `onDelete`/`onUpdate` cascade operations match the document

**API_CONTRACTS.md:**
- [ ] Every Query/Mutation/Subscription is implemented in the corresponding resolver
- [ ] Guards match the directives (`# @auth` → `JwtAuthGuard`, `# @public` → no guard)
- [ ] Return types match (including the pagination wrapper types)
- [ ] Mutation input types match the schema

**ARCHITECTURE.md:**
- [ ] All NestJS modules from §Backend are implemented and registered in AppModule
- [ ] All FSD slices from §Frontend are implemented (except the boilerplate slices)
- [ ] All pages from §Pages exist and route correctly
- [ ] Server state is managed via URQL, UI state — via Zustand

**TESTING_PLAN.md:**
- [ ] All scenarios from §Unit tests are implemented
- [ ] All scenarios from §E2E scenarios are implemented
- [ ] All scenarios from §Critical paths are covered
- [ ] All scenarios from §Negative scenarios are covered
- [ ] Coverage: statements / functions / lines ≥ 80%, branches ≥ 70%

**SPEC.md:**
- [ ] All business entities from §Entities are present in the code
- [ ] All business rules from §Key operations are implemented and enforced
- [ ] All pages from §Pages are implemented

**Prototype:**
- [ ] The prototype's key UX flows are reproduced in the finished product
- [ ] Forms, actions, and user feedback match the prototype
- [ ] The visual style (colors, typography, layout) matches the prototype
- [ ] Production code has no mocked data, hardcoded constants, or `TODO: replace with API`

---

### Step 7. Final commit

```bash
git add .
git commit -m "feat(<ProjectName>): full implementation — all tests green"
```

---

## Key constraints

- **Don't duplicate the starter template:** the `Profile` model, the `me`/`updateProfile`/`profileUpdated` operations, the `features/auth`, `widgets/Header`, `pages/404`, `shared/api/graphql-client` slices — are already done, don't touch them
- **Typed code only:** TypeScript with no `any`; GraphQL types — only from `npm run gen`
- **Soft delete:** when `deletedAt DateTime?` → filter `{ deletedAt: null }` in all Prisma queries
- **Pagination:** all list operations return `{ items: [...], total: Int, hasMore: Boolean }`
- **Tests are mandatory** for every scenario from TESTING_PLAN.md — don't skip, don't stub them out
- **Read the AGENTS.md** of both projects — they may contain specific commands and conventions
- **Implementation order:** backend → frontend (generating the GraphQL types requires a running backend)
- **No mocked data in production:** `mockData`, hardcoded arrays, `TODO: replace with API` — forbidden in production code; allowed only in test files
- **The theme comes from the prototype:** the real UI must reproduce the prototype's visual language — colors, typography, layout
- **Verification is blocking:** `npm run check && npm run test:all` must be green before the final commit
