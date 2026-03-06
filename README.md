# mcp-1c

MCP-сервер для интеграции AI-ассистентов (Claude, Cursor) с 1С:Предприятие.

AI видит метаданные вашей конфигурации 1С и генерирует точный код на BSL.

## Как это работает

```
Claude Desktop / Cursor
        │ stdio (JSON-RPC 2.0)
        ▼
┌──────────────────┐
│  MCP-сервер (Go) │  ← этот репозиторий
│  один бинарник   │
└────────┬─────────┘
         │ HTTP (localhost)
         ▼
┌──────────────────┐
│ Расширение 1С    │  ← файл extension/MCP_HTTPService.cfe
│ (HTTP-сервис)    │
│ в базе клиента   │
└──────────────────┘
```

## Установка

### Автоматическая (Windows)

```cmd
git clone https://github.com/Igorut/mcp-1c.git
cd mcp-1c
scripts\setup.cmd
```

Скрипт сам найдёт 1С (коммерческую или учебную), установит расширение в базу и покажет как настроить Claude Desktop.

### Ручная (пошагово)

#### Шаг 1. Скачать MCP-сервер

Скачайте бинарник для вашей ОС из [Releases](../../releases).

Или соберите из исходников:
```bash
go build -o mcp-1c.exe ./cmd/mcp-1c/
```

#### Шаг 2. Установить расширение в 1С

В репозитории лежит готовый файл расширения `extension/MCP_HTTPService.cfe`.

1. Откройте вашу базу 1С в **Конфигураторе**
2. **Конфигурация → Расширения конфигурации**
3. Нажмите **"Добавить из файла"**
4. Выберите `MCP_HTTPService.cfe`
5. Нажмите **F7** (обновить конфигурацию базы данных)

Расширение добавляет HTTP-сервис `MCPService` с двумя эндпоинтами:
- `/mcp/metadata` — дерево метаданных конфигурации
- `/mcp/object/{type}/{name}` — структура конкретного объекта

Через командную строку:
```cmd
:: Коммерческая версия
"C:\Program Files\1cv8\8.3.xx.xxxx\bin\1cv8.exe" DESIGNER /F "C:\путь\к\базе" /LoadCfg "extension\MCP_HTTPService.cfe" /Extension "MCP_HTTPService" /UpdateDBCfg

:: Учебная версия
"C:\Program Files\1cv8t\8.5.xx.xxxx\bin\1cv8t.exe" DESIGNER /F "C:\путь\к\базе" /LoadCfg "extension\MCP_HTTPService.cfe" /Extension "MCP_HTTPService" /UpdateDBCfg
```

#### Шаг 3. Запустить HTTP-сервис 1С

**Вариант A — Встроенный веб-сервер (просто):**
```cmd
:: Коммерческая
"C:\Program Files\1cv8\8.3.xx.xxxx\bin\1cv8.exe" ENTERPRISE /F "C:\путь\к\базе" /HTTPPort 8080

:: Учебная
"C:\Program Files\1cv8t\8.5.xx.xxxx\bin\1cv8t.exe" ENTERPRISE /F "C:\путь\к\базе" /HTTPPort 8080
```

**Вариант B — Через IIS / Apache:**

Откройте Конфигуратор **от администратора** → Администрирование → Публикация на веб-сервере → укажите имя `base` → Опубликовать.

**Проверка:**
```bash
# Встроенный сервер
curl http://localhost:8080/hs/mcp/metadata

# IIS/Apache
curl http://localhost/base/hs/mcp/metadata
```

#### Шаг 4. Настроить Claude Desktop

Добавьте в файл конфигурации:

- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`
- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "1c": {
      "command": "C:\\путь\\к\\mcp-1c.exe",
      "args": ["--base", "http://localhost:8080/hs/mcp"]
    }
  }
}
```

> Если публиковали через IIS/Apache, замените URL на `http://localhost/base/hs/mcp`

Перезапустите Claude Desktop.

#### Шаг 5. Готово!

Спросите Claude: **"Покажи структуру конфигурации моей базы 1С"**

## Доступные инструменты

| Tool | Описание |
|------|---------|
| `get_metadata_tree` | Дерево метаданных: справочники, документы, регистры |
| `get_object_structure` | Реквизиты и табличные части конкретного объекта |
| `bsl_syntax_help` | Справка по встроенным функциям BSL |

## Конфигурация

| Флаг | Env var | По умолчанию | Описание |
|------|---------|-------------|----------|
| `--base` | `MCP_1C_BASE_URL` | `http://localhost:8080/mcp` | URL HTTP-сервиса 1С |
| `--user` | `MCP_1C_USER` | — | Пользователь HTTP-сервиса |
| `--password` | `MCP_1C_PASSWORD` | — | Пароль HTTP-сервиса |

Приоритет: CLI флаги > env vars > значения по умолчанию.

## Разработка

```bash
# Сборка
go build -o mcp-1c ./cmd/mcp-1c

# Тесты
go test ./... -v -race

# Mock-сервер 1С (для разработки без реальной 1С)
go run ./cmd/mock-1c -port 9191
```

## Совместимость

| Платформа 1С | Статус |
|-------------|--------|
| 8.3.x (коммерческая) | Поддерживается |
| 8.5.x (коммерческая) | Поддерживается |
| 8.3.x / 8.5.x (учебная) | Поддерживается (Windows) |

| ОС | MCP-сервер | HTTP-сервис 1С |
|----|-----------|----------------|
| Windows | да | да |
| macOS | да | нет (учебная версия) |
| Linux | да | да (через ibsrv) |

## Лицензия

MIT
