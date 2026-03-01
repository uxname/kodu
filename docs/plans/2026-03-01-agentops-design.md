# AgentOps Design

## Context

Add a new `kodu ops` namespace for AI agents to manage remote servers through SSH. The output contract is strict JSON in stdout/stderr with no spinner UI, no colors, and no interactive prompts.

## Goals

- Add project-local config for remote server aliases in `kodu.json`.
- Add a shared SSH abstraction over `execa` that does not throw on remote command failures.
- Add `kodu ops` subcommands for server diagnostics, env management, Caddy routes, and Docker Compose service lifecycle.
- Preserve machine-readable behavior for all success and error paths.

## Config Design

Extend `src/core/config/config.schema.ts` with optional `ops` section:

- `ops.servers` is a record keyed by alias (`prod`, `dev`, etc.).
- Each server includes:
  - `host`, `user`, `sshKeyPath`
  - `port` default `22`
  - optional `description`
  - optional `paths` with:
    - `apps` default `/var/agent-apps`
    - optional `caddy` for custom Caddy project path
  - optional `env` map for SSH process environment variables

All commands validate:

1. alias exists in `ops.servers`
2. ssh key file exists (`fs.access`)

## SSH Shared Module Design

Create `src/shared/ssh/` with:

- `ssh.module.ts`
- `ssh.service.ts`

`SshService.execute(serverConfig, command)` behavior:

- Builds `ssh` args:
  - `-i <keyPath>`
  - `-p <port>`
  - `-o StrictHostKeyChecking=no`
  - `-o ConnectTimeout=10`
  - `<user>@<host>`
  - `<command>`
- Runs via `execa('ssh', args, { env })`
- Returns:

```ts
type SshResult = {
  success: boolean;
  stdout: string;
  stderr: string;
  exitCode: number;
  error?: string;
};
```

No throw for SSH command failures. Local/system failures also map to `SshResult` with `success: false`.

## Commands Design

Create `src/commands/ops/`:

- `ops.module.ts`
- `ops.command.ts`
- `subcommands/ops-sysinfo.command.ts`
- `subcommands/ops-env.command.ts`
- `subcommands/ops-routes.command.ts`
- `subcommands/ops-service.command.ts`

Register `OpsModule` in `src/app.module.ts`.

All `ops` commands:

- no `UiService`
- no spinners
- no prompts
- print JSON only:
  - success/data via `console.log(JSON.stringify(...))`
  - fatal CLI errors via `console.error(JSON.stringify(...))` and `process.exitCode = 1`

## JSON Contracts

Success pattern:

```json
{ "status": "ok", "data": {} }
```

Remote command failure pattern:

```json
{
  "status": "error",
  "code": 255,
  "stderr": "ssh: connect to host ... port 22: Connection refused",
  "command": "uptime"
}
```

Local validation/fatal pattern:

```json
{
  "status": "error",
  "code": "VALIDATION_ERROR",
  "error": "Server alias 'prod' not found in kodu.json"
}
```

## Subcommand Behavior

### `ops sysinfo <alias>`

Runs the agreed payload and proxies JSON response:

- uptime
- root disk usage
- free memory (MB)

### `ops env <alias> <action> <project>`

Actions: `get`, `set`, `unset`.

- target file: `${appsPath}/${project}/.env`
- `get`: read file contents
- `set`: update or append key
- `unset`: remove key line

Required options:

- `set`: `--key`, `--val`
- `unset`: `--key`

### `ops routes <alias> <action>`

Use Caddy project workflow from `/home/dex/Рабочий стол/Work/caddy-caddy` model:

- Caddy runs via Docker Compose in separate project folder
- config source is `data/Caddyfile`
- apply command is `./caddy.sh` (default apply mode)

Actions:

- `list`: return raw Caddyfile
- `add`: add domain `reverse_proxy` block, then run `./caddy.sh`
- `remove`: remove domain block, then run `./caddy.sh`
- `update`: update domain upstream, then run `./caddy.sh`

Path resolution for Caddy project:

- `server.paths.caddy` if configured
- fallback `${appsPath}/caddy`

### `ops service <alias> <action> <project>`

Actions: `clone`, `pull`, `up`, `down`, `logs`, `status`.

- `clone`: clone repo to project path (skip/fail if exists)
- `pull`: git pull in project dir
- `up`: docker compose up -d
- `down`: docker compose down
- `logs`: docker compose logs
- `status`: docker compose ps (json format when available)

## Caddyfile Editing Strategy

For `routes add/remove/update`, perform robust remote text transforms with a small `node -e` script executed over SSH:

- read Caddyfile
- locate target domain block
- apply deterministic update
- write to temp file then rename
- return structured JSON from remote helper

Then run `cd <caddyPath> && ./caddy.sh`.

If domain not found for `remove`/`update`, return JSON error with code `NOT_FOUND`.

## Error Handling Rules

- SSH exit code non-zero must not crash command.
- Return JSON with `status:error`, numeric `code`, `stderr`, and original `command`.
- Keep stdout/stderr machine-friendly; no additional prose lines.

## Out of Scope

- Caddy Admin API integration.
- Interactive operator UX.
- Non-JSON presentation layer.
