include .env
export

export PROJECT_ROOT=${shell pwd}
ENV_SERVICES=todoapp-postgres todoapp-redis

env-up:
	@docker compose up -d ${ENV_SERVICES}

env-down:
	@docker compose stop ${ENV_SERVICES}
	@docker compose rm -f ${ENV_SERVICES}

env-cleanup:
	@read -p "Очистить все volume файлы окружения? Опасность утери данных. [y/n]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down --remove-orphans && \
		sudo rm -rf ${PROJECT_ROOT}/.out/pgdata ${PROJECT_ROOT}/.out/redis_data && \
		echo "Файлы окружения очищены"; \
	else \
		echo "Очистка окружения отменена"; \
	fi

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует необходимый параметр seq. Пример: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	docker compose run --rm todoapp-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up:
	@make migrate-action action=up
	
migrate-down:
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

logs-cleanup:
	@read -p "Очистить log файлы? Опасность утери логов. [y/n]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down --remove-orphans && \
		sudo rm -rf ${PROJECT_ROOT}/.out/logs && \
		echo "Файлы логов очищены"; \
	else \
		echo "Очистка логов отменена"; \
	fi

todoapp-run:
	@export LOGGER_FOLDER=${PROJECT_ROOT}/.out/logs && \
	export POSTGRES_HOST=localhost && \
	export POSTGRES_PORT=5433 && \
	export REDIS_HOST=localhost && \
	export REDIS_PORT=6379 && \
	go mod tidy && \
	go run ${PROJECT_ROOT}/cmd/todoapp/main.go

todoapp-deploy:
	@docker compose up -d --build todoapp

todoapp-undeploy:
	@docker compose stop todoapp && docker compose rm -f todoapp

swagger-gen:
	@docker compose run --rm swagger \
		init \
		-g cmd/todoapp/main.go \
		-o docs \
		--parseInternal \
		--parseDependency

todoapp-logs:
	@docker compose logs -f todoapp

ps:
	@docker compose ps
