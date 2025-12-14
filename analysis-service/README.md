# Analysis Service

Микросервис для анализа документов на плагиат с использованием алгоритма n-грамм и коэффициента Жаккара.

## Назначение

Сервис отвечает за:

- Анализ документов на плагиат
- Сравнение документов с использованием n-грамм
- Вычисление коэффициента Жаккара
- Генерацию облаков слов для визуализации
- Хранение результатов анализа в PostgreSQL

## Архитектура

Сервис следует принципам Clean Architecture:

- **transport** - gRPC handlers для обработки запросов
- **usecase** - бизнес-логика анализа
  - **service** - основной сервис анализа
  - **comparator** - реализация алгоритма сравнения
- **infrastructure** - реализация репозиториев и внешних клиентов
  - **pgdb** - репозиторий для работы с PostgreSQL
  - **minio** - клиент для работы с MinIO
  - **wordcloud** - клиент для генерации облаков слов через QuickChart API

## Алгоритм анализа

### N-граммы

Документ разбивается на последовательности из n символов (n-граммы). По умолчанию используется n=3 (триграммы).

### Коэффициент Жаккара

Для двух документов A и B:

1. Извлекаются множества уникальных n-грамм: A и B
2. Вычисляется пересечение: A ∩ B
3. Вычисляется объединение: A ∪ B
4. Коэффициент Жаккара: J(A,B) = |A ∩ B| / |A ∪ B|
5. Процент схожести: similarity = J(A,B) * 100%

### Определение плагиата

Документ считается плагиатом, если максимальный процент схожести с любым другим документом >= 50%.

## API

### AnalyseTask

Запускает анализ документа на плагиат.

**Request:**
```protobuf
message AnalyzeTaskRequest {
  string task_id = 1;
  string object_key = 2;
}
```

**Response:**
```protobuf
message AnalyseTaskResponse {
  bool status = 1;
}
```

### GetReport

Получает результат анализа документа.

**Request:**
```protobuf
message GetReportRequest {
  string task_id = 1;
}
```

**Response:**
```protobuf
message GetReportResponse {
  string task_id = 1;
  bool is_plagiarism = 4;
  float plagiarism_percentage = 5;
}
```

### GenerateWordCloud

Генерирует URL облака слов для документа.

**Request:**
```protobuf
message GenerateWordCloudRequest {
  bytes file_content = 1;
}
```

**Response:**
```protobuf
message GenerateWordCloudResponse {
  string image_url = 1;
}
```

## Конфигурация

Переменные окружения:

- `GRPC_PORT` - порт gRPC сервера (по умолчанию 50052)
- `DB_HOST` - хост PostgreSQL
- `DB_PORT` - порт PostgreSQL (по умолчанию 5432)
- `DB_NAME` - имя базы данных
- `DB_USER` - пользователь базы данных
- `DB_PASSWORD` - пароль базы данных
- `MINIO_ENDPOINT` - endpoint MinIO (формат: host:port)
- `MINIO_ACCESS_KEY` - ключ доступа MinIO
- `MINIO_SECRET_KEY` - секретный ключ MinIO
- `MINIO_BUCKET` - имя bucket в MinIO (по умолчанию tasks)
- `LOG_LEVEL` - уровень логирования (debug, info, warn, error, prod)

## База данных

### Таблица reports

```sql
CREATE TABLE reports (
    task_id UUID PRIMARY KEY NOT NULL,
    is_plagiarism BOOLEAN DEFAULT FALSE,
    plagiarism_percentage float DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
```

Миграции находятся в директории `migrations/`.

## Генерация облака слов

Сервис использует QuickChart API для генерации облаков слов:

1. Извлекает слова из текста документа (минимальная длина 3 символа)
2. Фильтрует стоп-слова (the, be, to, of, and и др.)
3. Подсчитывает частоту слов
4. Формирует конфигурацию для QuickChart API
5. Возвращает URL изображения облака слов

Параметры облака слов:

- Размер: 800x600 пикселей
- Шрифт: Arial
- Масштаб: sqrt
- Цвета: палитра из 6 цветов

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
protoc -I api api/analysis_service.proto \
  --go_out=./pkg/api \
  --go_opt=paths=source_relative \
  --go-grpc_out=./pkg/api \
  --go-grpc_opt=paths=source_relative
```

## Процесс анализа

1. Получение всех ключей файлов из MinIO
2. Фильтрация ключей (исключение текущего файла)
3. Загрузка текущего файла из MinIO
4. Для каждого другого файла:
   - Загрузка файла из MinIO
   - Извлечение n-грамм из обоих файлов
   - Вычисление коэффициента Жаккара
   - Обновление максимального процента схожести
5. Определение наличия плагиата (порог 50%)
6. Сохранение результата в БД

При ошибках загрузки или сравнения отдельных файлов процесс продолжается с остальными файлами.

