### 1. Выбор "Fresh & Modern" стека (Библиотеки)

Вместо старых монстров берем инструменты экосистемы UnJS, Sindre Sorhus и современные стандарты:

*   **File System:** `node:fs/promises` (нативный, мощный) + **`tinyglobby`** (вместо тяжелого `globby` или `fast-glob`, супер-быстрый паттерн-матчинг).
*   **CLI UI:** **`@inquirer/prompts`** (новый модульный инквайрер, тянет меньше зависимостей) + **`picocolors`** (в 10 раз быстрее и меньше `chalk`).
*   **Spinners:** **`yocto-spinner`** (современная, легкая замена `ora`).
*   **Process Execution:** **`execa`** (лучшая обертка над child_process для работы с Git).
*   **Validation:** **`zod`** (стандарт де-факто).
*   **Clipboard:** **`clipboardy`** (надежный кроссплатформ, альтернатив мало).
*   **AI/HTTP:** **`mastra`** (как договорились) + нативный `fetch`.

---

### 2. Архитектура NestJS (Module Map)

Разбиваем монолит на функциональные домены.

```text
src/
├── app.module.ts            # Корневой модуль (оркестратор)
│
├── core/                    # Global Modules (инфраструктура)
│   ├── config/              # ConfigModule (Zod + cosmiconfig)
│   ├── file-system/         # FsModule (tinyglobby, node:fs обертки)
│   └── ui/                  # UiModule (Спиннеры, цвета, промпты)
│
├── shared/                  # Общая бизнес-логика
│   ├── tokenizer/           # TokenizerModule (js-tiktoken)
│   ├── git/                 # GitModule (execa обертки над git)
│   └── ai/                  # AiModule (Mastra client setup)
│
└── commands/                # Feature Modules (команды)
    ├── init/                # InitModule (kodu init)
    ├── pack/                # PackModule (kodu pack)
    ├── clean/               # CleanModule (kodu clean /w ts-morph)
    ├── review/              # ReviewModule (kodu review)
    └── commit/              # CommitModule (kodu commit)
```

---

### 3. План разработки (Roadmap)

Разбиваем на 4 спринта (фазы).

#### Фаза 1: The Foundation (Фундамент)
*Цель: Приложение запускается, читает конфиг и умеет работать с файлами.*

1.  **Core Setup:**
    *   Настройка `ConfigModule`: чтение `kodu.json`, валидация через `Zod`.
    *   Настройка `UiModule`: сделать красивые хелперы для вывода (success, error, warning).
2.  **Command: Init:**
    *   Реализация `InitCommand`.
    *   Интерактивный опросник через `@inquirer/prompts`.
    *   Генерация дефолтного конфига.

#### Фаза 2: The Packer & Tokenizer (Локальная магия)
*Цель: Утилита полезна даже без интернета.*

1.  **File System Logic:**
    *   Интеграция `tinyglobby` для умного поиска файлов (учет `.gitignore` + `kodu.json`).
2.  **Tokenizer:**
    *   Сервис подсчета токенов на базе `js-tiktoken`.
3.  **Command: Pack:**
    *   Сборка контента.
    *   Реализация флага `--copy` (буфер обмена).
    *   Реализация Template Engine (простая подстановка шаблонов промптов).

#### Фаза 3: The Cleaner (AST Engineering)
*Цель: Умная очистка кода.*

1.  **Clean Logic:**
    *   Подключение `ts-morph`.
    *   Реализация алгоритма обхода AST.
    *   Реализация whitelist (System + Biome + User).
    *   Особое внимание удалению JSX-комментариев `{/* ... */}`.
2.  **Command: Clean:**
    *   Интеграция логики в команду.
    *   Добавление флага `--dry-run` (чтобы показать, что удалится, не удаляя).

#### Фаза 4: AI Integration (Мозги)
*Цель: Подключение Mastra и Git.*

1.  **Git Integration:**
    *   Сервис на `execa` для получения `git diff --staged`.
2.  **Mastra Agent:**
    *   Настройка агента в `AiModule`.
3.  **Command: Review:**
    *   Пайплайн: Git Diff -> Tokenizer Check -> Mastra -> Report.
4.  **Command: Commit:**
    *   Пайплайн: Git Diff -> Mastra -> Generate Msg -> Inquirer Confirm -> Git Commit.
