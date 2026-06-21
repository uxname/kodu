# Stack Profile: Rust   (id: rust)
Tier: general

Профиль уровня general: нейтральные идиомы + ориентиры по инструментам без
гарантий их наличия. Стек-специфичные находки без однозначного evidence
помечай `🔍 UNVERIFIED`.

## 1. Detection signals
- `Cargo.toml`
- доп.: `Cargo.lock`, `rust-toolchain.toml`

## 2. Tooling by category
| Категория | Команда | Как читать вывод |
|-----------|---------|------------------|
| unused-code | `cargo +nightly udeps 2>/dev/null \|\| true`; warnings `cargo build` | неиспользуемое → YAGNI-02 |
| clone-detection | — (нет стандартного инструмента) | grep повторов → `🔍 UNVERIFIED` |
| dep-audit | `cargo audit 2>/dev/null \|\| true` | CVE в crates |
| env-extraction | `grep -rEoh '(std::)?env::var\(?"[A-Z0-9_]+"' . 2>/dev/null \| sort -u` | env из кода → DOC-02 |
| arch-lint | — | ручной разбор слоёв → `🔍 UNVERIFIED` |
| lint/format | `cargo clippy 2>/dev/null \|\| true`; `cargo fmt --check 2>/dev/null` | — |
| type-check | `cargo check 2>/dev/null \|\| true` | компиляция = проверка типов |
| test-run | `cargo test 2>/dev/null \|\| true` | — |
| secret-scan | `gitleaks detect --no-banner 2>/dev/null \|\| true` | стек-нейтрально |

## 3. Idioms
- **Errors:** `Result<T, E>` + `?`; `thiserror`/`anyhow`; нет `.unwrap()`/`.expect()` в production-путях.
- **Concurrency:** `tokio`/`async`; отмена через drop/`CancellationToken`; shared state через `Arc<Mutex<...>>`/каналы.
- **Env/config:** `std::env::var` с обработкой `Result`; централизованный config (`serde`/`config`).
- **Logging:** `tracing`/`log` со структурой; нет `println!` в production.
- **Null-safety:** `Option<T>` + сопоставление с образцом; компилятор гарантирует отсутствие null.
- **Deps:** crates экосистемы (`serde`, `itertools`, `tokio`); stdlib-итераторы.

## 4. Anti-patterns
- `.unwrap()`/`.expect()` на пути, где ошибка возможна.
- `println!`/`eprintln!` для логирования в production.
- `unsafe` без обоснования.
- Игнор `Result` (`let _ = ...`) там, где ошибку нужно обработать.

## 5. Check-ID hints
- `LOG-01` → `println!`/`eprintln!` вместо `tracing`/`log`.
- `ERR-01` → `.unwrap()`/`.expect()`/проигнорированный `Result`.
- `BUG-03` → в основном **N/A** (нет null; `Option` проверяется компилятором).
- `BUG-10` → ReDoS обычно **N/A** (crate `regex` без catastrophic backtracking).
- Прочие стек-специфичные → при нехватке evidence `🔍 UNVERIFIED`.
