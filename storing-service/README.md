# Storing Service

Микросервис для управления задачами загрузки файлов и хранения метаданных.

## Назначение

Сервис отвечает за:

- Создание задач загрузки файлов
- Генерацию presigned URL для загрузки файлов в MinIO
- Хранение метаданных задач в PostgreSQL
- Получение содержимого файлов из MinIO
- Асинхронный запуск анализа после загрузки файла

## Архитектура

Сервис следует принципам Clean Architecture:

- **transport** - gRPC handlers для обработки запросов
- **usecase** - бизнес-логика сервиса
- **infrastructure** - реализация репозиториев и внешних клиентов
  - **pgdb** - репозиторий для работы с PostgreSQL
  - **minio** - клиент для работы с MinIO
  - **analysis** - gRPC клиент для вызова analysis-service

## API

### UploadTask

Создает новую задачу загрузки файла.

**Request:**
```protobuf
message UploadTaskRequest {
  string filename = 1;
  string uploaded_by = 2;
}
```

**Response:**
```protobuf
message UploadTaskResponse {
  string file_id = 1;
  string upload_url = 2;
}
```

### GetTask

Получает информацию о задаче и presigned URL для скачивания файла.

**Request:**
```protobuf
message GetTaskRequest {
  string file_id = 1;
}
```

**Response:**
```protobuf
message GetTaskResponse {
  string file_id = 1;
  string filename = 2;
  string url = 3;
  string uploaded_by = 4;
  string uploaded_at = 5;
}
```

### GetFileContent

Получает содержимое файла из MinIO.

**Request:**
```protobuf
message GetFileContentRequest {
  string file_id = 1;
}
```

**Response:**
```protobuf
message GetFileContentResponse {
  bytes content = 1;
}
```

## Конфигурация

Переменные окружения:

- `GRPC_PORT` - порт gRPC сервера (по умолчанию 50051)
- `DB_HOST` - хост PostgreSQL
- `DB_PORT` - порт PostgreSQL (по умолчанию 5432)
- `DB_NAME` - имя базы данных
- `DB_USER` - пользователь базы данных
- `DB_PASSWORD` - пароль базы данных
- `MINIO_ENDPOINT` - endpoint MinIO (формат: host:port)
- `MINIO_ACCESS_KEY` - ключ доступа MinIO
- `MINIO_SECRET_KEY` - секретный ключ MinIO
- `MINIO_BUCKET` - имя bucket в MinIO (по умолчанию tasks)
- `ANALYSIS_URL` - endpoint analysis-service (формат: host:port)
- `LOG_LEVEL` - уровень логирования (debug, info, warn, error, prod)

## База данных

### Таблица tasks

```sql
CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    uploaded_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
```

Миграции находятся в директории `migrations/`.

## Запуск

### Docker

Сервис запускается через Docker Compose (см. корневой README).

### Локально

1. Установите зависимости:

```bash
go mod download
```

2. Настройте переменные окружения (создайте `.env` файл)

3. Выполните миграции БД

4. Запустите сервис:

```bash
go run cmd/main.go
```

## Генерация proto файлов

```bash
make proto
```

Или вручную:

```bash
protoc -I api api/storing_service.proto \
  --go_out=./pkg/api \
  --go_opt=paths=source_relative \
  --go-grpc_out=./pkg/api \
  --go-grpc_opt=paths=source_relative
```

## Асинхронный анализ

После создания задачи и генерации presigned URL сервис запускает фоновую горутину, которая:

1. Ожидает загрузки файла в MinIO (проверка каждые 2 секунды, максимум 30 попыток)
2. После подтверждения загрузки вызывает analysis-service через gRPC
3. Логирует результаты операции

Таймаут ожидания файла: 5 минут.

