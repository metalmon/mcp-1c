# BSL-код HTTP-сервиса расширения 1С

Эндпоинты HTTP-сервиса `MCPService` для интеграции с MCP-сервером.

## Эндпоинты

| Эндпоинт | Метод | Файл | Описание |
|----------|-------|------|----------|
| `/metadata` | GET | [metadata.bsl](metadata.bsl) | Дерево метаданных конфигурации |
| `/object/{type}/{name}` | GET | [object.bsl](object.bsl) | Структура объекта метаданных |
| `/module/{type}/{name}/{kind}` | GET | [module.bsl](module.bsl) | Код модуля объекта |
| `/query` | POST | [query.bsl](query.bsl) | Выполнение запроса (только SELECT) |
| `/version` | GET | [version.bsl](version.bsl) | Версия расширения |
| `/search` | POST | [search.bsl](search.bsl) | Поиск по коду модулей |
| `/form/{type}/{name}` | GET | [form-structure.bsl](form-structure.bsl) | Структура формы объекта |
| `/validate-query` | POST | [validate-query.bsl](validate-query.bsl) | Валидация синтаксиса запроса |

## Установка

1. Создайте расширение конфигурации в 1С:Предприятие
2. Добавьте HTTP-сервис `MCPService` с корневым URL `/mcp`
3. Для каждого эндпоинта создайте шаблон URL и обработчик
4. Скопируйте BSL-код из соответствующего файла

## Аутентификация

HTTP-сервис использует базовую аутентификацию 1С (Basic Auth). Укажите логин и пароль пользователя 1С в настройках MCP-сервера.

## Формат ответов

Все эндпоинты возвращают JSON с Content-Type: application/json; charset=utf-8.
