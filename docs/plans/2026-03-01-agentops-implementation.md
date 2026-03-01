# AgentOps Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a new JSON-only `kodu ops` command namespace for remote server diagnostics and operations over SSH.

**Architecture:** Extend config with an optional `ops.servers` map, add a shared `SshService` that wraps `execa('ssh')` and never throws for remote failures, and implement `ops` subcommands (`sysinfo`, `env`, `routes`, `service`) that validate inputs locally and return strict JSON responses.

**Tech Stack:** NestJS + nest-commander, zod, execa, node:fs/promises.

---

### Task 1: Extend Config Schema For Ops

**Files:**
- Modify: `src/core/config/config.schema.ts`

**Step 1: Write the failing check (type-level usage target)**

Use this target snippet as acceptance criteria while editing:

```ts
const cfg = configService.getConfig();
const maybeServer = cfg.ops?.servers?.prod;
```

**Step 2: Run type check to confirm current gap**

Run: `npm run ts:check`
Expected: type errors in future ops command code until `ops` is added.

**Step 3: Add minimal schema/types**

Add `serverConfigSchema`, `opsSchema`, and `ops: opsSchema.optional()` in root schema.

**Step 4: Run type check**

Run: `npm run ts:check`
Expected: PASS for config layer.

**Step 5: Commit**

```bash
git add src/core/config/config.schema.ts
git commit -m "feat(config): add ops servers schema"
```

### Task 2: Add Shared SSH Module And Service

**Files:**
- Create: `src/shared/ssh/ssh.service.ts`
- Create: `src/shared/ssh/ssh.module.ts`

**Step 1: Define failing usage target**

Target call shape:

```ts
const result = await sshService.execute(server, 'uptime');
if (!result.success) {
  console.log(result.stderr);
}
```

**Step 2: Run type check to verify service does not exist yet**

Run: `npm run ts:check`
Expected: FAIL when referenced by future ops command wiring.

**Step 3: Implement minimal `SshService`**

Implement:
- `SshResult` type
- `execute(serverConfig, command)`
- `execa('ssh', args, { env: serverConfig.env })`
- capture non-zero exits and system errors without throw

**Step 4: Export via `SshModule`**

Use Nest module pattern consistent with shared modules.

**Step 5: Run type check**

Run: `npm run ts:check`
Expected: PASS for ssh module/service.

**Step 6: Commit**

```bash
git add src/shared/ssh/ssh.service.ts src/shared/ssh/ssh.module.ts
git commit -m "feat(ops): add shared ssh service"
```

### Task 3: Scaffold Ops Command Module

**Files:**
- Create: `src/commands/ops/ops.module.ts`
- Create: `src/commands/ops/ops.command.ts`
- Modify: `src/app.module.ts`

**Step 1: Create failing command target**

Expected CLI shape:

```bash
kodu ops --help
```

**Step 2: Run build to confirm command missing**

Run: `npm run build`
Expected: `ops` command not present before scaffolding.

**Step 3: Implement root ops module/command**

Add root namespace command and register providers/imports.

**Step 4: Register in AppModule**

Import `OpsModule` (and `SshModule` if needed at root level).

**Step 5: Run build**

Run: `npm run build`
Expected: PASS and `kodu ops --help` shows namespace.

**Step 6: Commit**

```bash
git add src/commands/ops/ops.module.ts src/commands/ops/ops.command.ts src/app.module.ts
git commit -m "feat(ops): register ops root namespace"
```

### Task 4: Implement Shared Ops Validation + JSON Helpers

**Files:**
- Create: `src/commands/ops/ops.types.ts`
- Create: `src/commands/ops/ops.utils.ts`

**Step 1: Define failing usage target**

Helpers should support:

```ts
const server = await resolveServerOrThrow(config, alias);
await assertSshKeyExists(server);
printOk({ ping: true });
printSshError(result, command);
```

**Step 2: Implement helpers**

Include:
- server alias resolution
- ssh key path normalization (`absolute` or relative to `process.cwd()`)
- `fs.access` validation
- consistent JSON output builders

**Step 3: Run type check**

Run: `npm run ts:check`
Expected: PASS.

**Step 4: Commit**

```bash
git add src/commands/ops/ops.types.ts src/commands/ops/ops.utils.ts
git commit -m "feat(ops): add validation and json response helpers"
```

### Task 5: Implement `ops sysinfo`

**Files:**
- Create: `src/commands/ops/subcommands/ops-sysinfo.command.ts`
- Modify: `src/commands/ops/ops.module.ts`

**Step 1: Write failing command run target**

```bash
kodu ops sysinfo dev
```

Expected shape:

```json
{"status":"ok","data":{"uptime":"...","disk_usage":"...","mem_free":"..."}}
```

**Step 2: Run build or command to confirm missing subcommand**

Run: `npm run build`
Expected: missing subcommand before implementation.

**Step 3: Implement command**

Use agreed SSH payload and `SshService`. Map non-zero SSH result to JSON error without crash.

**Step 4: Verify**

Run: `npm run build && npm run ts:check`
Expected: PASS.

**Step 5: Commit**

```bash
git add src/commands/ops/subcommands/ops-sysinfo.command.ts src/commands/ops/ops.module.ts
git commit -m "feat(ops): add sysinfo subcommand"
```

### Task 6: Implement `ops env`

**Files:**
- Create: `src/commands/ops/subcommands/ops-env.command.ts`
- Modify: `src/commands/ops/ops.module.ts`

**Step 1: Write failing run targets**

```bash
kodu ops env dev get my-app
kodu ops env dev set my-app --key PORT --val 3001
kodu ops env dev unset my-app --key PORT
```

**Step 2: Implement action validation and command builders**

Implement `get|set|unset`, enforce required flags, derive `.env` path from `apps` root.

**Step 3: Implement JSON-only outputs**

- success: `status: ok`
- remote failure: `status: error` + code/stderr/command
- local validation failure: JSON error in stderr

**Step 4: Verify**

Run: `npm run build && npm run ts:check`
Expected: PASS.

**Step 5: Commit**

```bash
git add src/commands/ops/subcommands/ops-env.command.ts src/commands/ops/ops.module.ts
git commit -m "feat(ops): add env management subcommand"
```

### Task 7: Implement `ops routes` (list/add/remove/update)

**Files:**
- Create: `src/commands/ops/subcommands/ops-routes.command.ts`
- Modify: `src/commands/ops/ops.module.ts`

**Step 1: Write failing run targets**

```bash
kodu ops routes dev list
kodu ops routes dev add --domain api.example.com --upstream 127.0.0.1:3000
kodu ops routes dev update --domain api.example.com --upstream 127.0.0.1:4000
kodu ops routes dev remove --domain api.example.com
```

**Step 2: Implement path resolution and list**

- caddy root: `server.paths.caddy ?? path.posix.join(appsPath, 'caddy')`
- list: read `data/Caddyfile` and return raw text

**Step 3: Implement add/remove/update edits**

Execute remote `node -e` transformation scripts over SSH for deterministic block edits.

**Step 4: Apply config via caddy script**

After successful edit, run: `cd <caddyRoot> && ./caddy.sh`

**Step 5: Add NOT_FOUND behavior**

If domain absent for `remove`/`update`, return structured JSON error with code `NOT_FOUND`.

**Step 6: Verify**

Run: `npm run build && npm run ts:check`
Expected: PASS.

**Step 7: Commit**

```bash
git add src/commands/ops/subcommands/ops-routes.command.ts src/commands/ops/ops.module.ts
git commit -m "feat(ops): add caddy routes management subcommand"
```

### Task 8: Implement `ops service`

**Files:**
- Create: `src/commands/ops/subcommands/ops-service.command.ts`
- Modify: `src/commands/ops/ops.module.ts`

**Step 1: Write failing run targets**

```bash
kodu ops service dev status my-app
kodu ops service dev up my-app
```

**Step 2: Implement actions**

Support `clone|pull|up|down|logs|status` with remote compose/git commands.

**Step 3: Add status JSON fallback**

Attempt `docker compose ps --format json`; fallback to raw `docker compose ps` output if unsupported.

**Step 4: Verify**

Run: `npm run build && npm run ts:check`
Expected: PASS.

**Step 5: Commit**

```bash
git add src/commands/ops/subcommands/ops-service.command.ts src/commands/ops/ops.module.ts
git commit -m "feat(ops): add service lifecycle subcommand"
```

### Task 9: Full Validation And Manual Smoke Checks

**Files:**
- Modify: `kodu.json` (local test aliases only if needed)

**Step 1: Run static checks**

Run: `npm run check`
Expected: PASS.

**Step 2: Build final artifact**

Run: `npm run build`
Expected: PASS.

**Step 3: Manual command checks against test server**

Run and confirm valid JSON output each time:

```bash
kodu ops sysinfo <alias>
kodu ops env <alias> get <project>
kodu ops routes <alias> list
kodu ops service <alias> status <project>
```

**Step 4: Record results in PR/body notes**

Capture one success sample and one remote error sample to prove agent-readable behavior.

**Step 5: Commit (if test fixture/config changed)**

```bash
git add <changed-files>
git commit -m "test(ops): validate json outputs and command flows"
```
