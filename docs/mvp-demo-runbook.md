# MVP demo runbook (BP 3.0)

Этот runbook описывает минимальный сценарий демонстрации MCP-агента в БП 3.0:
анализ данных + создание документов продаж.

## Профильный запуск (родной сценарий)

Рекомендуемая схема — два отдельных MCP-инстанса:

1. **business-buh30** (рабочий сценарий продаж)
   - `--profile buh_3_0`
   - write tools доступны при `--access read_write`
2. **business-generic** (безопасный fallback)
   - `--profile generic`
   - business endpoints возвращают controlled `not implemented`

Примеры:

- установка расширения БП-профиля: `make install-buh30`
- установка generic-профиля: `make install-generic`
- запуск БП-инстанса: `make run-mcp-business-buh30`
- запуск generic-инстанса: `make run-mcp-business-generic`

## Предусловия

- Запущен `mcp-1c` с `--toolset business --profile buh_3_0` (или `generic`).
- Расширение установлено в базу и HTTP-сервис доступен.
- Контракт `ref` во всех business tools: UUID.

## Шаг 1. Проверить справочные данные

1. Найти контрагента:
   - `read_counterparties` (`search` или `inn/kpp`).
2. Найти организацию:
   - `read_organizations` (`search` или `inn/kpp`).
3. Найти договор:
   - `read_contracts` (`counterparty_ref`, `organization_ref`, `search/number`).
4. Найти номенклатуру:
   - `read_nomenclature` (`search` или `article`).

## Шаг 2. Fallback для пустой базы

Если данных не хватает:

1. Создать контрагента:
   - `create_counterparty`.
2. Создать номенклатуру:
   - `create_nomenclature`.
3. Повторить чтение организаций/договоров.
4. Если договора нет, создать вручную в 1С UI (в текущем MVP create_contract не реализован).

## Шаг 3. Создать счет покупателю

Вызвать `create_sales_invoice`:

- обязательные поля: `organization_ref`, `counterparty_ref`;
- опционально: `contract_ref`, `comment`, `items[]`;
- флаг `post`:
  - `false` — записать черновик;
  - `true` — записать и провести.

Сразу проверить:

- `read_sales_invoices` по `ref` из результата `create_sales_invoice`.

## Шаг 4. Создать реализацию

Вызвать `create_sales_document`:

- обязательные поля: `organization_ref`, `counterparty_ref`;
- опционально: `contract_ref`, `invoice_ref`, `comment`, `items[]`;
- флаг `post`:
  - `false` — записать черновик;
  - `true` — записать и провести.

Сразу проверить:

- `read_sales_documents` по `ref` из результата `create_sales_document`.

## Шаг 5. Проверка корректности `ref`-контракта

- Любой `ref` из `create_*` должен читаться через `read_*(ref=...)`.
- Невалидный `ref` должен возвращать `400` (`...must be valid UUID`).

## Рекомендации для демо

- Для первого прогона использовать `post=false`, чтобы избежать блокировок проведения на пустой базе.
- Затем повторить с `post=true`, когда базовые справочники заполнены.
- Все ключевые шаги фиксировать по `ref` (а не по номеру), чтобы демонстрация была детерминированной.
