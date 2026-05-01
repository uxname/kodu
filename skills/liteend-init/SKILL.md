---
name: liteend-init
description: Bootstrap a new backend project from LiteEnd template with Docker, Prisma, and initial setup
---

You are a senior backend engineer. Your task is to initialize a new project from the LiteEnd template in a clean, reproducible way.

## Input
- project_name: string (required)
- use_docker: boolean (default: true)
- install_deps: boolean (default: true)

## Steps

1. Validate input
- project_name must be a valid folder name
- fail if directory already exists and is not empty

2. Scaffold project
   Run:
   npx degit uxname/liteend {{project_name}}

3. Enter directory
   cd {{project_name}}

4. Initialize git
   git init

5. Setup environment
   cp .env.example .env

6. Install dependencies (if install_deps = true)
   npm install

7. Start infrastructure (if use_docker = true)
   docker-compose up -d db redis

Wait until services are ready (retry connection to DB up to 30s)

8. Run database migrations
   npm run db:migrations:apply

9. (Optional) Seed database
   If script exists:
   npm run db:seed

10. Verify setup
- check that app can start:
  npm run start:dev

- verify endpoints:
    - /health
    - /swagger

11. Output result

Return:
- project path
- next steps:
    - edit .env
    - run docker-compose up -d (if skipped)
    - npm run start:dev

## Failure handling

- If docker is not running → suggest fallback to manual PostgreSQL + Redis
- If migrations fail → show prisma error and suggest checking DATABASE_URL
- If port is busy → suggest changing PORT in .env

## Idempotency rules

- Do not overwrite existing .env
- Do not re-init git if .git exists
- Do not reinstall deps if node_modules exists

## Notes

- Requires Node.js LTS (>=18)
- Requires Docker for full setup
- Uses Prisma + PostgreSQL + Redis