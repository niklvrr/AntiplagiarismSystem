# API Gateway

HTTP API Gateway для системы антиплагиата. Единая точка входа для внешних клиентов.

## Назначение

API Gateway предоставляет:

- REST API для работы с системой антиплагиата
- Маршрутизацию запросов к микросервисам
- Преобразование HTTP запросов в gRPC вызовы
- Маппинг gRPC ошибок в HTTP статусы
- Swagger UI для документации API
- Middleware для логирования и обработки ошибок

## Архитектура

- **transport** - HTTP handlers и роутинг
  - **handler** - обработчики HTTP запросов
  - **middleware** - middleware для логирования и recovery
  - **router** - настройка маршрутов
  - **swagger** - обслуживание Swagger UI
  - **error_mapper** - маппинг gRPC ошибок в HTTP статусы
- **infrastructure** - gRPC клиенты для микросервисов
  - **storing** - клиент для storing-service
  - **analysis** - клиент для analysis-service

## API Endpoints

### POST /api/v1/task

Создает новую задачу загрузки файла.

**Request:**
```json
{
  "filename": "document.pdf",
  "uploaded_by": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response:**
```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440000",
  "upload_url": "https://minio:9000/tasks/..."
}
```

### GET /api/v1/task/{task_id}

Получает информацию о задаче.

**Response:**
```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440000",
  "filename": "document.pdf",
  "url": "https://minio:9000/tasks/...",
  "uploaded_by": "550e8400-e29b-41d4-a716-446655440000",
  "uploaded_at": "2024-01-01T00:00:00Z"
}
```

### POST /api/v1/analyse

Запускает анализ документа на плагиат.

**Request:**
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "filename": "document.pdf"
}
```

**Response:**
```json
{
  "status": true
}
```

### GET /api/v1/report/{task_id}

Получает результат анализа документа.

**Response:**
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "is_plagiarism": false,
  "plagiarism_percentage": 15.5
}
```

### GET /api/v1/wordcloud/{task_id}

Генерирует облако слов для документа.

**Response:**
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "image_url": "https://quickchart.io/wordcloud?c=..."
}
```

## Swagger UI

Интерактивная документация API доступна по адресу:

- Swagger UI: http://localhost:8080/swagger
- OpenAPI спецификация: http://localhost:8080/swagger/openapi.yaml

## Конфигурация

Переменные окружения:

- `HTTP_PORT` - порт HTTP сервера (по умолчанию 8080)
- `STORING_SERVICE_ENDPOINT` - endpoint storing-service (формат: host:port)
- `ANALYSIS_SERVICE_ENDPOINT` - endpoint analysis-service (формат: host:port)
- `LOG_LEVEL` - уровень логирования (debug, info, warn, error, prod)

## Маппинг ошибок

gRPC ошибки преобразуются в HTTP статусы:

- `codes.InvalidArgument` → `400 Bad Request`
- `codes.NotFound` → `404 Not Found`
- `codes.AlreadyExists` → `409 Conflict`
- `codes.Unavailable` → `503 Service Unavailable`
- `codes.Internal` → `500 Internal Server Error`

## Middleware

### LoggingMiddleware

Логирует все HTTP запросы с информацией о:
- Методе и пути запроса
- Query параметрах
- IP адресе клиента
- HTTP статусе ответа
- Размере ответа
- Времени выполнения

### RecoveryMiddleware

Перехватывает паники и возвращает HTTP 500 с логированием ошибки.

## Запуск

### Docker

Сервис запускается через Docker Compose (см. корневой README).

### Локально

1. Установите зависимости:

```bash
go mod download
```

2. Настройте переменные окружения (создайте `.env` файл)

3. Запустите сервис:

```bash
go run cmd/main.go
```

API Gateway будет доступен по адресу http://localhost:8080

## Graceful Shutdown

Сервис поддерживает graceful shutdown:

1. При получении SIGTERM или SIGINT сервер прекращает прием новых запросов
2. Ожидает завершения текущих запросов (таймаут 30 секунд)
3. Закрывает соединения с микросервисами
4. Завершает работу

## Зависимости

- `github.com/go-chi/chi/v5` - HTTP роутер
- `go.uber.org/zap` - структурированное логирование
- `google.golang.org/grpc` - gRPC клиенты

