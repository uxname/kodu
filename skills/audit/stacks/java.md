# Stack Profile: Java / JVM   (id: java)
Tier: general

A general-tier profile: neutral idioms + tooling pointers without
guarantees they are present. Mark stack-specific findings without unambiguous evidence
as `🔍 UNVERIFIED`.

## 1. Detection signals
- `pom.xml` (Maven) / `build.gradle*` / `settings.gradle*` (Gradle)

## 2. Tooling by category
| Category | Command | How to read the output |
|-----------|---------|------------------|
| unused-code | SpotBugs / IDE inspections | unused → YAGNI-02 (often `🔍 UNVERIFIED` without a tool) |
| clone-detection | `pmd cpd --minimum-tokens 50 --dir . 2>/dev/null \|\| true` | duplicates → REINV-03 |
| dep-audit | `mvn org.owasp:dependency-check-maven:check 2>/dev/null \|\| true` | CVEs in dependencies |
| env-extraction | `grep -rEoh 'System\.getenv\("[A-Z0-9_]+"\)' . 2>/dev/null \| sort -u` | env from code → DOC-02 (account for `@Value`/`application.yml`) |
| arch-lint | ArchUnit tests | layers/dependencies |
| lint/format | `mvn checkstyle:check 2>/dev/null \|\| true`; SpotBugs | — |
| type-check | `mvn compile 2>/dev/null \|\| ./gradlew compileJava 2>/dev/null \|\| true` | compilation |
| test-run | `mvn test 2>/dev/null \|\| ./gradlew test 2>/dev/null \|\| true` | — |
| secret-scan | `gitleaks detect --no-banner 2>/dev/null \|\| true` | stack-neutral |

## 3. Idioms
- **Errors:** specific exceptions; try-with-resources for resources; no empty `catch`.
- **Concurrency:** `ExecutorService`/`CompletableFuture`; cancellation via `Future.cancel`/interruption; immutability or `synchronized`/`java.util.concurrent`.
- **Env/config:** Spring `@Value`/`application.yml` or `System.getenv`, centralized.
- **Logging:** SLF4J/Logback with structure; no `System.out.println` in production.
- **Null-safety:** `Optional<T>`; `@Nullable`/`@NonNull` annotations; null checks.
- **Lifecycle:** `@Transactional` for read-modify-write; graceful shutdown via lifecycle hooks.
- **Deps:** proven ecosystem libraries instead of hand-written code.

## 4. Anti-patterns
- Empty `catch (Exception e) {}`.
- `System.out.println`/`printStackTrace` in production.
- Shared mutable state without synchronization.
- Scattered `System.getenv` without centralization.

## 5. Check-ID hints
- `LOG-01` → `System.out.println`/`printStackTrace` instead of SLF4J.
- `ERR-01` → empty `catch`.
- `CON-02` → read-modify-write without `@Transactional`.
- `ARC-05` → scattered `System.getenv`.
- Other stack-specific → when evidence is lacking, `🔍 UNVERIFIED`.
