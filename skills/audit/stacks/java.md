# Stack Profile: Java / JVM   (id: java)
Tier: general

Профиль уровня general: нейтральные идиомы + ориентиры по инструментам без
гарантий их наличия. Стек-специфичные находки без однозначного evidence
помечай `🔍 UNVERIFIED`.

## 1. Detection signals
- `pom.xml` (Maven) / `build.gradle*` / `settings.gradle*` (Gradle)

## 2. Tooling by category
| Категория | Команда | Как читать вывод |
|-----------|---------|------------------|
| unused-code | SpotBugs / IDE-инспекции | неиспользуемое → YAGNI-02 (часто `🔍 UNVERIFIED` без инструмента) |
| clone-detection | `pmd cpd --minimum-tokens 50 --dir . 2>/dev/null \|\| true` | дубли → REINV-03 |
| dep-audit | `mvn org.owasp:dependency-check-maven:check 2>/dev/null \|\| true` | CVE в зависимостях |
| env-extraction | `grep -rEoh 'System\.getenv\("[A-Z0-9_]+"\)' . 2>/dev/null \| sort -u` | env из кода → DOC-02 (учти `@Value`/`application.yml`) |
| arch-lint | ArchUnit-тесты | слои/зависимости |
| lint/format | `mvn checkstyle:check 2>/dev/null \|\| true`; SpotBugs | — |
| type-check | `mvn compile 2>/dev/null \|\| ./gradlew compileJava 2>/dev/null \|\| true` | компиляция |
| test-run | `mvn test 2>/dev/null \|\| ./gradlew test 2>/dev/null \|\| true` | — |
| secret-scan | `gitleaks detect --no-banner 2>/dev/null \|\| true` | стек-нейтрально |

## 3. Idioms
- **Errors:** конкретные исключения; try-with-resources для ресурсов; нет пустых `catch`.
- **Concurrency:** `ExecutorService`/`CompletableFuture`; отмена через `Future.cancel`/interruption; иммутабельность или `synchronized`/`java.util.concurrent`.
- **Env/config:** Spring `@Value`/`application.yml` или `System.getenv`, централизованно.
- **Logging:** SLF4J/Logback со структурой; нет `System.out.println` в production.
- **Null-safety:** `Optional<T>`; аннотации `@Nullable`/`@NonNull`; проверки null.
- **Lifecycle:** `@Transactional` для read-modify-write; graceful shutdown через lifecycle-хуки.
- **Deps:** проверенные библиотеки экосистемы вместо самописного.

## 4. Anti-patterns
- Пустой `catch (Exception e) {}`.
- `System.out.println`/`printStackTrace` в production.
- Shared mutable state без синхронизации.
- Разбросанный `System.getenv` без централизации.

## 5. Check-ID hints
- `LOG-01` → `System.out.println`/`printStackTrace` вместо SLF4J.
- `ERR-01` → пустой `catch`.
- `CON-02` → read-modify-write без `@Transactional`.
- `ARC-05` → `System.getenv` вразброс.
- Прочие стек-специфичные → при нехватке evidence `🔍 UNVERIFIED`.
