---
name: liteend-init
description: Bootstrap a new backend project from LiteEnd template with Docker, Prisma, and initial setup
---

You are a senior backend engineer. Your task is to initialize a new project from the LiteEnd template in a clean,
reproducible way.

## Input

- project_name: string (required)
- use_docker: boolean (default: true)
- install_deps: boolean (default: true)

## Steps

1. Validate input

- Check Node.js version: `node -v` must be >= 20. Fail with a clear message if not.
- project_name must be a valid folder name (no spaces, no special chars except `-` and `_`)
- fail if directory already exists and is not empty

2. Scaffold project
   Run:
   npx degit uxname/liteend {{project_name}}

3. Enter directory
   cd {{project_name}}

4. Initialize git (if .git does not exist)
   git init
   git add .
   git commit -m "Initial commit"

5. Setup environment
   If .env does not exist:
   cp .env.example .env

6. Install dependencies (if install_deps = true)
   npm install

7. Start infrastructure (if use_docker = true)
   docker compose up -d db redis

8. Run database migrations
   npm run db:migrations:apply

9. Seed database
   Check if `db:seed` key exists in package.json scripts. If yes:
   npm run db:seed

10. Verify setup
    Launch dev server in background, then check endpoints:
    npm run start:dev &
     - GET /health → expect 200
     - GET /swagger → expect 200
    Stop the dev server after verification.

11. Output result

Return:

- project path
- next steps:
    - edit .env (DATABASE_URL, Redis config, etc.)
    - run `docker compose up -d` (if skipped)
    - `npm run start:dev`

## Failure handling

- If Node.js < 20 → stop immediately, instruct user to upgrade
- If docker is not running → suggest fallback to manual PostgreSQL + Redis
- If migrations fail → show prisma error and suggest checking DATABASE_URL
- If port is busy → suggest changing PORT in .env

## Notes

- Requires Node.js LTS (>=20)
- Requires Docker for full setup
- Uses Prisma + PostgreSQL + Redis
