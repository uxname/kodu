---
name: audit-deployment
description: >
  Build and deployment audit: Dockerfile optimization, environment variables, CI/CD configs,
  secrets in configs, non-root users. Run on /audit-deployment.
---

## Relevance Rule

Applies when there is a Dockerfile, docker-compose.yml, CI/CD configs (`.github/workflows`, `gitlab-ci.yml`, `Jenkinsfile`), `.env` files, or Kubernetes manifests. For projects without deployment configuration, return an empty response.

## Runtime Detection & Stack Profile

This audit is stack-agnostic: the checks are framed neutrally, and the specifics
(tools, idioms, anti-patterns, examples) come from the stack profile.

1. **Profile passed in context?** If the `/audit` orchestrator passed
   `runtime=<id>` and/or the profile contents, use it and skip steps 2–3.

2. **Otherwise, detect EXACTLY ONE runtime** for this directory:
   ```bash
   if   [ -f package.json ]; then echo "runtime=node"
   elif [ -f go.mod ]; then echo "runtime=go"
   elif [ -f pyproject.toml ] || [ -f requirements.txt ] || [ -f setup.py ]; then echo "runtime=python"
   elif [ -f Cargo.toml ]; then echo "runtime=rust"
   elif [ -f pom.xml ] || ls build.gradle* settings.gradle* >/dev/null 2>&1; then echo "runtime=java"
   else echo "runtime=generic"; fi
   ```
   One run = one runtime; do not mix backend and frontend. If several markers
   are found (monorepo), pick the one matching the current scope / files under
   analysis and record the choice in the Audit Coverage section.

3. **Load the profile** via Read: `./skills/audit/stacks/<runtime>.md`
   (fallback `./skills/audit/stacks/_generic.md` if the file is not found).

Then use the profile:
- **Tools** — from the profile's "Tooling by category" section (the
  "Tooling Support" section below references categories, not commands).
- **PASS expectations** — from "Idioms"; **FAIL wording** — from "Anti-patterns".
- **Targeted hints** — from "Check-ID hints" by the prefix `DEP-`.
- If the profile is `tier: general` or `runtime=generic`, mark stack-specific
  findings without unambiguous evidence as `🔍 UNVERIFIED` rather than `❌ FAIL`.
  Mark checks whose mechanism is absent in the runtime as `N/A`.

## Severity Guide

| Severity | Assignment criterion |
|----------|---------------------|
| 🔴 Critical | RCE, auth bypass, data corruption, irreversible financial risk |
| 🟠 High | production outage, privilege escalation, data leak |
| 🟡 Medium | performance or maintainability degradation without an immediate outage |
| 🟢 Low | style, readability, minor convention violation |

Rule: severity = impact × exploitability × blast radius. The same pattern → the same severity across audits.

## Checklist

| Check ID | Check |
|----------|----------|
| DEP-01 | Docker images use pinned versions (no :latest) |
| DEP-02 | Containers run as an unprivileged user (USER nonroot) |
| DEP-03 | Multi-stage build: the final image has no build tools or dev artifacts |
| DEP-04 | .dockerignore excludes build artifacts/dependencies, the VCS directory, and secrets |
| DEP-05 | HEALTHCHECK is defined in the Dockerfile |
| DEP-06 | Secrets are not hardcoded in the Dockerfile (not in ENV) |
| DEP-07 | .env is excluded from VCS |
| DEP-08 | .env.example documents all environment variables |
| DEP-09 | The environment is switched to production mode: dev artifacts are disabled in prod |
| DEP-10 | Deterministic installation from the lockfile with integrity verification |
| DEP-11 | Container resource limits are defined (CPU limits, Memory limits) |
| DEP-12 | The ability to run with a read-only root filesystem is verified |

## Verification Rules

1. **Checklist only**: evaluate ONLY the checks above. Do not add new ones.
2. **Explicit verification = PASS**: assign `✅ PASS` only if you explicitly verified the mechanism (found the schema, config, guard) and confirmed there is no violation — state exactly what was checked.
3. **No evidence = UNVERIFIED**: if you cannot point to a `file:line` for either a violation or a confirmation, assign `🔍 UNVERIFIED`.
4. **Baseline takes priority**: if the check_id is in `docs/audit-baseline.yml` → `⏸ ACCEPTED`.
5. **Only 🔴/🟠 FAILs require a solution**: 🟡/🟢 — a solution is optional.

## Evidence Quality Rules

Every `❌ FAIL` must include:
- An exact `file:line`
- A minimal code snippet (1–3 lines)
- Causal chain: why this specific violation → what risk it creates

Not allowed:
- Assuming runtime behavior without evidence in the code
- Inferring the prod configuration from the dev configuration
- Assuming middleware is absent without checking the entire router chain
- If a conclusion rests on an assumption — only `🔍 UNVERIFIED`

## Language Rule

Audit results must be written in plain, clear language. Avoid complex terms, jargon, and abstract concepts unless necessary. Common technical terms (Docker, HTTP, API, JSON, URL) are fine. Describe problems so they are understandable to a developer of any level, not only a narrow specialist in the area.

## Baseline

Before analysis:
```bash
if [ ! -f ./docs/audit-baseline.yml ]; then
  mkdir -p ./docs
  cp ./skills/audit/audit-baseline-template.yml ./docs/audit-baseline.yml 2>/dev/null || \
    printf "accepted: []\n" > ./docs/audit-baseline.yml
fi
cat ./docs/audit-baseline.yml
```

## Analysis Context

> The examples below are illustrative. Take the specific build/deployment tools, idioms, and
> anti-patterns for the current runtime from the loaded profile
> (`stacks/<runtime>.md`, the Idioms/Anti-patterns/Check-ID hints sections by the
> `DEP-` prefix). Node: `npm ci`/`package-lock.json`, `NODE_ENV`, `node_modules`.
> Go: builder→distroless, `go.mod`+`go.sum`, `CGO_ENABLED=0`, no `net/http/pprof`.

**DEP-01 — Pinned versions:**
- A base image `:latest` without version pinning (unpredictable updates)
- A `:alpine` tag without a specific version
- Digest-based pinning is absent for critical images

**DEP-02 — Unprivileged user:**
- Running the container as root without switching to a non-root user (Node: `USER node`; Go: distroless `nonroot`)
- No creation of a non-root user before switching (or an image with a built-in nonroot)

**DEP-03 — Multi-stage build: the final image has no build tools or dev artifacts:**
- No multi-stage build (build tools/dev dependencies in the final image)
- Node: `devDependencies` are installed in the production stage
- Go: the compiler/tests are in the final image instead of `builder → distroless`
- Build artifacts are not copied from the builder stage

**DEP-04 — .dockerignore:**
- No `.dockerignore`, or it does not exclude dependencies/build artifacts (Node: `node_modules`; Go: a local `vendor/`, build cache) and the VCS directory `.git`
- `.env` and other secrets are not excluded from the Docker build context
- Tests and documentation end up in the production image

**DEP-05 — HEALTHCHECK:**
- HEALTHCHECK is absent from the Dockerfile
- A health check endpoint does not exist in the application

**DEP-06 — Secrets not in Dockerfile ENV:**
- Secrets in `ENV` directives of the Dockerfile (visible in docker inspect and image layers)
- Credentials in `ARG` without using build secrets (`--secret`)

**DEP-07 — .env excluded from VCS:**
- A `.env` file with real credentials committed to the repository
- Missing `.env*` in `.gitignore`

**DEP-08 — .env.example documents variables:**
- `.env.example` is missing
- `.env.example` contains real credentials
- Not all required variables are documented

**DEP-09 — Environment switched to production mode:**
- Dev artifacts are not disabled in prod: debug endpoints, verbose logs, dev dependencies
- Node: `NODE_ENV` is not set or is `development` in the prod image (leads to loading devDependencies at runtime)
- Go: no signs of prod mode (no `APP_ENV`/custom flag), `net/http/pprof` and debug routes are accessible in prod, no `-ldflags "-s -w"`, no `CGO_ENABLED=0`

**DEP-10 — Deterministic installation from the lockfile with integrity verification:**
- Node: `npm install` instead of `npm ci` (non-deterministic, slower); missing `package-lock.json`
- Go: `go.mod`/`go.sum` not committed; in the Dockerfile `go get` instead of `go mod download` from the lockfile; no `GOFLAGS=-mod=readonly` / `go mod verify`

**Resource limits:**
- `docker-compose.yml` without `deploy.resources.limits.memory` and `cpus`
- A Kubernetes Deployment without `resources.limits` on the container
- No limits → one container with a memory leak takes down the whole host
- Go note: with a container memory limit, set `GOMEMLIMIT`/`GOMAXPROCS` to the cgroup quotas (otherwise the GC sees the host's memory/CPU)

**Read-only filesystem:**
- A container without `read_only: true` (docker-compose) or `readOnlyRootFilesystem: true` (K8s)
- If the application writes temporary files, check for a tmpfs mount for `/tmp`

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| DEP-01 | Docker images use pinned versions (no :latest) | ✅ PASS | High | `Dockerfile:1` — a pinned version is specified | — | — |
| DEP-02 | Containers run as an unprivileged user (USER nonroot) | ❌ FAIL 🟠 | High | `Dockerfile:12` | **1. Add `RUN addgroup -S app && adduser -S app -G app` and `USER app`** \\ 2. Use an image with a built-in nonroot (Node: node:alpine; Go: distroless nonroot) \\ 3. Add `USER node` if the base image is node | No |
| DEP-05 | HEALTHCHECK is defined in the Dockerfile | ⏸ ACCEPTED | Medium | — | In baseline: health check is managed by the Kubernetes liveness probe | — |

Statuses: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED` / `🔍 UNVERIFIED`

Confidence: `High` — checked several key files, the pattern is obvious / `Medium` — checked selectively, the pattern is likely / `Low` — limited context, full certainty is impossible

For `❌ FAIL`: exactly 3 solution options, separated by `\\`, with option 1 in bold.

`Fixed`: FAIL → `No` (the developer changes it to `✅ Yes` manually after the fix). PASS / ACCEPTED / UNVERIFIED → `—`.

Solution requirements:
- Mutually exclusive (not rephrasings of the same thing)
- Match the project's current stack (do not propose switching frameworks)
- Do not require rewriting the whole system — realistic migration cost
- Option 3 may be "keep it, document the reason" if there is justification

At the end of the report, add a coverage section:
```
## Audit Coverage
Checked: src/module1/**, src/module2/**
Skipped: scripts/**, migrations/**, tests/**
Files checked: N | Skipped: N
```

If everything is PASS, output: `✅ The build and deployment configuration is in good shape.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-deployment.md`

```
# Audit Report: Build & Deployment Configuration — <YYYY-MM-DD HH:MM>
<table>
```
