# BSL-код HTTP-сервиса расширения 1С

Эндпоинты HTTP-сервиса `MCPService` для интеграции с MCP-сервером.

## Эндпоинты

| Эндпоинт | Метод | Файл | Описание |
|----------|-------|------|----------|
| `/metadata` | GET | [metadata.bsl](metadata.bsl) | Дерево метаданных конфигурации |
| `/subsystems` | GET | [subsystems.bsl](subsystems.bsl) | Дерево подсистем и полные имена объектов конфигурации |
| `/object/{type}/{name}` | GET | [object.bsl](object.bsl) | Структура объекта метаданных |
| `/query` | POST | [query.bsl](query.bsl) | Выполнение запроса (только SELECT) |
| `/version` | GET | [version.bsl](version.bsl) | Версия расширения |
| `/form/{type}/{name}` | GET | [form-structure.bsl](form-structure.bsl) | Структура формы объекта |
| `/validate-query` | POST | [validate-query.bsl](validate-query.bsl) | Валидация синтаксиса запроса |
| `/eventlog` | POST | [eventlog.bsl](eventlog.bsl) | Чтение журнала регистрации |
| `/configuration` | GET | [configuration.bsl](configuration.bsl) | Информация о конфигурации и платформе |
| `/extensions` | GET | [extensions.bsl](extensions.bsl) | Бинарные данные (.cfe) всех расширений в base64 |
| `/counterparty` | POST | [counterparty.bsl](counterparty.bsl) | Создание контрагента |
| `/counterparties` | GET | [counterparties.bsl](counterparties.bsl) | Чтение контрагентов |
| `/nomenclature` | POST | [nomenclature.bsl](nomenclature.bsl) | Создание номенклатуры |
| `/nomenclatures` | GET | [nomenclatures.bsl](nomenclatures.bsl) | Чтение номенклатуры |
| `/organizations` | GET | [organizations.bsl](organizations.bsl) | Чтение организаций |
| `/contracts` | GET | [contracts.bsl](contracts.bsl) | Чтение договоров |
| `/sales-invoice` | POST | [sales-invoice.bsl](sales-invoice.bsl) | Создание и обновление счёта покупателю |
| `/sales-invoices` | GET | [sales-invoices.bsl](sales-invoices.bsl) | Чтение счетов покупателю |
| `/sales-document` | POST | [sales-document.bsl](sales-document.bsl) | Создание и обновление реализации |
| `/incoming-goods` | POST | [incoming-goods.bsl](incoming-goods.bsl) | Создание поступления товаров и услуг |
| `/sales-documents` | GET | [sales-documents.bsl](sales-documents.bsl) | Чтение документов реализации |
| `/document-attachment` | POST | [document-attachment.bsl](document-attachment.bsl) | Присоединение файла к документу |
| `/document-attachments` | GET | [document-attachments.bsl](document-attachments.bsl) | Чтение вложений документа |
| `/document-attachment-content` | GET | [document-attachment-content.bsl](document-attachment-content.bsl) | Чтение содержимого вложения |
| `/nomenclature/recent_for_counterparty` | GET | [nomenclature-recent-for-counterparty.bsl](nomenclature-recent-for-counterparty.bsl) | Недавняя номенклатура поставщика |
| `/refs/validate` | POST | [refs-validate.bsl](refs-validate.bsl) | Проверка типизированных ссылок |
| `/document-attachment-metadata` | POST | [document-attachment-metadata.bsl](document-attachment-metadata.bsl) | Обновление реквизитов вложения |

## Установка

1. Создайте расширение конфигурации в 1С:Предприятие
2. Добавьте HTTP-сервис `MCPService` с корневым URL `/mcp`
3. Для каждого эндпоинта создайте шаблон URL и обработчик
4. Скопируйте BSL-код из соответствующего файла
5. Скопируйте вспомогательные функции, которые названы в шапке каждого файла
   (`РазобратьJSON`, `СформироватьJSON`, `ОтветJSON`, `ОтветОшибка`,
   `НайтиМетаданныеПоТипуИИмени` и остальные): они лежат в модуле HTTP-сервиса
   и без них обработчик не заработает

Все эндпоинты из таблицы нужны целиком: сервер вызывает каждый из них, и версия,
которую вернёт `/version`, о неполной установке ничего не скажет. Готовое
расширение поставляется вместе с сервером, и `mcp-1c --install` устанавливает
его целиком; сборка вручную по этим файлам нужна тогда, когда расширение
встраивают в уже существующее.

## Аутентификация

HTTP-сервис использует базовую аутентификацию 1С (Basic Auth). Укажите логин и пароль пользователя 1С в настройках MCP-сервера.

## Формат ответов

Все эндпоинты возвращают JSON с Content-Type: application/json; charset=utf-8.
