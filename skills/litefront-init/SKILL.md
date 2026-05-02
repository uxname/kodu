---
name: litefront-init
description: Bootstrap a new frontend project from LiteFront template with Vite, React, GraphQL, and OIDC setup
---

You are a senior frontend engineer. Your task is to initialize a new project from the LiteFront template in a
deterministic and production-aware way.

## Input

- project_name: string (required)
- install_deps: boolean (default: true)
- setup_env: boolean (default: true)
- run_dev: boolean (default: true)

## Steps

1. Validate input

- Check Node.js version: `node -v` must be >= 20. Fail with a clear message if not.
- project_name must be a valid folder name (no spaces, no special chars except `-` and `_`)
- fail if directory exists and is not empty

2. Scaffold project
   Run:
   npx degit uxname/litefront {{project_name}}

3. Enter directory
   cd {{project_name}}

4. Initialize git (if .git does not exist)
   git init
   git add .
   git commit -m "Initial commit"

5. Install dependencies (if install_deps = true)
   npm install
   After install, commit the generated lockfile:
   git add package-lock.json && git commit -m "Add package-lock.json"

6. Setup environment (if setup_env = true)
   If .env does not exist:
   cp .env.example .env

   Warn user — these variables are REQUIRED before the app works:
    - VITE_GRAPHQL_API_URL (GraphQL endpoint)
    - VITE_BASE_URL
    - VITE_OIDC_AUTHORITY
    - VITE_OIDC_CLIENT_ID

7. Generate GraphQL types
   Only run if VITE_GRAPHQL_API_URL is set in .env AND the endpoint is reachable.
   Run:
   npm run gen

   If endpoint is not reachable or gen fails:
    - warn user that types generation was skipped
    - instruct to run `npm run gen` manually once the backend is available

8. Run dev server (if run_dev = true)
   npm run start:dev

9. Verify setup (only if run_dev = true)
    - app is accessible at http://localhost:3000 (or PORT from .env)
    - no critical runtime errors in console
    - GraphQL client initializes

10. Output result

Return:

- project path
- dev URL
- checklist:
    - configure .env (OIDC + GraphQL) if not done yet
    - run `npm run gen` after every schema change
    - run `npm run check` before commits

## Failure handling

- If Node.js < 20 → stop immediately, instruct user to upgrade
- If npm install fails → suggest clearing cache: `npm cache clean --force`, then retry
- If dev server fails → check PORT conflict, suggest changing PORT in .env
- If GraphQL types fail → backend likely not running; skip gen and continue
- If auth fails → OIDC misconfiguration; check VITE_OIDC_* vars

## Notes

- Requires Node.js LTS (>=20)
- Designed to pair with LiteEnd backend
- Uses Vite dev server (fast HMR)
- GraphQL types MUST be regenerated on schema change
