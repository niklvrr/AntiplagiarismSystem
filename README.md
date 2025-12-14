# Система антиплагиата

Микросервисная система для проверки документов на плагиат с использованием алгоритма n-грамм и коэффициента Жаккара.

## Архитектура

Система состоит из трех микросервисов:

- **api-gateway** - HTTP API Gateway, единая точка входа для клиентов
- **storing-service** - сервис хранения файлов и метаданных задач
- **analysis-service** - сервис анализа документов на плагиат

### Компоненты инфраструктуры

- **PostgreSQL** - две базы данных для хранения метаданных (storing-db, analysis-db)
- **MinIO** - объектное хранилище для файлов
- **gRPC** - протокол межсервисного взаимодействия
- **HTTP/REST** - протокол для внешних клиентов

## Алгоритм определения плагиата

Система использует алгоритм сравнения документов на основе n-грамм и коэффициента Жаккара:

1. Извлечение n-грамм из текста документа (последовательности из n символов)
2. Вычисление множества уникальных n-грамм для каждого документа
3. Расчет коэффициента Жаккара: J(A,B) = |A ∩ B| / |A ∪ B|
4. Определение процента схожести: similarity = J(A,B) * 100%
5. Порог плагиата: 50% и выше

## Пользовательские сценарии

### Сценарий 1: Загрузка документа

1. Клиент отправляет POST запрос на `/api/v1/task` с именем файла и идентификатором пользователя
2. API Gateway перенаправляет запрос в storing-service
3. Storing-service создает запись в БД и генерирует presigned URL для загрузки в MinIO
4. Клиент получает file_id и upload_url
5. Клиент загружает файл напрямую в MinIO по presigned URL
6. Storing-service асинхронно запускает анализ после загрузки файла

### Сценарий 2: Анализ документа

1. Storing-service проверяет наличие файла в MinIO
2. После подтверждения загрузки вызывается analysis-service
3. Analysis-service получает все файлы из MinIO
4. Текущий файл сравнивается с каждым существующим файлом
5. Вычисляется максимальный процент схожести
6. Результат сохраняется в БД analysis-service

### Сценарий 3: Получение отчета

1. Клиент отправляет GET запрос на `/api/v1/report/{task_id}`
2. API Gateway перенаправляет запрос в analysis-service
3. Analysis-service возвращает результат анализа из БД
4. Клиент получает информацию о наличии плагиата и проценте схожести

### Сценарий 4: Визуализация облака слов

1. Клиент отправляет GET запрос на `/api/v1/wordcloud/{task_id}`
2. API Gateway получает содержимое файла из storing-service
3. API Gateway передает содержимое в analysis-service
4. Analysis-service извлекает слова, подсчитывает частоту и генерирует URL облака слов через QuickChart API
5. Клиент получает URL изображения облака слов

## Требования

- Docker 20.10+
- Docker Compose 2.0+
- Go 1.25+ (для локальной разработки)
- protoc (для генерации proto файлов)

## Запуск системы

### Использование Docker Compose

1. Клонируйте репозиторий
2. Создайте файлы `.env` для каждого сервиса (см. примеры в директориях сервисов)
3. Запустите систему:

```bash
docker-compose up -d
```

4. Проверьте статус сервисов:

```bash
docker-compose ps
```

5. API Gateway доступен по адресу: http://localhost:8080
6. Swagger UI доступен по адресу: http://localhost:8080/swagger

### Переменные окружения

Основные переменные окружения задаются в `docker-compose.yaml` и могут быть переопределены через файлы `.env`:

- `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD` - учетные данные MinIO
- `STORING_DB_*` - параметры подключения к БД storing-service
- `ANALYSIS_DB_*` - параметры подключения к БД analysis-service
- `STORING_GRPC_PORT` - порт gRPC storing-service (по умолчанию 50051)
- `ANALYSIS_GRPC_PORT` - порт gRPC analysis-service (по умолчанию 50052)
- `GATEWAY_HTTP_PORT` - порт HTTP API Gateway (по умолчанию 8080)

## API Endpoints

### Загрузка задачи

```
POST /api/v1/task
Content-Type: application/json

{
  "filename": "document.pdf",
  "uploaded_by": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Получение задачи

```
GET /api/v1/task/{task_id}
```

### Запуск анализа

```
POST /api/v1/analyse
Content-Type: application/json

{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "filename": "document.pdf"
}
```

### Получение отчета

```
GET /api/v1/report/{task_id}
```

### Получение облака слов

```
GET /api/v1/wordcloud/{task_id}
```

Полная документация API доступна в Swagger UI по адресу http://localhost:8080/swagger

## Структура проекта

```
.
├── api-gateway/          # HTTP API Gateway
├── storing-service/      # Сервис хранения файлов
├── analysis-service/     # Сервис анализа плагиата
└── docker-compose.yaml  # Конфигурация Docker Compose
```

## Разработка

### Генерация proto файлов

Для каждого сервиса выполните:

```bash
cd storing-service && make proto
cd analysis-service && make proto
```

### Локальный запуск

1. Запустите инфраструктуру (PostgreSQL, MinIO):

```bash
docker-compose up -d minio storing-db analysis-db
```

2. Выполните миграции БД
3. Запустите сервисы локально:

```bash
cd storing-service && go run cmd/main.go
cd analysis-service && go run cmd/main.go
cd api-gateway && go run cmd/main.go
```

## Логирование

Все сервисы используют структурированное логирование через zap. Уровни логирования:

- `debug` - детальная информация для отладки
- `info` - информационные сообщения
- `warn` - предупреждения
- `error` - ошибки

Уровень логирования настраивается через переменную окружения `LOG_LEVEL`.

## Остановка системы

```bash
docker-compose down
```

Для удаления всех данных (включая volumes):

```bash
docker-compose down -v
```

