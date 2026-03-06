@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo  ══════════════════════════════════════════════
echo  ║  MCP для 1С — Полная установка             ║
echo  ║  github.com/Igorut/mcp-1c                  ║
echo  ══════════════════════════════════════════════
echo.

:: ─────────────────────────────────────────────────
:: 1. Поиск платформы 1С
:: ─────────────────────────────────────────────────
echo [1/5] Поиск платформы 1С...

set "PLATFORM_DIR="
set "PLATFORM_EXE="
set "PLATFORM_TYPE="

:: Коммерческая версия
for /d %%d in ("C:\Program Files\1cv8\8.*") do (
    if exist "%%d\bin\1cv8.exe" (
        set "PLATFORM_DIR=%%d"
        set "PLATFORM_EXE=%%d\bin\1cv8.exe"
        set "PLATFORM_TYPE=commercial"
    )
)
for /d %%d in ("C:\Program Files (x86)\1cv8\8.*") do (
    if exist "%%d\bin\1cv8.exe" (
        set "PLATFORM_DIR=%%d"
        set "PLATFORM_EXE=%%d\bin\1cv8.exe"
        set "PLATFORM_TYPE=commercial"
    )
)

:: Учебная версия (1cv8t)
if not defined PLATFORM_DIR (
    for /d %%d in ("C:\Program Files\1cv8t\8.*") do (
        if exist "%%d\bin\1cv8t.exe" (
            set "PLATFORM_DIR=%%d"
            set "PLATFORM_EXE=%%d\bin\1cv8t.exe"
            set "PLATFORM_TYPE=educational"
        )
    )
    for /d %%d in ("C:\Program Files (x86)\1cv8t\8.*") do (
        if exist "%%d\bin\1cv8t.exe" (
            set "PLATFORM_DIR=%%d"
            set "PLATFORM_EXE=%%d\bin\1cv8t.exe"
            set "PLATFORM_TYPE=educational"
        )
    )
)

:: Учебная версия — альтернативные пути
if not defined PLATFORM_DIR (
    for /d %%d in ("C:\1cv8t\8.*") do (
        if exist "%%d\bin\1cv8t.exe" (
            set "PLATFORM_DIR=%%d"
            set "PLATFORM_EXE=%%d\bin\1cv8t.exe"
            set "PLATFORM_TYPE=educational"
        )
    )
)

if not defined PLATFORM_DIR (
    echo.
    echo  [!] Платформа 1С не найдена автоматически.
    echo.
    echo  Укажите путь к исполняемому файлу 1С вручную:
    echo  Примеры:
    echo    C:\Program Files\1cv8\8.3.25.1234\bin\1cv8.exe
    echo    C:\Program Files\1cv8t\8.5.1.1150\bin\1cv8t.exe
    echo.
    set /p "PLATFORM_EXE=Путь к 1cv8.exe или 1cv8t.exe: "
    if not exist "!PLATFORM_EXE!" (
        echo  [ОШИБКА] Файл не найден: !PLATFORM_EXE!
        pause
        exit /b 1
    )
    set "PLATFORM_TYPE=manual"
)

echo       Платформа: %PLATFORM_EXE%
if "%PLATFORM_TYPE%"=="educational" echo       Тип: учебная версия
if "%PLATFORM_TYPE%"=="commercial" echo       Тип: коммерческая версия
echo.

:: ─────────────────────────────────────────────────
:: 2. Путь к базе
:: ─────────────────────────────────────────────────
echo [2/5] Укажите путь к информационной базе 1С
echo.
echo  Это папка, в которой находится файл 1Cv8.1CD.
echo  Пример: C:\Users\User\Documents\InfoBase
echo.
set /p "BASE_PATH=Путь к базе: "

if not exist "%BASE_PATH%\1Cv8.1CD" (
    echo.
    echo  [ОШИБКА] Файл 1Cv8.1CD не найден в: %BASE_PATH%
    echo  Проверьте путь. Это должна быть папка с файлом базы данных.
    pause
    exit /b 1
)

echo       База: %BASE_PATH%
echo.

:: ─────────────────────────────────────────────────
:: 3. Установка расширения
:: ─────────────────────────────────────────────────
echo [3/5] Установка расширения MCP_HTTPService...

set "SCRIPT_DIR=%~dp0"
set "CFE_PATH=%SCRIPT_DIR%..\extension\MCP_HTTPService.cfe"

if not exist "%CFE_PATH%" (
    echo.
    echo  [ОШИБКА] Файл расширения не найден: %CFE_PATH%
    echo  Убедитесь, что папка extension содержит MCP_HTTPService.cfe
    pause
    exit /b 1
)

echo       Расширение: %CFE_PATH%
echo.
echo  Загрузка расширения в базу (это может занять 30-60 секунд)...
echo  Убедитесь, что база НЕ открыта в Конфигураторе или Предприятии!
echo.

"%PLATFORM_EXE%" DESIGNER /F "%BASE_PATH%" /LoadCfg "%CFE_PATH%" /Extension "MCP_HTTPService" /UpdateDBCfg

if %errorlevel% neq 0 (
    echo.
    echo  [ОШИБКА] Не удалось загрузить расширение (код: %errorlevel%)
    echo.
    echo  Возможные причины:
    echo    - База открыта в другом сеансе (закройте все окна 1С)
    echo    - Недостаточно прав (запустите скрипт от администратора)
    echo.
    echo  Альтернативный способ (вручную):
    echo    1. Откройте базу в Конфигураторе
    echo    2. Конфигурация → Расширения конфигурации
    echo    3. "Добавить из файла" → выберите %CFE_PATH%
    echo    4. Нажмите F7
    echo.
    pause
    exit /b 1
)

echo       Расширение установлено!
echo.

:: ─────────────────────────────────────────────────
:: 4. Проверка MCP-сервера
:: ─────────────────────────────────────────────────
echo [4/5] Проверка MCP-сервера...

set "MCP_EXE=%SCRIPT_DIR%..\mcp-1c.exe"
if not exist "%MCP_EXE%" (
    :: Проверяем в текущей папке
    set "MCP_EXE=%SCRIPT_DIR%..\..\mcp-1c.exe"
)
if not exist "%MCP_EXE%" (
    echo.
    echo  [!] mcp-1c.exe не найден.
    echo  Скачайте из Releases: https://github.com/Igorut/mcp-1c/releases
    echo  Или соберите: go build -o mcp-1c.exe ./cmd/mcp-1c/
    echo.
    echo  Пропускаем проверку, продолжаем настройку...
    set "MCP_EXE=mcp-1c.exe"
) else (
    echo       MCP-сервер: %MCP_EXE%
)
echo.

:: ─────────────────────────────────────────────────
:: 5. Выбор способа публикации и запуск
:: ─────────────────────────────────────────────────
echo [5/5] Публикация HTTP-сервиса
echo.
echo  Выберите способ:
echo    1 - Встроенный веб-сервер (рекомендуется, самый простой)
echo    2 - Только показать инструкцию для IIS/Apache
echo.
set /p "PUBLISH_MODE=Ваш выбор (1 или 2): "

if "%PUBLISH_MODE%"=="2" goto :show_manual_publish

:: Встроенный веб-сервер
set "HTTP_PORT=8080"
set /p "HTTP_PORT=Порт HTTP-сервиса (по умолчанию 8080): "

echo.
echo  ══════════════════════════════════════════════
echo  ║  УСТАНОВКА ЗАВЕРШЕНА!                      ║
echo  ══════════════════════════════════════════════
echo.
echo  Для запуска выполните ДВЕ команды в РАЗНЫХ терминалах:
echo.
echo  Терминал 1 — HTTP-сервис 1С:
echo    "%PLATFORM_EXE%" ENTERPRISE /F "%BASE_PATH%" /HTTPPort %HTTP_PORT%
echo.
echo  Терминал 2 — проверка:
echo    curl http://localhost:%HTTP_PORT%/hs/mcp/metadata
echo.
echo  ─────────────────────────────────────────────
echo  Настройка Claude Desktop
echo  ─────────────────────────────────────────────
echo.
echo  Добавьте в %%APPDATA%%\Claude\claude_desktop_config.json:
echo.
echo  {
echo    "mcpServers": {
echo      "1c": {
echo        "command": "%MCP_EXE:\=\\%",
echo        "args": ["--base", "http://localhost:%HTTP_PORT%/hs/mcp"]
echo      }
echo    }
echo  }
echo.
echo  Перезапустите Claude Desktop и спросите:
echo  "Покажи структуру конфигурации моей базы 1С"
echo.
pause
exit /b 0

:show_manual_publish
echo.
echo  ─────────────────────────────────────────────
echo  Публикация через IIS:
echo  ─────────────────────────────────────────────
echo  1. Включите IIS: Панель управления → Программы → Компоненты Windows → IIS
echo  2. Откройте базу в Конфигураторе ОТ АДМИНИСТРАТОРА
echo  3. Администрирование → Публикация на веб-сервере
echo  4. Имя: base, отметьте MCPService → Опубликовать
echo  5. URL: http://localhost/base/hs/mcp/metadata
echo.
echo  ─────────────────────────────────────────────
echo  Публикация через Apache:
echo  ─────────────────────────────────────────────
echo  1. Установите Apache 2.4 (https://www.apachelounge.com/download/)
echo  2. Откройте Конфигуратор ОТ АДМИНИСТРАТОРА
echo  3. Администрирование → Публикация на веб-сервере
echo  4. Выберите Apache 2.4, имя: base → Опубликовать
echo  5. URL: http://localhost/base/hs/mcp/metadata
echo.
pause
exit /b 0
