---
name: audit-matrix
description: >
   Builds an architectural model of the system and analyzes failure scenarios
   for the discovered components and their connections. Runs on /audit-matrix.
---

## Task

A high-level architecture audit: find the system's major components, their connections,
and systemic risks (cascading failures, bottlenecks, lack of protection at the architectural level).
Do not look for small bugs — that is the job of other audits.

Do not rely on predefined lists of technologies. Adapt to any stack.

**Group similar entities.** Several services, replicas, or workers of the same kind — one component.
Example: 3 replicas of payment-worker → the "Payment Workers" component. Do not multiply rows for the sake of detail.

---

## Language

Write so a junior in their first week on the job would understand.
Short phrases, no bureaucratese, no fancy terms.

**Explaining a risk** — as if telling a colleague over lunch:
✅ "If the DB is slow, the site goes down for everyone in 5 seconds"
❌ "Connection pool exhaustion under concurrent load leads to degradation"
❌ "Everything breaks"

**A solution** — what to do and why:
✅ "Wait no more than 5 seconds for a response from the DB (`timeout: 5000` in the config)"
❌ "Configure connectionTimeoutMillis"
❌ "Adopt a service mesh"

If a term is needed, explain it right away:
"Circuit breaker (a fuse: if an external service is failing, we stop calling it)".

---

## Step 1 — Session

```bash
LATEST=$(ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1)
SESSION=${LATEST:-./docs/audits/$(date +"%Y-%m-%d_%H-%M")}
mkdir -p "$SESSION" && echo "Session: $SESSION"
```

---

## Step 2 — Component discovery

For each one, record: name, role, placement (Docker/cloud/SaaS),
how it was discovered (`file:line`).

**Grouping rule.** Merge into one component:
- replicas of the same service
- workers of the same queue
- microservices of the same area, if their risks are the same
- several tables/collections of one DB → the "Database" component, not "Users DB", "Orders DB"

If more than 12 components remain after grouping, merge even more boldly.

---

## Step 3 — Component graph

Do not search for addresses via grep — **trace the lifecycle of the configuration**.

Build a matrix: rows are who calls, columns are who is called.
In the cell, put **why** they are connected, not the protocol. Verb + object, up to 5 words.

An empty cell = no connection.

```
| Who \ Whom   | Frontend | Backend | DB | Redis | Payments |
|--------------|----------|---------|------|---------|---------|
| Frontend     |          | sends orders |   |         |         |
| Backend      |          |         | stores orders | session cache | charges money |
| Worker       |          |         | reads tasks |      |         |
| Payments     |          | sends status |   |         |         |
```

If it does not fit in the cell, simplify. Not "asynchronously publishes a payment event to the queue" → "sends payment status".

---

## Step 4 — Risk analysis

Match **components and their clients** (Step 2)
against the connection settings (Step 3).
If a library supports timeouts and retries but the code does not set them, that is a risk.

Approach it like an "evil genius": do not trust components inside the network, look for rare but destructive scenarios.
Find **real** risks. If a component is reliable, say so:
"No risks: a cloud service with auto-recovery".
Do not invent problems for the sake of count.

### How to describe the danger

The **"What's the danger"** column is the main part of the table. In a single phrase:
1. What breaks (cause)
2. What happens because of it (mechanics)
3. What the user or business sees (consequence)

✅ "An external API stalls → requests pile up → in 5 seconds the site goes down for everyone"
✅ "Two transactions charged money at the same time → the balance went negative, we'll only find out at month's end"
❌ "A service failure leads to cascading degradation" (a junior won't get it)
❌ "Race condition in billing" (a term without explanation)

### How to describe the solution

The Pareto principle: one action that closes 80% of the problem.
Formula: **what to do + why + the specific parameter in parentheses**.

✅ "Wait no more than 5 seconds for a response from the DB so that stuck requests don't pile up (`timeout: 5000`)"
✅ "Add a fuse: if Payments don't respond 3 times in a row, stop calling them for a minute (the opossum library)"
✅ "Wrap the charge and balance update in a single transaction (`BEGIN` … `COMMIT`)"

❌ "Rewrite the service to event-sourcing" (a refactor, not a fix)
❌ "Adopt a service mesh" (a huge task)
❌ `connectionTimeoutMillis: 5000` (without explaining why)

If the problem cannot be fixed locally, write `⚠ refactor needed` and one line on which direction.

### Cascading scenarios

Describe one separately **only** if the scenario:
- involves **≥3 components**, AND
- has its own mechanics (does not reduce to a single component's risk).

```
X1: Frontend → Backend → Payments
What's the danger: Payments stall → the backend waits → in 10 seconds the whole site returns 503, even though the DB and frontend themselves are alive
Solution: wait no more than 3 seconds for a response from Payments + a fuse
Confirmation: `src/payment.ts:15` — Stripe without a timeout
```

---

## Step 5 — Output

Save `{SESSION}/audit-matrix.md`:

```markdown
# Matrix — {project} — {date}

## TL;DR
- **Architecture:** [monolith / microservices / serverless / monorepo]
- **Components:** N
- **Critical risks (🔴/🟠):** N
- **Top risk:** [one sentence]

## Component graph

| Who \ Whom | Frontend | Backend | DB | Redis | Payments |
|------------|----------|---------|------|---------|---------|
| Frontend   |          | sends orders |   |         |         |
| Backend    |          |         | stores orders | session cache | charges money |
| Worker     |          |         | reads tasks |      |         |

Connections: N. Synchronous (waiting for a response): N — this is the main source of cascading failures.

## Critical tasks

A summary of all 🔴/🟠 — both per component and cascading.

| # | What's the danger | Component | Risk | Solution | File |
|---|----------------|-----------|------|---------|------|
| 1 | DB stalls → requests pile up → in 5 seconds the site goes down for everyone | DB | 🔴 | Wait no more than 5 seconds for the DB, increase the connection pool (`timeout: 5000`, `max: 20`) | `src/db.ts:12` ❌ |
| 2 | Payments stall → the backend waits → the whole site returns 503 | Payments | 🔴 | 3-second timeout + a fuse | `src/payment.ts:15` ❌ |

## {Component}

**Role:** … • **Where:** Docker • **Found:** `file:line`

| # | Scenario | What's the danger | Risk | Solution | File |
|---|----------|-----------------|------|---------|------|
| A1 | DB unavailable | DB stalls → requests pile up → in 5 seconds the site goes down for everyone at 100 requests/sec | 🔴 | Wait no more than 5 seconds for the DB, retry 3 times (`timeout: 5000`, `maxRetries: 3`) | `src/db.ts:12` ❌ |
| Z1 | Transaction interrupted | The charge and balance update are not wrapped in a transaction → money was charged, the balance was not updated. We'll only find out during reconciliation | 🔴 | Wrap in a single transaction (`BEGIN` … `COMMIT`) | `src/billing.ts:44` ❌ |
| N1 | Redis goes down | The cache drops out → the backend goes straight to the DB. Slower, but it works | 🟡 | Protection is already in place, nothing to do | `src/cache.ts:8` ✅ |

If there are no risks, write one line: "No risks: a cloud service with auto-recovery".

## Cascading scenarios

| # | Chain | What's the danger | Risk | Solution | File |
|---|---------|-----------------|------|---------|------|
| X1 | Frontend → Backend → Payments | Payments stall → the backend waits → in 10 seconds the whole site returns 503, even though the DB and frontend are alive | 🔴 | 3-second timeout + a fuse (opossum) | `src/payment.ts:15` ❌ |
```

---

## Rules

1. **One problem = one row.** Only 🔴/🟠 go into the summary; all of them go into the component table.
2. **Only major components.** Not classes, not functions, not individual tables.
3. **Similar entities — into one component.** Replicas, workers, tables of one DB.
4. **Every scenario = `file:line`.** Otherwise `⚠ not verified — reason`.
5. **"What's the danger" is self-contained.** Cause + mechanics + consequence in one phrase, without terms.
6. **No real risk — don't invent one.** Write "No risks".
7. **The solution — by Pareto and in plain words.** What to do + why + the parameter in parentheses. If a refactor is needed — `⚠ refactor needed`.
8. **A cascade — only with ≥3 components** and unique mechanics.
9. **Baseline** — mark risks accepted via baseline as `⏸ ACCEPTED`.

---

## Severity

| 🔴 Critical | Data loss, security hole, complete failure without auto-recovery |
| 🟠 High | The site slows down or fails under load, secondary features do not work |
| 🟡 Medium | Rare errors, fixed by a restart |
| 🟢 Low | Tech debt, no monitoring in non-critical places |

---

## Bash

Bash is for navigation only. Found a file — open and read it.
No result — check the negative scenario (other names, commented out, a different extension).
Do not copy blindly — adapt to the project's structure.

---

## Baseline

```bash
cat ./docs/audit-baseline.yml 2>/dev/null
```

---

At the end of the message, one line:
**Paradigm:** … · **Components:** N · **Scenarios:** N · **🔴/🟠:** N

## Step 6 — Verification

`Skill("audit-verify")`

# Motto: Find one real hole, not a hundred imagined ones.
The auditor's four principles:
1. Look for a catastrophe, not a typo — you care about cascading failures and data loss, not code style
2. Think like an evil genius — do not trust the internal network, clouds, or "stable" APIs
3. Write as if for a junior — if you can't explain it in plain words, you didn't understand the risk yourself
4. Confirm with a line of code — every "maybe" = file:line, otherwise it's a hallucination