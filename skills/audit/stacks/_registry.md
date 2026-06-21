# Stack Profiles — реестр

Таблица соответствия «маркер-файл → id профиля». Используется блоком
Runtime Detection (см. `../runtime-detect.md`). Порядок проверки — сверху вниз,
выбирается первый совпавший рантайм (один запуск = один рантайм).

| Приоритет | Маркер-файл(ы) | runtime id | Профиль | Tier |
|-----------|----------------|------------|---------|------|
| 1 | `package.json` | `node` | `node.md` | first-class |
| 2 | `go.mod` | `go` | `go.md` | first-class |
| 3 | `pyproject.toml` / `requirements.txt` / `setup.py` | `python` | `python.md` | general |
| 4 | `Cargo.toml` | `rust` | `rust.md` | general |
| 5 | `pom.xml` / `build.gradle*` / `settings.gradle*` | `java` | `java.md` | general |
| — | ничего из перечисленного | `generic` | `_generic.md` | fallback |

## Tier — что означает

- **first-class** — профиль содержит конкретные инструменты и идиомы; находки
  могут быть `❌ FAIL` с точным evidence.
- **general** — профиль даёт нейтральные формулировки и общие идиомы, но без
  гарантированных инструментов; стек-специфичные находки без однозначного
  evidence помечай `🔍 UNVERIFIED`.
- **fallback** (`generic`) — инструментов нет; работают только стек-нейтральные
  проверки (docs-ссылки, secrets, naming-читаемость), остальное → `🔍 UNVERIFIED`.

## Категории инструментов (общие для всех профилей)

Секция «Tooling by category» в каждом профиле использует один и тот же набор
ключей, чтобы скиллы ссылались на категорию, а не на команду:

`unused-code`, `clone-detection`, `dep-audit`, `env-extraction`, `arch-lint`,
`lint/format`, `type-check`, `test-run`, `secret-scan`.

## Структура профиля

Каждый `stacks/<id>.md` имеет одинаковые секции:
1. **Detection signals** — маркер-файлы.
2. **Tooling by category** — таблица по категориям выше.
3. **Idioms** — как выглядит «правильно» (ожидания PASS).
4. **Anti-patterns** — как выглядит FAIL (1 строка кода на пункт).
5. **Check-ID hints** — точечные подсказки по префиксу Check ID.

## Добавление нового стека

1. Создай `stacks/<id>.md` по структуре выше.
2. Добавь строку в таблицу реестра.
3. Добавь ветку в bash-детект в `../runtime-detect.md` и в инлайн-блоках скиллов.
