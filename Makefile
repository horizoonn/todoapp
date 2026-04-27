-include .env
export

.DEFAULT_GOAL := help

export PROJECT_ROOT=${shell pwd}
ENV_SERVICES=todoapp-postgres todoapp-redis
MOCKGEN_VERSION=v0.6.0
COVERAGE_DIR=${PROJECT_ROOT}/.out/coverage
COVERAGE_PROFILE=${COVERAGE_DIR}/coverage.out
COVERAGE_HTML=${COVERAGE_DIR}/coverage.html
INTEGRATION_COVERAGE_PROFILE=${COVERAGE_DIR}/integration_coverage.out
INTEGRATION_COVERAGE_HTML=${COVERAGE_DIR}/integration_coverage.html

.PHONY: help check check-all fmt fmt-check vet \
	env-up env-down env-cleanup \
	migrate-create migrate-up migrate-down migrate-action \
	logs-cleanup \
	test test-unit test-all test-cover test-cover-profile test-cover-func test-cover-html test-race \
	test-integration test-integration-cover test-integration-cover-html \
	todoapp-run todoapp-deploy todoapp-undeploy \
	swagger-gen mock-gen mock-gen-stats mock-gen-users mock-gen-tasks \
	load-test todoapp-logs web-server-logs ps

help: ## Справка: показать доступные make-команды
	@awk 'BEGIN {FS = ":.*## "; printf "\nДоступные команды:\n"} /^[a-zA-Z0-9_-]+:.*## / {printf "  %-26s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

check: fmt-check vet test ## Проверки: форматирование, go vet и unit-тесты

check-all: fmt-check vet test test-integration ## Проверки: unit + integration тесты

fmt: ## Проверки: отформатировать Go-код
	@gofmt -w $$(find . -path './.out' -prune -o -name '*.go' -print)

fmt-check: ## Проверки: проверить gofmt без изменения файлов
	@files="$$(gofmt -l $$(find . -path './.out' -prune -o -name '*.go' -print))"; \
	if [ -n "$$files" ]; then \
		echo "Файлы требуют gofmt:"; \
		echo "$$files"; \
		exit 1; \
	fi

vet: ## Проверки: go vet
	@go vet ./...

env-up: ## Окружение: запустить PostgreSQL и Redis
	@docker compose up -d ${ENV_SERVICES}

env-down: ## Окружение: остановить PostgreSQL и Redis
	@docker compose stop ${ENV_SERVICES}
	@docker compose rm -f ${ENV_SERVICES}

env-cleanup: ## Окружение: удалить volume-файлы PostgreSQL, Redis и Caddy
	@read -p "Очистить все volume файлы окружения? Опасность утери данных. [y/n]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down --remove-orphans && \
		sudo rm -rf ${PROJECT_ROOT}/.out/pgdata ${PROJECT_ROOT}/.out/redis_data ${PROJECT_ROOT}/.out/caddy_data && \
		echo "Файлы окружения очищены"; \
	else \
		echo "Очистка окружения отменена"; \
	fi

migrate-create: ## Миграции: создать новую миграцию seq=name
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует необходимый параметр seq. Пример: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	docker compose run --rm todoapp-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up: ## Миграции: применить миграции
	@make migrate-action action=up
	
migrate-down: ## Миграции: откатить миграции
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует необходимый параметр action. Пример: make migrate-action action=up"; \
		exit 1; \
	fi; \
	docker compose run --rm todoapp-postgres-migrate \
	-path /migrations \
	-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@todoapp-postgres:5432/${POSTGRES_DB}?sslmode=disable \
	"$(action)"

logs-cleanup: ## Логи: удалить локальные log-файлы приложения
	@read -p "Очистить log файлы? Опасность утери логов. [y/n]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down --remove-orphans && \
		sudo rm -rf ${PROJECT_ROOT}/.out/logs && \
		echo "Файлы логов очищены"; \
	else \
		echo "Очистка логов отменена"; \
	fi

test: ## Тесты: unit и HTTP handler тесты
	@go test ./...

test-unit: test ## Тесты: алиас для unit и HTTP handler тестов

test-all: test test-integration ## Тесты: запуск unit + integration тестов

test-cover: ## Покрытие: unit coverage в консоль
	@packages="$$(go list ./... | grep -v '/docs$$' | grep -v '/mocks$$' | grep -v '/scripts/load_test$$')"; \
	go test $$packages -cover

test-cover-profile:
	@mkdir -p ${COVERAGE_DIR}
	@packages="$$(go list ./... | grep -v '/docs$$' | grep -v '/mocks$$' | grep -v '/scripts/load_test$$')"; \
	go test $$packages -coverprofile=${COVERAGE_PROFILE}

test-cover-func: test-cover-profile ## Покрытие: unit coverage по функциям
	@go tool cover -func=${COVERAGE_PROFILE}

test-cover-html: test-cover-profile ## Покрытие: HTML-отчет unit-тестов
	@go tool cover -html=${COVERAGE_PROFILE} -o ${COVERAGE_HTML}
	@echo "Coverage HTML: ${COVERAGE_HTML}"

test-race: ## Тесты: race detector
	@go test ./... -race

test-integration: ## Тесты: интеграционные тесты через Testcontainers
	@go test -tags=integration ./...

test-integration-cover: ## Покрытие: integration coverage по функциям
	@mkdir -p ${COVERAGE_DIR}
	@packages="$$(go list ./... | grep -v '/docs$$' | grep -v '/mocks$$' | grep -v '/scripts/load_test$$')"; \
	go test -tags=integration $$packages -coverprofile=${INTEGRATION_COVERAGE_PROFILE}
	@go tool cover -func=${INTEGRATION_COVERAGE_PROFILE}

test-integration-cover-html: ## Покрытие: HTML-отчет integration-тестов
	@mkdir -p ${COVERAGE_DIR}
	@packages="$$(go list ./... | grep -v '/docs$$' | grep -v '/mocks$$' | grep -v '/scripts/load_test$$')"; \
	go test -tags=integration $$packages -coverprofile=${INTEGRATION_COVERAGE_PROFILE}
	@go tool cover -html=${INTEGRATION_COVERAGE_PROFILE} -o ${INTEGRATION_COVERAGE_HTML}
	@echo "Integration coverage HTML: ${INTEGRATION_COVERAGE_HTML}"

todoapp-run: ## Приложение: запустить локально через go run
	@export LOGGER_FOLDER=${PROJECT_ROOT}/.out/logs && \
	export POSTGRES_HOST=localhost && \
	export POSTGRES_PORT=5433 && \
	export REDIS_HOST=localhost && \
	export REDIS_PORT=6379 && \
	go mod tidy && \
	go run ${PROJECT_ROOT}/cmd/todoapp/main.go

todoapp-deploy: ## Приложение: запустить todoapp и Caddy в Docker
	@docker compose up -d --build todoapp web-server

todoapp-undeploy: ## Приложение: остановить todoapp и Caddy
	@docker compose stop web-server todoapp && docker compose rm -f web-server todoapp

swagger-gen: ## Swagger: сгенерировать docs/*
	@docker compose run --rm --user "$$(id -u):$$(id -g)" swagger \
		init \
		-g cmd/todoapp/main.go \
		-o docs \
		--parseInternal \
		--parseDependency

mock-gen: mock-gen-stats mock-gen-users mock-gen-tasks ## Моки: сгенерировать все GoMock-моки

mock-gen-stats: ## Моки: сгенерировать моки stats
	@go run go.uber.org/mock/mockgen@${MOCKGEN_VERSION} \
		-source=internal/features/stats/service/service.go \
		-destination=internal/features/stats/service/mocks/mock_service.go \
		-package=stats_service_mocks
	@go run go.uber.org/mock/mockgen@${MOCKGEN_VERSION} \
		-source=internal/features/stats/transport/http/transport.go \
		-destination=internal/features/stats/transport/http/mocks/mock_transport.go \
		-package=stats_transport_http_mocks

mock-gen-users: ## Моки: сгенерировать моки users
	@go run go.uber.org/mock/mockgen@${MOCKGEN_VERSION} \
		-source=internal/features/users/service/service.go \
		-destination=internal/features/users/service/mocks/mock_service.go \
		-package=users_service_mocks
	@go run go.uber.org/mock/mockgen@${MOCKGEN_VERSION} \
		-source=internal/features/users/transport/http/transport.go \
		-destination=internal/features/users/transport/http/mocks/mock_transport.go \
		-package=users_transport_http_mocks

mock-gen-tasks: ## Моки: сгенерировать моки tasks
	@go run go.uber.org/mock/mockgen@${MOCKGEN_VERSION} \
		-source=internal/features/tasks/service/service.go \
		-destination=internal/features/tasks/service/mocks/mock_service.go \
		-package=tasks_service_mocks
	@go run go.uber.org/mock/mockgen@${MOCKGEN_VERSION} \
		-source=internal/features/tasks/transport/http/transport.go \
		-destination=internal/features/tasks/transport/http/mocks/mock_transport.go \
		-package=tasks_transport_http_mocks

load-test: ## Тесты: нагрузочное тестирование
	@go run scripts/load_test/main.go \
		-users 10 \
		-tasks-per-user 1000 \
		-concurrency 100 \
		-phase-duration 30s \
		-read-burst 50 \
		-mixed-reads 10 \
		-mixed-writes 1 \
		-report ${PROJECT_ROOT}/.out/load_test/result.txt

todoapp-logs: ## Логи: смотреть логи todoapp
	@docker compose logs -f todoapp

web-server-logs: ## Логи: смотреть логи Caddy
	@docker compose logs -f web-server

ps: ## Docker: показать запущенные сервисы
	@docker compose ps
