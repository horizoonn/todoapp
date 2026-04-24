# Golang Todo App

REST API приложение на Go для управления пользователями, задачами и статистикой.
Проект собран как небольшой backend-сервис с PostgreSQL, Redis-кешем,
Swagger-документацией, статическим web UI и Caddy reverse proxy.

## Технологический Стек

| Компонент | Технология |
| --- | --- |
| Язык | Go 1.26+ |
| HTTP | стандартный `net/http` |
| База данных | PostgreSQL, `jackc/pgx/v5` |
| Кеш | Redis, `redis/go-redis/v9` |
| Web-server | Caddy 2 |
| Логирование | `go.uber.org/zap` |
| Конфигурация | `kelseyhightower/envconfig` |
| Валидация | `go-playground/validator/v10` |
| Документация API | Swagger, `swaggo/swag` |
| Миграции | `golang-migrate` |
| Запуск | Docker Compose, Makefile |

## Архитектура

Проект разделен по фичам:

- `users` - CRUD пользователей.
- `tasks` - CRUD задач.
- `stats` - статистика по задачам.

Внутри каждой фичи используется слоистая структура:

```text
Transport HTTP
    |
    | decode request, path/query params, response mapping
    v
Service
    |
    | business logic, validation, cache invalidation
    v
Repository
    |
    | PostgreSQL / Redis access
    v
Domain
```

Интерфейсы объявляются на стороне потребителя. Например, сервисы зависят от
интерфейсов репозиториев, а конкретные реализации PostgreSQL и Redis
подключаются вручную в `cmd/todoapp/main.go`.

Redis используется как кеш для:

- отдельных задач;
- списков задач с версионированием ключей;
- статистики с версионированием ключей.

Инвалидация кеша выполняется при создании, обновлении и удалении задач.

## Структура Проекта

```text
.
├── cmd/
│   └── todoapp/
│       ├── Dockerfile
│       └── main.go                    # Точка входа: DI, конфигурация, запуск HTTP server
├── docs/                              # Swagger docs, генерируется через make swagger-gen
├── internal/
│   ├── core/
│   │   ├── config/                    # Общая конфигурация
│   │   ├── domain/                    # Доменные сущности: User, Task, Stats
│   │   ├── errors/                    # Sentinel errors
│   │   ├── logger/                    # Zap logger
│   │   ├── repository/
│   │   │   ├── postgres/pool/         # Абстракция пула PostgreSQL + pgx adapter
│   │   │   └── redis/pool/            # Абстракция Redis + go-redis adapter
│   │   └── transport/http/
│   │       ├── middleware/            # CORS, RequestID, Logger, Trace, Panic recovery
│   │       ├── request/               # Decode, path/query helpers
│   │       ├── response/              # JSON/Error response helpers
│   │       ├── server/                # HTTPServer, routes, API version router
│   │       └── types/                 # Nullable[T] для PATCH-запросов
│   └── features/
│       ├── stats/
│       │   ├── repository/postgres/
│       │   ├── repository/redis/
│       │   ├── service/
│       │   └── transport/http/
│       ├── tasks/
│       │   ├── repository/postgres/
│       │   ├── repository/redis/
│       │   ├── service/
│       │   └── transport/http/
│       └── users/
│           ├── repository/postgres/
│           ├── service/
│           └── transport/http/
├── migrations/                        # SQL migrations
├── web/
│   ├── Caddyfile                      # Caddy: static web + reverse proxy
│   └── public/                        # Static web UI
├── docker-compose.yaml
├── Makefile
└── README.md
```

## Быстрый Старт

### Требования

- Docker и Docker Compose
- Go 1.26+
- `make`

### Настройка `.env`

```bash
cp .env.example .env
```

Минимально заполни PostgreSQL-переменные:

```env
POSTGRES_USER=todoapp
POSTGRES_PASSWORD=todoapp
POSTGRES_DB=todoapp
```

Остальные значения можно оставить из `.env.example`.

## Локальный Запуск Go-Приложения

Этот режим запускает PostgreSQL и Redis в Docker, а Go-приложение запускается
локально через `go run`.

```bash
make env-up
make migrate-up
make todoapp-run
```

После запуска:

- Web UI: `http://localhost:5050/`
- Swagger UI: `http://localhost:5050/swagger/`
- API: `http://localhost:5050/api/v1/`

## Запуск Через Docker Compose

Этот режим поднимает PostgreSQL, Redis, миграции, Go-приложение и Caddy.
Caddy становится внешней точкой входа на портах `80` и `443`.

```bash
make env-up
make migrate-up
make todoapp-deploy
```

После запуска:

- Web UI: `http://localhost/` или `https://localhost/`
- Swagger UI: `https://localhost/swagger/`
- API: `https://localhost/api/v1/`
- Go-приложение напрямую: `http://localhost:5050/`

Для `localhost` браузер может показать предупреждение по локальному TLS-сертификату Caddy.

## Makefile Команды

| Команда | Описание |
| --- | --- |
| `make env-up` | Поднять PostgreSQL и Redis |
| `make env-down` | Остановить и удалить PostgreSQL и Redis containers |
| `make env-cleanup` | Остановить окружение и удалить local volumes |
| `make migrate-create seq=name` | Создать новую SQL migration |
| `make migrate-up` | Применить миграции |
| `make migrate-down` | Откатить миграции |
| `make todoapp-run` | Запустить приложение локально через `go run` |
| `make todoapp-deploy` | Запустить Go-приложение и Caddy в Docker |
| `make todoapp-undeploy` | Остановить Go-приложение и Caddy |
| `make swagger-gen` | Перегенерировать Swagger docs |
| `make todoapp-logs` | Смотреть логи Go-приложения |
| `make web-server-logs` | Смотреть логи Caddy |
| `make ps` | Показать containers |

## Swagger

Swagger UI доступен:

- напрямую через Go-приложение: `http://localhost:5050/swagger/`
- через Caddy: `https://localhost/swagger/`

Перегенерация документации:

```bash
make swagger-gen
```

Сгенерированные файлы находятся в `docs/`:

- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`

## API

Все endpoints находятся под префиксом `/api/v1`.

### Users

| Метод | Путь | Описание |
| --- | --- | --- |
| `POST` | `/api/v1/users` | Создать пользователя |
| `GET` | `/api/v1/users` | Получить список пользователей |
| `GET` | `/api/v1/users/{id}` | Получить пользователя по ID |
| `PATCH` | `/api/v1/users/{id}` | Частично обновить пользователя |
| `DELETE` | `/api/v1/users/{id}` | Удалить пользователя |

### Tasks

| Метод | Путь | Описание |
| --- | --- | --- |
| `POST` | `/api/v1/tasks` | Создать задачу |
| `GET` | `/api/v1/tasks` | Получить список задач |
| `GET` | `/api/v1/tasks/{id}` | Получить задачу по ID |
| `PATCH` | `/api/v1/tasks/{id}` | Частично обновить задачу |
| `DELETE` | `/api/v1/tasks/{id}` | Удалить задачу |

`GET /api/v1/tasks` поддерживает query params:

- `user_id` - фильтр по автору задачи;
- `limit` - размер страницы;
- `offset` - смещение.

### Stats

| Метод | Путь | Описание |
| --- | --- | --- |
| `GET` | `/api/v1/stats` | Получить статистику по задачам |

`GET /api/v1/stats` поддерживает query params:

- `user_id` - фильтр по пользователю;
- `from` - начало периода, формат `YYYY-MM-DD`;
- `to` - конец периода, формат `YYYY-MM-DD`.

## Примеры Curl

Создать пользователя:

```bash
curl -k -X POST https://localhost/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"full_name":"Ivan Ivanov","phone_number":"+79998887766"}'
```

Создать задачу:

В `author_user_id` нужно передать `id` пользователя, созданного предыдущим запросом.

```bash
curl -k -X POST https://localhost/api/v1/tasks \
  -H 'Content-Type: application/json' \
  -d '{
    "title":"Homework",
    "description":"Finish math homework by Thursday",
    "author_user_id":"550e8400-e29b-41d4-a716-446655440000"
  }'
```

Получить статистику:

```bash
curl -k 'https://localhost/api/v1/stats'
```

## Переменные Окружения

| Переменная | Описание | Пример |
| --- | --- | --- |
| `POSTGRES_HOST` | Host PostgreSQL | `localhost`, `todoapp-postgres` |
| `POSTGRES_PORT` | Port PostgreSQL | `5432`, `5433` |
| `POSTGRES_USER` | Пользователь БД | `todoapp` |
| `POSTGRES_PASSWORD` | Пароль БД | `todoapp` |
| `POSTGRES_DB` | Имя БД | `todoapp` |
| `POSTGRES_TIMEOUT` | Таймаут операций PostgreSQL | `10s` |
| `REDIS_HOST` | Host Redis | `localhost`, `todoapp-redis` |
| `REDIS_PORT` | Port Redis | `6379` |
| `REDIS_PASSWORD` | Пароль Redis | `admin` |
| `REDIS_DB` | Redis DB number | `0` |
| `REDIS_TTL` | TTL кеша | `5m` |
| `LOGGER_LEVEL` | Уровень логирования | `DEBUG` |
| `LOGGER_FOLDER` | Папка логов | `.out/logs` |
| `HTTP_ADDR` | Адрес HTTP server | `:5050` |
| `HTTP_SHUTDOWN_TIMEOUT` | Graceful shutdown timeout | `30s` |
| `HTTP_ALLOWED_ORIGINS` | CORS origins через запятую | `http://localhost:3000,http://localhost:5050` |
| `CADDY_HOST` | Host/domain для Caddy | `localhost` |
| `TIME_ZONE` | IANA time zone | `UTC` |
| `PROJECT_ROOT` | Корень проекта для Docker volumes | задается в `Makefile` |

## Caddy

Caddy отдает frontend и проксирует backend:

```text
GET /           -> /web/public/index.html
GET /api/*      -> todoapp:5050
GET /swagger*   -> todoapp:5050
```

Конфигурация находится в `web/Caddyfile`.

## Миграции

Создать миграцию:

```bash
make migrate-create seq=add_some_table
```

Применить:

```bash
make migrate-up
```

Откатить:

```bash
make migrate-down
```

После применения миграции лучше не редактировать старые migration files. Для изменений схемы создавай новую миграцию.

## Проверки

```bash
go test ./...
go build -o /tmp/todoapp-check ./cmd/todoapp
docker compose config --quiet
```
