# Сервис для работы с ПВЗ

Система управления пунктами выдачи заказов.

## Основные возможности

- **Авторизация пользователей**: Регистрация и вход через JWT-токены
- **Управление ПВЗ**: Создание и просмотр пунктов выдачи заказов
- **Приёмка товаров**: Инициирование приёмки, добавление и удаление товаров
- **История операций**: Полная информация о ПВЗ и их приёмках
- **Поддержка высокой нагрузки**: Оптимизация для 1000 RPS с временем ответа < 100 мс

## Стек технологий

- **Язык**: Go 1.24
- **Веб-фреймворк**: Gin
- **Аутентификация**: JWT
- **База данных**: PostgreSQL
- **Доступ к данным**: SQL с использованием Squirrel для построения запросов
- **Логирование**: Zerolog
- **Метрики**: Prometheus
- **API документация**: OpenAPI (Swagger)
- **gRPC**: для получения данных о ПВЗ
- **Тестирование**: Testify, gomock
- **Контейнеризация**: Docker и Docker Compose
- **Кодогенерация**: Dto через oapi-codegen

## Архитектура

Проект построен с использованием чистой архитектуры:

- `/cmd` - Точки входа в приложение (HTTP API, gRPC сервер)
- `/internal` - Внутренние компоненты:
  - `/api` - HTTP API слой (handlers, DTO, middleware)
  - `/domain` - Доменные модели
  - `/repository` - Слой доступа к данным
  - `/services` - Сервисы приложения
  - `/auth` - Аутентификация и авторизация
- `/pkg` - Переиспользуемые компоненты (конфигурация, логирование, метрики)
- `/tests` - Интеграционные и юнит-тесты
- `/migrations` - Миграции схемы базы данных

## Установка и запуск

1. Клонировать репозиторий:
```bash
git clone https://github.com/mihailpestrikov/avito-backend-trainee-assignment-spring-2025
cd avito-backend-trainee-assignment-spring-2025
```

2. Запустить приложение:
```bash
docker-compose --env-file .env -f docker-compose.yml up -d --build
```
или
```bash
make up
```

3. Сервисы доступны по адресам:
```
HTTP API: http://localhost:8080
Prometheus: http://localhost:9090
gRPC: localhost:3000
```

## API

### Аутентификация и пользователи

- **POST /dummyLogin** - Получение тестового токена с выбранной ролью
- **POST /register** - Регистрация нового пользователя
- **POST /login** - Авторизация пользователя

### Управление ПВЗ

- **POST /pvz** - Создание нового пункта выдачи заказов (только модераторы)
- **GET /pvz** - Получение списка ПВЗ с фильтрацией и пагинацией

### Приёмка товаров

- **POST /receptions** - Создание новой приёмки товаров
- **POST /pvz/{pvzId}/close_last_reception** - Закрытие последней открытой приёмки товаров
- **POST /products** - Добавление товара в текущую приёмку
- **POST /pvz/{pvzId}/delete_last_product** - Удаление последнего добавленного товара (LIFO)

## gRPC API

Сервис также предоставляет gRPC-метод для получения списка всех ПВЗ:
- **GetPVZList** - Возвращает все добавленные в систему ПВЗ без авторизации

Пример использования с помощью grpcurl:
```bash
grpcurl -plaintext localhost:3000 pvz.v1.PVZService/GetPVZList
```
```bash
docker run --rm -it --network=host fullstorydev/grpcurl -plaintext localhost:3000 pvz.v1.PVZService/GetPVZList
```

## Тестирование

Проект включает:
- Юнит-тесты:
```bash
go test -v ./internal/... ./pkg/...
```
или
```bash
make unit-test
```

- Интеграционные тесты:
```bash
docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit integration-tests
```
или 
```bash
make test
```

## Команды
```bash
build:
	docker-compose --env-file .env -f docker-compose.yml build

up: build
	docker-compose --env-file .env -f docker-compose.yml up -d

test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit integration-tests

unit-test:
	go test -v ./internal/... ./pkg/...

down:
	docker-compose -f docker-compose.yml down
```

## Пример конфигурации .env
```bash
APP_NAME=pvz-service
APP_ENV=development # development, production, testing
APP_PORT=8080
APP_GRPC_PORT=3000
APP_PROMETHEUS_PORT=9000

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=pvz_db
POSTGRES_USER=postgres
POSTGRES_PASSWORD=production_password
POSTGRES_SSL_MODE=disable
POSTGRES_MAX_CONNECTIONS=20
POSTGRES_IDLE_CONNECTIONS=5
POSTGRES_CONNECTION_LIFETIME=300

JWT_SECRET=very_secure_jwt_secret_key
JWT_EXPIRATION=24h

LOG_LEVEL=debug  # debug, info, warn, error, fatal, panic
LOG_FORMAT=console  # json, console
LOG_OUTPUT=stdout  # stdout, file
LOG_FILE_PATH=./logs/app.log  # Используется, только если LOG_OUTPUT=file

HTTP_READ_TIMEOUT=5s
HTTP_WRITE_TIMEOUT=10s
HTTP_IDLE_TIMEOUT=120s
DB_QUERY_TIMEOUT=5s
```
