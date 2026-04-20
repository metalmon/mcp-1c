.PHONY: build test clean release mock install run-1c-dev run-mcp run-mcp-dev run-mcp-business run-mcp-dump run-mcp-reindex check-http dump-config cursor-install help
.DEFAULT_GOAL := help

VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)
USER_HOME ?= $(if $(USERPROFILE),$(USERPROFILE),$(HOME))
INFOBASE_PATH ?= $(USER_HOME)/Documents/InfoBase
DUMP_DIR ?= $(CURDIR)/../dump-config
HTTP_PORT ?= 80
BASE_URL ?= http://localhost:$(HTTP_PORT)/hs/mcp-1c
TOOLSET ?= all
PROFILE ?= auto
MCP_USER ?=
MCP_PASSWORD ?=
PLATFORM_EXE ?= 1cv8.exe
ENTERPRISE_EXE ?= 1cv8.exe
DB_USER ?=
DB_PASSWORD ?=
CURSOR_MCP_FILE ?= $(CURDIR)/../.cursor/mcp.json
CURSOR_SERVER_NAME ?= 1c-runtime
CURSOR_TOOLSET ?= $(TOOLSET)
CURSOR_PROFILE ?= $(PROFILE)

ifeq ($(OS),Windows_NT)
MCP_BIN := bin/mcp-1c.exe
ROOT_BIN := mcp-1c.exe
MCP_RUN := .\$(ROOT_BIN)
else
MCP_BIN := bin/mcp-1c
ROOT_BIN := mcp-1c
MCP_RUN := ./$(ROOT_BIN)
endif

build:
	go build -ldflags "$(LDFLAGS)" -o $(MCP_BIN) ./cmd/mcp-1c
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -Command "Copy-Item -Path '$(MCP_BIN)' -Destination '$(ROOT_BIN)' -Force"
else
	cp $(MCP_BIN) $(ROOT_BIN)
endif

test:
	go test ./... -v -race

clean:
ifeq ($(OS),Windows_NT)
	@powershell -NoProfile -ExecutionPolicy Bypass -Command "if (Test-Path 'bin') { Remove-Item -Recurse -Force 'bin' }; if (Test-Path 'dist') { Remove-Item -Recurse -Force 'dist' }"
	@powershell -NoProfile -ExecutionPolicy Bypass -Command "if (Test-Path '$(ROOT_BIN)') { Remove-Item -Force '$(ROOT_BIN)' }"
else
	rm -rf bin/ dist/ $(ROOT_BIN)
endif

release:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/mcp-1c-windows-amd64.exe ./cmd/mcp-1c
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/mcp-1c-windows-arm64.exe ./cmd/mcp-1c
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/mcp-1c-linux-amd64 ./cmd/mcp-1c
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/mcp-1c-linux-arm64 ./cmd/mcp-1c
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/mcp-1c-darwin-amd64 ./cmd/mcp-1c
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/mcp-1c-darwin-arm64 ./cmd/mcp-1c

mock:
	go run ./cmd/mock-1c

install:
	$(MCP_RUN) --install "$(INFOBASE_PATH)" --db-user "$(DB_USER)" --db-password "$(DB_PASSWORD)"

run-1c-dev:
	powershell -NoProfile -ExecutionPolicy Bypass -Command "$$exe='$(ENTERPRISE_EXE)'; if (-not (Test-Path $$exe)) { $$cmd=Get-Command $$exe -ErrorAction SilentlyContinue; if ($$cmd) { $$exe=$$cmd.Source } }; if (-not (Test-Path $$exe)) { $$paths=@('C:\Program Files\1cv8','C:\Program Files\1cv8t','C:\Program Files (x86)\1cv8','C:\Program Files (x86)\1cv8t'); $$existing=$$paths | Where-Object { Test-Path $$_ }; if ($$existing) { $$all=Get-ChildItem -Path $$existing -Filter 1cv8*.exe -File -Recurse -ErrorAction SilentlyContinue | Where-Object { $$_.Name -in @('1cv8.exe','1cv8t.exe') } | Sort-Object FullName -Descending; if ($$all) { $$exe=($$all | Select-Object -First 1).FullName } } }; if (-not (Test-Path $$exe)) { throw '1C Enterprise executable not found. Set ENTERPRISE_EXE explicitly, e.g.: make run-1c-dev ENTERPRISE_EXE=C:/Program Files/1cv8/8.3.xx.xxxx/bin/1cv8.exe' }; Write-Output ('Starting 1C Enterprise: ' + $$exe); & $$exe ENTERPRISE /F '$(INFOBASE_PATH)' /DisableStartupDialogs /DisableStartupMessages /HTTPPort $(HTTP_PORT)"

run-mcp: build
	$(MCP_RUN) --base "$(BASE_URL)" --toolset "$(TOOLSET)" --profile "$(PROFILE)"

run-mcp-dev: build
	$(MCP_RUN) --base "$(BASE_URL)" --toolset developer --profile auto

run-mcp-business: build
	$(MCP_RUN) --base "$(BASE_URL)" --toolset business --profile auto

run-mcp-dump: build
	$(MCP_RUN) --base "$(BASE_URL)" --dump "$(DUMP_DIR)" --toolset "$(TOOLSET)" --profile "$(PROFILE)"

run-mcp-reindex: build
	$(MCP_RUN) --base "$(BASE_URL)" --dump "$(DUMP_DIR)" --reindex --toolset "$(TOOLSET)" --profile "$(PROFILE)"

check-http:
	curl "$(BASE_URL)/version"

dump-config:
	powershell -NoProfile -ExecutionPolicy Bypass -Command "$$exe='$(PLATFORM_EXE)'; if (-not (Test-Path $$exe)) { $$cmd=Get-Command $$exe -ErrorAction SilentlyContinue; if ($$cmd) { $$exe=$$cmd.Source } }; if (-not (Test-Path $$exe)) { $$paths=@('C:\Program Files\1cv8','C:\Program Files\1cv8t','C:\Program Files (x86)\1cv8','C:\Program Files (x86)\1cv8t'); $$existing=$$paths | Where-Object { Test-Path $$_ }; if ($$existing) { $$all=Get-ChildItem -Path $$existing -Filter 1cv8*.exe -File -Recurse -ErrorAction SilentlyContinue | Where-Object { $$_.Name -in @('1cv8.exe','1cv8t.exe') } | Sort-Object FullName -Descending; if ($$all) { $$exe=($$all | Select-Object -First 1).FullName } } }; if (-not (Test-Path $$exe)) { throw '1C DESIGNER executable not found. Set PLATFORM_EXE explicitly, e.g.: make dump-config PLATFORM_EXE=C:/Program Files/1cv8/8.3.xx.xxxx/bin/1cv8.exe' }; New-Item -ItemType Directory -Path '$(DUMP_DIR)' -Force | Out-Null; $$log=Join-Path '$(DUMP_DIR)' 'dump.log'; $$result=Join-Path '$(DUMP_DIR)' 'dump-result.txt'; $$cfg=Join-Path '$(DUMP_DIR)' 'Configuration.xml'; if (Test-Path $$cfg) { Remove-Item -Force $$cfg }; if (Test-Path $$result) { Remove-Item -Force $$result }; & $$exe DESIGNER /F '$(INFOBASE_PATH)' /DisableStartupDialogs /DisableStartupMessages /DumpConfigToFiles '$(DUMP_DIR)' /Out $$log /DumpResult $$result; if (Test-Path $$result) { $$code=(Get-Content -Raw $$result).Trim(); if ($$code -ne '0') { throw ('DumpResult=' + $$code + '. Check log: ' + $$log) }; if (Test-Path $$cfg) { Write-Output ('Dump complete: $(DUMP_DIR)') } else { Write-Output ('Dump finished without Configuration.xml yet. Check log: ' + $$log) } } else { Write-Output ('Dump started asynchronously. Check progress in: ' + $$log) }"

cursor-install:
	powershell -NoProfile -ExecutionPolicy Bypass -Command "$$path='$(CURSOR_MCP_FILE)'; $$dir=Split-Path -Parent $$path; New-Item -ItemType Directory -Path $$dir -Force | Out-Null; if (Test-Path $$path) { $$raw=Get-Content -Raw -Path $$path; if ([string]::IsNullOrWhiteSpace($$raw)) { $$cfg=[pscustomobject]@{} } else { $$cfg=$$raw | ConvertFrom-Json } } else { $$cfg=[pscustomobject]@{} }; if ($$null -eq $$cfg) { $$cfg=[pscustomobject]@{} }; if (-not ($$cfg.PSObject.Properties.Name -contains 'mcpServers')) { $$cfg | Add-Member -NotePropertyName mcpServers -NotePropertyValue ([pscustomobject]@{}) }; $$args=@('--base','$(BASE_URL)','--toolset','$(CURSOR_TOOLSET)','--profile','$(CURSOR_PROFILE)'); if ('$(CURSOR_TOOLSET)' -ne 'business') { $$args += @('--dump','$(DUMP_DIR)') }; if ('$(MCP_USER)' -ne '') { $$args += @('--user','$(MCP_USER)') }; if ('$(MCP_PASSWORD)' -ne '') { $$args += @('--password','$(MCP_PASSWORD)') }; $$server=[pscustomobject]@{ command='$(CURDIR)/$(ROOT_BIN)'; args=$$args }; $$cfg.mcpServers | Add-Member -NotePropertyName '$(CURSOR_SERVER_NAME)' -NotePropertyValue $$server -Force; $$cfg | ConvertTo-Json -Depth 10 | Set-Content -Path $$path -Encoding UTF8; Write-Output ('Updated ' + $$path + ' with server $(CURSOR_SERVER_NAME)')"

help:
	@echo mcp-1c Makefile help
	@echo Build/Test:
	@echo   make build              - build $(MCP_BIN) and copy runnable $(ROOT_BIN) to repo root
	@echo   make test               - run unit and integration tests
	@echo   make clean              - remove build artifacts (bin/, dist/)
	@echo   make release            - build release binaries for all target OS/arch
	@echo   make mock               - run local mock 1C HTTP server
	@echo 1C local development flow:
	@echo   make run-1c-dev          - start 1C Enterprise with /HTTPPort $(HTTP_PORT) for local HTTP service
	@echo   make install            - install MCP extension into INFOBASE_PATH
	@echo   make dump-config        - dump InfoBase metadata to DUMP_DIR
	@echo   make run-mcp            - run MCP against BASE_URL
	@echo   make run-mcp-dev        - run MCP with developer tools only
	@echo   make run-mcp-business   - run MCP with business tools only (profile auto)
	@echo   make run-mcp-dump       - run MCP with local dump index (--dump)
	@echo   make run-mcp-reindex    - run MCP with forced index rebuild (--reindex)
	@echo   make check-http         - check 1C HTTP service version endpoint
	@echo   make cursor-install     - add/update this MCP server in .cursor/mcp.json
	@echo Examples:
	@echo   make install INFOBASE_PATH=C:/Users/username/Documents/InfoBase
	@echo   make install DB_USER=mcp DB_PASSWORD=secret
	@echo   make run-1c-dev ENTERPRISE_EXE=C:/Program Files/1cv8/8.3.xx.xxxx/bin/1cv8.exe
	@echo   make dump-config PLATFORM_EXE=C:/Program Files/1cv8/8.3.xx.xxxx/bin/1cv8.exe
	@echo   make run-mcp BASE_URL=http://localhost:80/hs/mcp-1c
	@echo   make run-mcp TOOLSET=developer PROFILE=auto
	@echo   make run-mcp-dump TOOLSET=business PROFILE=buh_3_0
	@echo   make run-mcp-dump DUMP_DIR=$(CURDIR)/../dump-config
	@echo   make cursor-install CURSOR_SERVER_NAME=1c-dev CURSOR_TOOLSET=developer CURSOR_PROFILE=auto
	@echo   make cursor-install CURSOR_SERVER_NAME=1c-business CURSOR_TOOLSET=business CURSOR_PROFILE=auto  (без --dump)
	@echo   make cursor-install CURSOR_SERVER_NAME=1c-business MCP_USER=mcp MCP_PASSWORD=secret
	@echo   make cursor-install CURSOR_MCP_FILE=$(CURDIR)/../.cursor/mcp.json
	@echo Current defaults:
	@echo   INFOBASE_PATH=$(INFOBASE_PATH)
	@echo   DUMP_DIR=$(DUMP_DIR)
	@echo   BASE_URL=$(BASE_URL)
	@echo   TOOLSET=$(TOOLSET)
	@echo   PROFILE=$(PROFILE)
	@echo   MCP_USER=$(MCP_USER)
	@echo   MCP_PASSWORD=$(if $(MCP_PASSWORD),***,)
	@echo   PLATFORM_EXE=$(PLATFORM_EXE)
	@echo   ENTERPRISE_EXE=$(ENTERPRISE_EXE)
	@echo   DB_USER=$(DB_USER)
	@echo   DB_PASSWORD=$(if $(DB_PASSWORD),***,)
	@echo   (dump-config auto-detects 1cv8.exe/1cv8c.exe in standard folders if PLATFORM_EXE not found in PATH)
	@echo   MCP_BIN=$(MCP_BIN)
	@echo   ROOT_BIN=$(ROOT_BIN)
	@echo   MCP_RUN=$(MCP_RUN)
	@echo   CURSOR_MCP_FILE=$(CURSOR_MCP_FILE)
	@echo   CURSOR_SERVER_NAME=$(CURSOR_SERVER_NAME)
	@echo   CURSOR_TOOLSET=$(CURSOR_TOOLSET)
	@echo   CURSOR_PROFILE=$(CURSOR_PROFILE)
