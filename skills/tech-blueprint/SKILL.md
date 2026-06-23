---
name: tech-blueprint
description: Generates technical contracts (Prisma schema, GraphQL API, architecture, testing plan) based on SPEC.md and a UI prototype. Also reviews an already-created blueprint on request. Run when SPEC.md is approved and it's time to move to technical design, or when the user says "check the spec", "validate the blueprint". Do NOT run if SPEC.md is still a draft or the task is refactoring existing code.
license: MIT
compatibility: opencode
metadata:
  level: multi
  output: folder 3_TECH_BLUEPRINT/
---

## Purpose

The skill creates five documents that fully describe the product's technical implementation:
- **IMPLEMENTATION_GUIDE.md** — developer onboarding: stack, what's already done, how to run it
- **DATABASE_MODEL.md** — Prisma schema, relations, indexes
- **API_CONTRACTS.md** — GraphQL schema with types, operations, and access rules
- **ARCHITECTURE.md** — NestJS modules, FSD slices, state management
- **TESTING_PLAN.md** — unit tests and Playwright E2E scenarios

**When to run:**
- SPEC.md has the status "in review" or "approved"
- A UI prototype exists in the `prototype/` folder
- The team is ready to start development

**When NOT to run:**
- SPEC.md is still a draft
- There's no understanding of the business entities
- The task is refactoring or fixing a bug in an existing project

---

## Inputs

Before generation, you **must read and analyze**:

1. **`1_PRODUCT_VISION/VISION.md`** — goal, scope, success metrics
2. **`2_PRODUCT_SPEC/SPEC.md`** — entities, operations, pages, testing (**the main source of truth**)
3. **`prototype/`** — only to understand the scope: which pages exist, the flows between them, the forms. **Do not copy technical decisions from the prototype.** The prototype was built quickly and without regard for best practices — it shows "what", not "how". Everything architectural follows the stack's best practices, regardless of how it's implemented in the prototype.
4. **Existing schemas** — if the project already has a `schema.prisma` or `schema.graphql`: read it before generation, this means update mode

If there isn't enough input — ask **no more than 2 clarifying questions**.

Before generation, build an internal list:
- business entities (from SPEC.md §Entities)
- key operations (from SPEC.md §Key operations)
- pages (from SPEC.md §Pages, the prototype — to cross-check the scope)
- user roles (from SPEC.md §Glossary or §Entities)

---

## Starter templates: what's already implemented

Projects are built on top of two pre-built starter templates (boilerplate — a ready codebase with the basic tasks already solved): a backend template (NestJS) and a frontend template (React). They already implement authentication, user handling, and the test infrastructure. **Don't design them again** — only extend them.

### Backend

**The Prisma model `Profile` (already exists, do not duplicate):**
```prisma
model Profile {
  id        Int           @id @default(autoincrement())
  createdAt DateTime      @default(now())
  updatedAt DateTime      @updatedAt
  oidcSub   String        @unique   // link to the OIDC provider
  roles     ProfileRole[] @default([USER])
  avatarUrl String?
}

enum ProfileRole { ADMIN  USER }
```

**GraphQL operations (already implemented, do not include in the Blueprint):**
| Operation | Type | Access |
|---------|-----|--------|
| `me` | Query | JwtAuthGuard |
| `updateProfile(input)` | Mutation | JwtAuthGuard |
| `profileUpdated` | Subscription | by userId |
| `debug` | Query | JwtOptionalAuthGuard |
| `echo` | Query + Mutation | public |

**Infrastructure (configured, just use it):**
- Auth guards: `JwtAuthGuard`, `JwtOptionalAuthGuard`, `RolesGuard`, `@CurrentUser()`, `@Roles()`
- GraphQL: **MercuriusDriver** (Fastify) — not Apollo Server
- Subscriptions: WebSocket + Redis pub/sub (mqemitter-redis)
- Error handling: `gqlErrorFormatter` — maps `ZodValidationException` and `HttpException` to GraphQL errors
- 404 backend: `@All()` → `NotFoundException`
- Health check: `GET /health`
- Queue: BullMQ + Redis
- i18n: nestjs-i18n (en/ru)

**Backend test infrastructure:**
- `E2EClient.loginAs(profile)` — auth via the `x-mock-sub` header
- `clearDatabase()` + `clearRedis()` — mandatory in `beforeEach()`
- `OIDC_MOCK_ENABLED=true` — mock-auth mode in `.env.test`

### Frontend

**FSD slices** (FSD — Feature-Sliced Design, a methodology for structuring frontend code by business layers) **already implemented, do not include in the Blueprint:**
| Layer | Slice | What it contains |
|------|-------|-------------|
| `features` | `auth` | OIDC PKCE configuration, `AuthGuard`, `MockAuthProvider` |
| `widgets` | `Header` | Navigation + auth buttons (auth state) |
| `pages` | `404` | Fully implemented with "Back" and "Home" buttons |
| `shared/api` | `graphql-client` | URQL client factory with a Bearer token |

**Routes (already implemented):**
| Route | Purpose |
|------|-----------|
| `/callback` | OIDC redirect handler (after login) |
| `*` (any non-existent) | 404 page via `defaultNotFoundComponent` |

**Frontend test infrastructure:**
- `VITE_MOCK_AUTH=true` — `MockAuthProvider` for Playwright e2e
- `tests/setup.ts` — a global mock of `react-oidc-context` for Vitest unit/component tests
- Coverage: lines / functions / statements ≥ 80%, branches ≥ 70%

---

## Output structure

The technical contracts are stored **separately** from the product documentation (doc-gen), so as not to mix business artifacts with technical ones:

```
docs/<project_name>/          ← product documentation (doc-gen)
├── INDEX.md
├── 1_PRODUCT_VISION/
└── 2_PRODUCT_SPEC/

blueprint/<project_name>/     ← technical contracts (this skill)
└── 3_TECH_BLUEPRINT/
    ├── IMPLEMENTATION_GUIDE.md   ← developer onboarding (generate first)
    ├── DATABASE_MODEL.md
    ├── API_CONTRACTS.md
    ├── ARCHITECTURE.md
    └── TESTING_PLAN.md
```

Create the folder before generation:
```bash
mkdir -p blueprint/<ProjectName>/3_TECH_BLUEPRINT
```

---

## Generation rules: DATABASE_MODEL.md

### File structure

```markdown
# Data model: <Product name>

**Status:** draft | **Date:** YYYY-MM-DD

## Links
- Specification: [SPEC.md](../2_PRODUCT_SPEC/SPEC.md)

## Implementation context
- **ORM:** Prisma (TypeScript ORM) — the schema is in `prisma/schema.prisma`, migrations via `prisma migrate dev`
- **DB:** PostgreSQL
- **`Profile` — the current user's model** (already exists in the starter template, do not duplicate): contains `id`, `oidcSub @unique` (the identifier from the OIDC provider), `roles`, `avatarUrl`. To link a business entity to the user: `profileId Int` + `@relation(fields: [profileId], references: [id])`.
- **Authentication via OIDC** (OpenID Connect — a sign-in protocol via an external service, for example Logto or Keycloak): the `User`, `Session`, `AuthToken` tables — **do not create**. Authentication is entirely delegated to the external provider.
- **Soft delete:** `deletedAt DateTime?` — the entity is not deleted physically. Queries to such models always filter: `WHERE deletedAt IS NULL`.

## Entities and relations
<for each model: what it stores, key fields, relations — business language, not database language>

## Database schema

```prisma
// schema.prisma — <Product name>
// Generated based on SPEC.md

generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

// <models and enums>
```

## Enums
<for each enum: name, values, when and why it's used>

## Indexes and constraints
| Model | Field(s) | Type | Reason |
|--------|---------|-----|---------|

## Cascade operations
| Model | Relation field | onDelete | onUpdate | Rationale |
|--------|-----------|----------|----------|-------------|
```

### Mandatory rules

**Traceability.** Add a reference comment to each model:
```prisma
// The "Order" entity — SPEC.md §Entities
model Order {
```

**Technical fields.** Each model (except join tables, see below) must contain:
```prisma
createdAt DateTime @default(now())
updatedAt DateTime @updatedAt
```

**Soft delete.** For critical business entities (Users, Orders, Projects, and the like per SPEC.md), add `deletedAt DateTime?` unless SPEC.md forbids it. All queries to such models account for `WHERE deletedAt IS NULL`.

**Enums instead of strings.** For enumerable values — only `enum`, no `String`:
```prisma
// Correct:
enum OrderStatus { DRAFT CONFIRMED COMPLETED CANCELLED }

// Forbidden:
status String // instead of an enum
```

**Relations.** Each relation — with an explicit `@relation`, `onDelete`, `onUpdate`.

**Unique indexes.** `@@unique` for business-unique combinations of fields.

**Join tables.** Name them with a `_` in the name (for example, `_UserToRole`, `Post_Tag`). Technical fields are optional for them.

**Profile already exists.** Don't add a `User`, `Account`, `Session`, or `AuthToken` model — authentication is entirely on the side of the external OIDC provider. To link business entities to the user, use `profileId`:
```prisma
// Correct:
model Order {
  // ...
  profileId Int
  profile   Profile @relation(fields: [profileId], references: [id], onDelete: Cascade)
}

// Forbidden:
model User { /* duplicates Profile */ }
model Session { /* OIDC — external provider */ }
```

**Forbidden:**
- `String` for enumerable values (statuses, roles, types)
- Foreign keys without an explicit `onDelete`
- Models without an `id` field
- The `User`, `Account`, `Session`, `AuthToken` models (duplicate the boilerplate)

---

## Generation rules: API_CONTRACTS.md

### File structure

```markdown
# API contracts: <Product name>

**Status:** draft | **Date:** YYYY-MM-DD

## Links
- Specification: [SPEC.md](../2_PRODUCT_SPEC/SPEC.md)
- Data model: [DATABASE_MODEL.md](./DATABASE_MODEL.md)

## Implementation context
- **GraphQL engine:** MercuriusDriver — a GraphQL implementation for the Fastify web framework (used instead of Apollo Server; behavior is identical, only the configuration differs).
- **Auth guards** (a NestJS guard — a decorator that checks access rights before a resolver runs):
  - `@UseGuards(JwtAuthGuard)` — requires a valid JWT token
  - `@UseGuards(JwtOptionalAuthGuard)` — auth is optional (works both with and without a token)
  - `@UseGuards(JwtAuthGuard) @Roles(ProfileRole.ADMIN)` — only for the ADMIN role
  - `@CurrentUser()` — a resolver parameter decorator, returns the current `Profile`
- **Starter-template operations (already done, do not include in the Blueprint):** `me`, `updateProfile`, `profileUpdated`, `debug`, `echo`
- **Error handling:** `gqlErrorFormatter` automatically converts `ZodValidationException` and `HttpException` into correct GraphQL errors — no manual handling needed.
- **Subscriptions:** work via WebSocket + Redis pub/sub — the infrastructure is already configured, just use it.
- **OIDC operations** (login/logout/registration): `login`, `register`, `logout`, `refreshToken` — **do not describe**, they live in the OIDC provider (an external service).

## API overview
<number of operations, main domains, key decisions>

## Custom errors
| Code | Description | When it occurs |
|-----|----------|----------------|

## GraphQL schema

```graphql
# schema.graphql — <Product name>
# Generated based on SPEC.md

# <directives, scalars, types, inputs, queries, mutations>
```

## Operation descriptions
<for non-trivial operations: inputs, result, access rule, possible errors>
```

### Mandatory rules

**Traceability.** In the comment to each Query and Mutation:
```graphql
# Described in SPEC.md §Key operations — creating an order
createOrder(input: CreateOrderInput!): CreateOrderResult!
```

**Access directives.** Each Query and Mutation is marked with a comment:
```graphql
# @public
products(limit: Int! = 20, offset: Int! = 0): ProductsPage!

# @auth
myOrders(limit: Int! = 20, offset: Int! = 0): OrdersPage!

# @auth @hasRole(ADMIN)
deleteUser(id: ID!): DeleteUserResult!
```
If an operation is public — write `# @public`. Without this comment the operation is considered unmarked — an error.

**Strict pagination.** Any field returning a list must have pagination. Use a wrapper type:
```graphql
# Correct:
type Query {
  orders(limit: Int! = 20, offset: Int! = 0, filter: OrderFilterInput): OrdersPage!
}

type OrdersPage {
  items: [Order!]!
  total: Int!
  hasMore: Boolean!
}

# Forbidden:
type Query {
  orders: [Order!]!  # a bare array without pagination — an error
}
```

**Input types.** Each mutation takes a separate Input type, not a set of scalar arguments:
```graphql
# Correct:
createOrder(input: CreateOrderInput!): CreateOrderResult!

# Forbidden:
createOrder(userId: ID!, productId: ID!, quantity: Int!): Order!
```

**Custom errors.** For operations with business errors — a union result type:
```graphql
union CreateOrderResult = Order | OrderError | InsufficientStockError
```

**GraphQL engine: MercuriusDriver (Fastify).** Don't describe Apollo Server directives and mechanisms. Subscriptions — via WebSocket + Redis pub/sub (already configured in the boilerplate).

**Guard syntax in comments.** Use NestJS-style for precision:
```graphql
# @UseGuards(JwtAuthGuard)                      — requires auth
# @UseGuards(JwtOptionalAuthGuard)              — auth is optional
# @UseGuards(JwtAuthGuard) @Roles(ADMIN)        — only for the ADMIN role
# @public                                        — no auth
```

**Boilerplate operations are already implemented.** Don't include in the Blueprint:
`me`, `updateProfile(input)`, `profileUpdated`, `debug`, `echo`

**Forbidden to describe:** auth operations (login, register, logout, refreshToken) — that's the OIDC provider. File uploads and session management — if they don't carry unique business logic.

**Forbidden:**
- Bare arrays `[Entity!]!` in Query/Mutation
- Query/Mutation without an access directive
- Mutations with a set of scalar arguments instead of an Input type
- The `me`, `updateProfile`, `profileUpdated` operations (already in the boilerplate)
- The GraphQL types `LoginInput`, `RegisterInput`, `SessionType` (OIDC — external provider)

---

## Generation rules: ARCHITECTURE.md

### File structure

```markdown
# Architecture: <Product name>

**Status:** draft | **Date:** YYYY-MM-DD

## Links
- Specification: [SPEC.md](../2_PRODUCT_SPEC/SPEC.md)

## Implementation context

**Stack:**
- Backend: NestJS + Fastify, Prisma ORM, PostgreSQL, GraphQL (MercuriusDriver), Redis (BullMQ — task queues; pub/sub — GraphQL Subscriptions)
- Frontend: React + Vite, TanStack Router (file-based routing), FSD (Feature-Sliced Design — a methodology for structuring code by business layers), URQL (a GraphQL client with caching), Zustand (a UI-state store), ParaglideJS (an i18n library for Vite)

**Already implemented in the starter templates (do not include in the Blueprint):**

| Layer | What's done |
|----------|------------|
| Backend | Auth: the `JwtAuthGuard`, `JwtOptionalAuthGuard`, `RolesGuard` guards; OIDC JWT via an external provider |
| Backend | Profile GraphQL: the `me`, `updateProfile`, `profileUpdated` operations; the `Profile` model in the DB |
| Backend | Infrastructure: Health check `GET /health`, BullMQ, Redis pub/sub, i18n (en/ru), gqlErrorFormatter |
| Frontend | `features/auth` — OIDC PKCE auth, `AuthGuard` (blocks unauthenticated users), `MockAuthProvider` (for tests) |
| Frontend | `widgets/Header` (navigation + login/logout buttons), `pages/404`, `shared/api/graphql-client` (URQL with a Bearer token) |
| Frontend | Routes: `/callback` (OIDC redirect after login), `*` (404 via `defaultNotFoundComponent`) |

**Protected routes (TanStack Router):** the `beforeLoad` + `redirect()` pattern. The page content is wrapped in `AuthGuard` — a component that redirects an unauthenticated user.

## Backend: NestJS modules
| Module | Responsibility | Dependencies |
|--------|----------------|-------------|

## Frontend: FSD slices

### Entities (business entities)
| Entity | What it represents | Data model |
|---------|-----------------|---------------|

### Features (user actions)
| Feature | What the user does | Related API operation |
|------|------------------------|----------------------|

### Widgets (composite UI blocks)
| Widget | Which features it's made of | Where it's used |
|--------|---------------------|-----------------|

### Pages (application pages)
<a hierarchy of pages from prototype/ without file paths>

## Frontend: state management

### Server state (URQL cache)
<which data is cached via URQL: entities, lists, their invalidation>

### UI state (Zustand)
<which states are in Zustand: open modals, active filters, theme, local flags>

## Frontend: localization (i18n)
<ParaglideJS namespaces based on the prototype or "i18n not required">

## Environment variables
| Variable | Type | Purpose | BE/FE |
|-----------|-----|-----------|-------|
```

### Mandatory rules

**STRICT BAN on file paths.** Any strings like `src/features/auth`, `src/entities/user`, `app/pages/dashboard`, `components/Button` are forbidden. Only logical names: the `auth` feature, the `User` entity, the `Dashboard` page.

**State separation is mandatory.** Explicitly state for each data area:
- Data that came from the server and is cached → URQL
- Local UI state (not meant to be persisted to the server) → Zustand

You can't leave the "State management" section without concrete content or write "URQL is used" without naming the specific entities.

**FSD without implementation details.** Describe only the business meaning of the slice. Don't mention the file structure, barrel exports, or the folder model.

**ENV — business variables only.** Don't include `PORT`, `NODE_ENV`, `DATABASE_URL`, `REDIS_URL`, or the standard framework variables. Only the ones specific to this product.

**i18n.** If the prototype has text (labels, buttons, messages, notifications) — list the ParaglideJS namespaces. If the prototype has no text — write explicitly: "i18n not required".

**Router: TanStack Router (file-based).** Don't describe Next.js or React Router patterns. Protected pages use `beforeLoad` + `redirect()`:
```
// Protected route pattern (beforeLoad):
if not authenticated → redirect to home
the page content is wrapped in AuthGuard
```

**Existing FSD slices from the starter template — do not include in the Blueprint:**
- `features/auth` — OIDC auth, `AuthGuard`, `MockAuthProvider`
- `widgets/Header` — navigation with login/logout buttons
- `pages/404` — a "not found" page
- `shared/api/graphql-client` — URQL (the GraphQL client) with a Bearer token

**Zustand: the store pattern.** All new stores use the wrapper with devtools:
```typescript
// New Zustand store pattern:
create(devtools<MyStore>((set) => ({ ... })))
```

**Forbidden:**
- FSD file paths (`src/`, `app/`, `components/`)
- A "State management" section without a URQL/Zustand split
- Technical implementation patterns (DI containers, repositories, patterns)
- Mentioning libraries outside the approved stack
- Including existing FSD slices (`auth`, `Header`, `404`) in the Blueprint

---

## Generation rules: TESTING_PLAN.md

### File structure

```markdown
# Testing plan: <Product name>

**Status:** draft | **Date:** YYYY-MM-DD

## Links
- Specification: [SPEC.md](../2_PRODUCT_SPEC/SPEC.md) — the source of scenarios

## Testing context

**Backend (Vitest):**
- `E2EClient` — a built-in test utility from `test/utils/`, runs GraphQL queries with authentication via a special HTTP header `x-mock-sub` (without a real OIDC provider).
- Usage: `await client.loginAs(profile)` — run queries on behalf of a user.
- Mandatory in `beforeEach()`: `clearDatabase()` (clear the DB tables) + `clearRedis()` (reset the Redis cache) — so the tests don't affect each other.
- In `.env.test`: `OIDC_MOCK_ENABLED=true` — enables mock-auth mode, the JWT is verified via `x-mock-sub`, not via a real OIDC.

**Frontend (Vitest + React Testing Library + Playwright):**
- Unit/component tests: `tests/setup.ts` globally mocks `react-oidc-context` (the OIDC auth library) — the OIDC state is available in tests with no extra setup.
- E2E tests (Playwright): run with `VITE_MOCK_AUTH=true` — activates `MockAuthProvider`, which automatically authenticates the user in the browser without a real OIDC provider.

**Coverage targets:** lines / functions / statements ≥ 80%, branches ≥ 70%

## Unit tests: complex business logic
| What we test | Inputs | Expected result | Why it's non-trivial |
|--------------|----------------|--------------------|--------------------|

## E2E scenarios (Playwright)
| Scenario | Steps | Expected result | Source in SPEC.md |
|----------|------|--------------------|--------------------|

## Critical paths (from SPEC.md §Testing)
<a full transfer of the critical scenarios without abbreviations>

## Negative scenarios (from SPEC.md §Testing)
<a full transfer of the negative scenarios without abbreviations>
```

### Mandatory rules

- Unit tests — only for non-trivial logic: calculations, validations, state machines. Don't test simple CRUD.
- E2E scenarios are taken from SPEC.md §Testing. Each scenario references its source.
- Critical paths and negative scenarios from SPEC.md are transferred in full.

### Backend: concrete patterns

**E2E tests** use the ready-made infrastructure from `test/utils/`:
```typescript
// Authentication in a test:
await client.loginAs(profile)

// Query:
const result = await client.requestGraphQL<MyQuery>(query, variables)

// Cleanup (mandatory in beforeEach):
await clearDatabase()
await clearRedis()
```

**Auth mock:** `OIDC_MOCK_ENABLED=true` + the `x-mock-sub` header. Don't write your own mock mechanism.

**Unit tests:** Vitest + NestJS Testing Module. Service mocks — via the factories from `test/utils/mocks.ts`.

### Frontend: concrete patterns

**Three testing levels:**
| Level | Tool | What to test |
|---------|-----------|----------------|
| Unit | Vitest | Store logic, utilities, data transformations |
| Component | Vitest + React Testing Library | Component rendering, interactions, states |
| E2E | Playwright | User scenarios from SPEC.md |

**Auth in tests:**
- Unit/component: OIDC is mocked automatically via `tests/setup.ts` (unauthenticated by default). Override it locally in a specific test.
- E2E: build the app with `VITE_MOCK_AUTH=true` — `MockAuthProvider` authenticates automatically.

**Coverage targets (mandatory):**
- lines / functions / statements: ≥ 80%
- branches: ≥ 70%

---

## Generation rules: IMPLEMENTATION_GUIDE.md

### Purpose

An onboarding document: lets a new developer take the Blueprint and start working **without learning the boilerplate**. Generate it **last** — after DATABASE_MODEL.md, API_CONTRACTS.md, ARCHITECTURE.md, TESTING_PLAN.md, since it summarizes their contents.

### File structure

```markdown
# Implementation guide: <Product name>

**Status:** draft | **Date:** YYYY-MM-DD

## Stack

| Layer | Technology | Purpose |
|------|-----------|-----------|
| Backend | NestJS + Fastify | Node.js framework for the API server |
| GraphQL | MercuriusDriver | GraphQL server (Fastify adapter) |
| ORM | Prisma | TypeScript ORM for working with the DB |
| DB | PostgreSQL | Relational database |
| Queues | BullMQ + Redis | Background jobs and queues |
| Frontend | React + Vite | UI library + bundler |
| Router | TanStack Router | File-based routing for React |
| GraphQL client | URQL | API requests with caching |
| UI state | Zustand | Local state store |
| i18n | ParaglideJS | Localization (Vite plugin) |
| Backend tests | Vitest + E2EClient | Unit and E2E for the backend |
| Frontend tests | Vitest + RTL + Playwright | Unit, component, E2E for the frontend |

## What's already implemented

> These parts **don't need to be written** — they're ready in the project's starter templates.

### Backend
- **Authentication (OIDC JWT):** login via an external OIDC provider (Logto/Keycloak). NestJS guards: `JwtAuthGuard` (requires a token), `JwtOptionalAuthGuard` (optional), `RolesGuard` (by role). Resolver decorators: `@CurrentUser()` (get the current Profile), `@Roles(ProfileRole.ADMIN)` (restrict by role).
- **Profile — the user model:** `Profile { id, oidcSub @unique, roles, avatarUrl }`. The GraphQL operations are already implemented: `me` (get your profile), `updateProfile` (update it), `profileUpdated` (subscribe to changes).
- **Infrastructure:** Health check `GET /health`, BullMQ (task queues on Redis), Redis pub/sub (for GraphQL Subscriptions), nestjs-i18n en/ru, `gqlErrorFormatter` (automatic mapping of errors to the GraphQL format).

### Frontend
- **Authentication (OIDC PKCE):** `react-oidc-context` — a library for OAuth2/OIDC in React. The FSD slice `features/auth` contains: `AuthGuard` (blocks access for unauthenticated users), `MockAuthProvider` (automatic authentication in tests).
- **Ready-made UI components:** `widgets/Header` (navigation with login/logout buttons), `pages/404` (an error page with "Back" and "Home" buttons).
- **GraphQL API:** `shared/api/graphql-client` — a configured URQL client (a GraphQL client for React) with automatic insertion of the auth Bearer token.
- **Routing (TanStack Router):** `/callback` — the OIDC redirect handler after login; `*` — a 404 fallback via `defaultNotFoundComponent`.

## What needs to be implemented

> Everything described in this Blueprint.

### Backend modules
<the list from ARCHITECTURE.md §Backend: NestJS modules>

### Frontend slices (new)
<the list from ARCHITECTURE.md §Frontend: FSD slices — only the ones not in the boilerplate>

## Local run

### Requirements
- Node.js 20+, pnpm
- Docker (PostgreSQL + Redis)
- OIDC: mock mode is used for development

### Backend
```bash
pnpm install
cp .env.example .env
# Fill in: DATABASE_URL, REDIS_URL, JWT_SECRET, OIDC_ISSUER

docker compose up -d postgres redis
pnpm prisma migrate dev
pnpm start:dev
```

### Frontend
```bash
pnpm install
cp .env.example .env
# Fill in: VITE_API_URL, VITE_OIDC_AUTHORITY, VITE_OIDC_CLIENT_ID

VITE_MOCK_AUTH=true pnpm dev   # development with mock auth
```

## Testing

### Backend
```bash
pnpm test         # unit (Vitest)
pnpm test:e2e     # e2e (requires Postgres + Redis)
```

### Frontend
```bash
pnpm test                              # unit + component (Vitest + RTL)
VITE_MOCK_AUTH=true pnpm test:e2e      # e2e (Playwright)
```

**Coverage:** lines / functions / statements ≥ 80%, branches ≥ 70%

## Key decisions

<Rationale for non-trivial technical decisions.
Example: "Soft delete for Order — SPEC.md §Data forbids physical deletion".
If there are no non-trivial decisions — write "No special non-trivial decisions".>

## Links

- Specification: [SPEC.md](../../docs/<name>/2_PRODUCT_SPEC/SPEC.md)
- [DATABASE_MODEL.md](./DATABASE_MODEL.md)
- [API_CONTRACTS.md](./API_CONTRACTS.md)
- [ARCHITECTURE.md](./ARCHITECTURE.md)
- [TESTING_PLAN.md](./TESTING_PLAN.md)
```

### Mandatory rules

- `## Stack` — a complete technology table with no abbreviations
- `## What's already implemented` — a precise list of the boilerplate: the developer sees what **doesn't need to be written**
- `## What needs to be implemented` — a concrete list of modules and slices from ARCHITECTURE.md, not generic phrases
- `## Local run` — real shell commands with `.env` and docker; not "see README"
- `## Testing` — run commands + coverage targets
- `## Key decisions` — not an empty section: either a rationale for the decisions or the explicit phrase "no non-trivial decisions"

**Forbidden:**
- Empty sections
- `## What needs to be implemented` without a list of concrete modules/slices
- Generic phrases instead of commands in `## Local run`
- Duplicating the full contents of the other Blueprint documents

---

## Adaptability: project update mode

If the project already has a `schema.prisma` or `schema.graphql`:

1. **Don't rewrite the schema from scratch.** Read the existing schemas, describe only the delta.
2. **DATABASE_MODEL.md must contain a `## Migration plan` section:**
   ```markdown
   ## Migration plan
   
   ### New tables
   - `Notification` — user notifications
   
   ### Changes to existing tables
   - `User` — add the field `avatarUrl String?`
   - `Order` — add the enum value `OrderStatus.REFUNDED`
   
   ### Dangerous operations
   - None
   ```
3. **API_CONTRACTS.md** contains the sections `## New operations` and `## Changed operations`.
4. Don't break the existing schemas — only extend them.

---

## Reviewing the spec on request

When the user says **"check the spec"**, **"validate the blueprint"**, **"check blueprint"**, **"look at what I changed"** — perform the following sequence:

### Step 1. Run the validator script

```bash
# Standard mode:
python3 {BLUEPRINT} validate "ProjectName"

# If the project is an update to an existing one:
python3 {BLUEPRINT} validate "ProjectName" --update-mode
```

### Step 2. Read all 5 files in full

`IMPLEMENTATION_GUIDE.md`, `DATABASE_MODEL.md`, `API_CONTRACTS.md`, `ARCHITECTURE.md`, `TESTING_PLAN.md`

### Step 3. Manual check for conflicts with the boilerplate

- [ ] No `User`, `Session`, `AuthToken` models — the user is only via `Profile` (oidcSub)
- [ ] No `login`, `register`, `logout`, `refreshToken` GraphQL operations — OIDC is external
- [ ] No `me`, `updateProfile`, `profileUpdated` operations — already in the boilerplate
- [ ] ARCHITECTURE.md has no `features/auth`, `widgets/Header`, `pages/404` — they already exist
- [ ] Protected routes are described via the `beforeLoad` pattern
- [ ] The tests mention `E2EClient.loginAs()` (BE) or `VITE_MOCK_AUTH=true` (FE)
- [ ] No incompatible libraries (Apollo instead of Mercurius, React Router instead of TanStack)

### Step 4. Consistency check against SPEC.md

- [ ] All entities from SPEC.md §Entities are covered in DATABASE_MODEL.md
- [ ] All operations from SPEC.md §Key operations are present in API_CONTRACTS.md
- [ ] All pages from SPEC.md §Pages are reflected in ARCHITECTURE.md
- [ ] The scenarios from SPEC.md §Testing are transferred into TESTING_PLAN.md

### Step 5. Produce a structured report

```
## Blueprint review: <ProjectName>

### Validator script
✓ Passed without errors
— or —
✗ Errors: <list>
⚠ Warnings: <list>

### Conflicts with the boilerplate
✓ None found
— or —
✗ Found:
  - <issue> → <how to fix>

### Consistency with SPEC.md
✓ All requirements covered
— or —
⚠ Not covered:
  - <what's missing from the Blueprint>

### Conclusion
Ready for development / Needs fixes: <N issues>
```

After the fixes — re-run the validator and create a git commit.

---

## Workflow

> **Script:** `~/.config/opencode/skills/tech-blueprint/scripts/blueprint_validator.py`
> The script is NOT copied into the project — it is used directly from the skill folder.
> Below: `{BLUEPRINT}` = the full path above.

### Step 1. Input analysis

1. Read VISION.md and SPEC.md in full — this is the **single source of truth**
2. Check `prototype/`: record the list of pages and the transitions between them. **Don't transfer** the implementation — only the scope. The prototype is a draft, architectural decisions are made independently
3. Check for the presence of `schema.prisma`, `schema.graphql` → determine the mode (new / update)
4. Build a working list: entities, operations, pages, roles

### Step 2. Document generation

Fill in strictly in this order:
```
DATABASE_MODEL.md → API_CONTRACTS.md → ARCHITECTURE.md → TESTING_PLAN.md → IMPLEMENTATION_GUIDE.md
```
Each subsequent document builds on the previous one (the API on the DB models, the architecture on the API). IMPLEMENTATION_GUIDE.md — last: it summarizes all four documents.

### Step 3. Self-check before the validator

- [ ] Each Prisma model (except join tables with a `_` in the name) has `createdAt`/`updatedAt`
- [ ] `deletedAt DateTime?` is added for critical entities
- [ ] No bare arrays `[Type!]!` in Query/Mutation without pagination
- [ ] Each Query/Mutation has a comment with an access directive
- [ ] ARCHITECTURE.md has not a single file path (`src/`, `app/`)
- [ ] DATABASE_MODEL.md and API_CONTRACTS.md have links to SPEC.md
- [ ] If it's update mode — the `## Migration plan` section exists
- [ ] IMPLEMENTATION_GUIDE.md contains the sections `## Stack`, `## What's already implemented`, `## Local run`

### Step 4. Validation (mandatory)

```bash
# New project:
python3 {BLUEPRINT} validate "ProjectName"

# Updating an existing project:
python3 {BLUEPRINT} validate "ProjectName" --update-mode

# Non-standard output folder:
python3 {BLUEPRINT} validate "ProjectName" --output /path/to/blueprint
```

**The documentation is not considered ready until validation passes without errors.**

The script runs 8 or 9 checks:
1. The presence of all 5 files (including IMPLEMENTATION_GUIDE.md)
2. The presence of ` ```prisma ` and ` ```graphql ` blocks
3. The absence of FSD paths in ARCHITECTURE.md
4. Cross-check: Prisma models are covered in API_CONTRACTS.md (≥50%)
5. Traceability: links to SPEC.md in DATABASE_MODEL.md and API_CONTRACTS.md
6. Pagination: all Query/Mutation fields with `[...]` have pagination arguments
7. Technical fields: `createdAt`/`updatedAt` in each Prisma model (except join tables)
8. IMPLEMENTATION_GUIDE.md contains the mandatory sections (`## Stack`, `## What's already implemented`, `## Local run`)
9. *(only with `--update-mode`)* The presence of a "Migration plan" section in DATABASE_MODEL.md

### Step 5. Git commit (mandatory)

```bash
git add blueprint/<ProjectName>/3_TECH_BLUEPRINT/
git commit -m "feat(blueprint): technical contracts for <ProjectName>"
```

---

## Glossary

| Term | Meaning |
|--------|---------|
| Boilerplate | A starter project template with ready-made basic solutions (auth, infrastructure) |
| OIDC | OpenID Connect — an authentication protocol via an external service (Logto, Keycloak, Google) |
| JWT | JSON Web Token — a signed token confirming a user's identity |
| FSD | Feature-Sliced Design — a methodology for structuring the frontend by layers: app, pages, widgets, features, entities, shared |
| MercuriusDriver | A NestJS GraphQL adapter for the Fastify web framework (an alternative to Apollo Server) |
| URQL | A GraphQL client for React with a built-in cache and invalidation |
| Zustand | A minimalist state library for React (an alternative to Redux) |
| ParaglideJS | An i18n library with typed translation keys, runs as a Vite plugin |
| BullMQ | A task-queue library for Node.js based on Redis |
| E2EClient | A built-in test utility (`test/utils/`) for running GraphQL queries with mock auth |
| Profile | The user model in the DB; linked to OIDC via the `oidcSub @unique` field |
| TanStack Router | A file-based router for React: routes are defined by the structure of the `src/routes/` folders |
| AuthGuard | An FSD slice component that redirects an unauthenticated user away from a protected page |
| beforeLoad | A TanStack Router hook for checking conditions before a page loads (used for protecting routes) |

---

## Key constraints

- The documentation language is **Russian**, the code is Latin
- Stack: **NestJS, Prisma, PostgreSQL, GraphQL (MercuriusDriver), React, TanStack Router, FSD, URQL, Zustand, ParaglideJS**
- **FSD file paths are forbidden** in ARCHITECTURE.md — only logical names
- **All Query/Mutation with `[...]`** — must have pagination (limit/offset or first/after)
- **All Query/Mutation** — with an access-directive comment (`@public`, `@UseGuards(JwtAuthGuard)`, `@Roles(ADMIN)`)
- **Each Prisma model** (not a join table) — with `createdAt`/`updatedAt`
- **Enum instead of String** for all enumerable values
- **Traceability** — links to SPEC.md in the schema comments
- **Update mode** — a "Migration plan" section, not rewriting the schema
- **Don't duplicate the boilerplate**: Profile, me/updateProfile/profileUpdated, the auth slice, Header, the 404 page
- **Tests**: use `E2EClient.loginAs()` (BE) and `VITE_MOCK_AUTH=true` (FE), coverage ≥ 80%/70%
- After validation — a **git commit**
