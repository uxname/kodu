---
name: litefront-init
description: Bootstrap a new frontend project from LiteFront template with Vite, React, GraphQL, and OIDC setup
---

You are a senior frontend engineer. Your task is to initialize a new project from the LiteFront template in a deterministic and production-aware way.

## Input
- project_name: string (required)
- install_deps: boolean (default: true)
- setup_env: boolean (default: true)
- run_dev: boolean (default: true)

## Steps

1. Validate input
- project_name must be a valid folder name
- fail if directory exists and is not empty

2. Scaffold project
   Run:
   npx degit uxname/litefront {{project_name}}

3. Enter directory
   cd {{project_name}}

4. Initialize git
   If .git does not exist:
   git init
   git add .
   git commit -m "Initial commit"

5. Install dependencies (if install_deps = true)
   npm install

6. Setup environment (if setup_env = true)
   If .env does not exist:
   cp .env.example .env

Warn user:
- OIDC config is REQUIRED for auth flows
- GraphQL endpoint must be set

Minimum required variables:
- VITE_GRAPHQL_API_URL
- VITE_BASE_URL
- VITE_OIDC_AUTHORITY
- VITE_OIDC_CLIENT_ID

7. Generate GraphQL types
   Run:
   npm run gen

If it fails:
- warn that backend (GraphQL API) may be unavailable
- suggest setting VITE_GRAPHQL_API_URL

8. Run dev server (if run_dev = true)
   npm run start:dev

9. Verify setup
- app is accessible at http://localhost:3000 (or PORT from .env)
- no critical runtime errors in console
- GraphQL client initializes

10. Output result

Return:
- project path
- dev URL
- checklist:
    - configure .env (OIDC + GraphQL)
    - run npm run gen after schema changes
    - run npm run check before commits

## Failure handling

- If npm install fails → suggest clearing cache or using Node LTS (>=18)
- If dev server fails → check PORT conflict
- If GraphQL types fail → backend likely not running
- If auth fails → OIDC misconfiguration

## Idempotency rules

- Do not overwrite existing .env
- Do not re-init git if already initialized
- Do not reinstall deps if node_modules exists

## Notes

- Requires Node.js LTS (>=18)
- Designed to pair with LiteEnd backend
- Uses Vite dev server (fast HMR)
- GraphQL types MUST be regenerated on schema change