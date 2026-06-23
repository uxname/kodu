# Stack Profile: Rust   (id: rust)
Tier: general

A general-tier profile: neutral idioms + tooling pointers without
guarantees they are present. Mark stack-specific findings without unambiguous evidence
as `🔍 UNVERIFIED`.

## 1. Detection signals
- `Cargo.toml`
- additional: `Cargo.lock`, `rust-toolchain.toml`

## 2. Tooling by category
| Category | Command | How to read the output |
|-----------|---------|------------------|
| unused-code | `cargo +nightly udeps 2>/dev/null \|\| true`; `cargo build` warnings | unused → YAGNI-02 |
| clone-detection | — (no standard tool) | grep for duplicates → `🔍 UNVERIFIED` |
| dep-audit | `cargo audit 2>/dev/null \|\| true` | CVEs in crates |
| env-extraction | `grep -rEoh '(std::)?env::var\(?"[A-Z0-9_]+"' . 2>/dev/null \| sort -u` | env from code → DOC-02 |
| arch-lint | — | manual layer analysis → `🔍 UNVERIFIED` |
| lint/format | `cargo clippy 2>/dev/null \|\| true`; `cargo fmt --check 2>/dev/null` | — |
| type-check | `cargo check 2>/dev/null \|\| true` | compilation = type check |
| test-run | `cargo test 2>/dev/null \|\| true` | — |
| secret-scan | `gitleaks detect --no-banner 2>/dev/null \|\| true` | stack-neutral |

## 3. Idioms
- **Errors:** `Result<T, E>` + `?`; `thiserror`/`anyhow`; no `.unwrap()`/`.expect()` on production paths.
- **Concurrency:** `tokio`/`async`; cancellation via drop/`CancellationToken`; shared state via `Arc<Mutex<...>>`/channels.
- **Env/config:** `std::env::var` with `Result` handling; centralized config (`serde`/`config`).
- **Logging:** `tracing`/`log` with structure; no `println!` in production.
- **Null-safety:** `Option<T>` + pattern matching; the compiler guarantees the absence of null.
- **Deps:** ecosystem crates (`serde`, `itertools`, `tokio`); stdlib iterators.

## 4. Anti-patterns
- `.unwrap()`/`.expect()` on a path where an error is possible.
- `println!`/`eprintln!` for logging in production.
- `unsafe` without justification.
- Ignoring `Result` (`let _ = ...`) where the error needs handling.

## 5. Check-ID hints
- `LOG-01` → `println!`/`eprintln!` instead of `tracing`/`log`.
- `ERR-01` → `.unwrap()`/`.expect()`/an ignored `Result`.
- `BUG-03` → mostly **N/A** (no null; `Option` is checked by the compiler).
- `BUG-10` → ReDoS is usually **N/A** (the `regex` crate has no catastrophic backtracking).
- Other stack-specific → when evidence is lacking, `🔍 UNVERIFIED`.
