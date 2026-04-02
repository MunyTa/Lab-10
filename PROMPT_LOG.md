# PROMPT_LOG - Лабораторная работа №10

## Основная информация
- **Студент:** Кузьмищев Родион
- **Группа:** 221331
- **Вариант:** 8
- **Дата:** 02.04.2026
- **Репозиторий:** https://github.com/MunyTa/Lab-10

## Архитектура решения

### Микросервисы:
1. **Go Service (порт 8080)** - основной сервис на Gin
   - REST API для работы с пользователями
   - Middleware для JSON логирования
   - WebSocket чат с комнатами
   - OpenAPI документация (swaggo)

2. **Python Service (порт 8000)** - прокси на FastAPI
   - HTTP клиент к Go сервису
   - WebSocket клиент для чата
   - Swagger UI документация
   - Graceful shutdown

3. **Go Gateway (порт 8888)** - API шлюз
   - Маршрутизация /go/* → Go сервис
   - Маршрутизация /py/* → Python сервис
   - Rate limiting (10 req/sec)
   - Reverse proxy

### Технологии:
- **Go:** Gin, Gorilla WebSocket, Swaggo, httputil
- **Python:** FastAPI, Uvicorn, httpx, websockets
- **Тестирование:** pytest, testing (Go)

## Процесс разработки (Agentic Engineering)

### Итерация 1: Middleware для логирования в Go
- **RED:** Написаны тесты для middleware (7 тестов)
- **GREEN:** Реализован LoggingMiddleware с JSON форматом
- **REFACTOR:** Вынесена структура LogEntry
- **Коммит:** `feat(go): implement middleware for logging in JSON format`

### Итерация 2: FastAPI прокси к Go сервису
- **RED:** Тесты для HTTP клиента (9 тестов)
- **GREEN:** Реализованы эндпоинты /api/users/* с обработкой ошибок
- **REFACTOR:** Использован app.state для клиента
- **Коммит:** `feat(fastapi): implement HTTP client for calling GO service`

### Итерация 3: Swagger и OpenAPI документация
- **FastAPI:** Добавлены Pydantic модели, response_model, примеры
- **Gin:** Установлен swaggo, добавлены аннотации, сгенерирована docs
- **Коммиты:** `docs(fastapi): add Swagger`, `docs(go): add OpenAPI`

### Итерация 4: API шлюз на Go
- **RED:** Тесты для маршрутизации и rate limiting
- **GREEN:** Реализован reverse proxy с rate limiter
- **REFACTOR:** Вынесены middleware
- **Коммит:** `feat(gateway): add API gateway with routing`

### Итерация 5: WebSocket чат
- **Go:** Реализован Hub с комнатами, broadcast, goroutines
- **Python:** Асинхронный клиент на websockets
- **Тестирование:** Ручное с несколькими клиентами
- **Коммит:** `feat(websocket): add WebSocket chat to Go + Python client`

## Проблемы и решения

### Проблема 1: Асинхронные моки в pytest
- **Проблема:** Тесты падали из-за неправильной обработки async/await
- **Решение:** Использован `get_response_data()` для универсальной обработки

### Проблема 2: Rate limiting в шлюзе
- **Проблема:** Нужно ограничить количество запросов
- **Решение:** Использован token bucket алгоритм из `golang.org/x/time/rate`

### Проблема 3: Graceful shutdown WebSocket
- **Проблема:** Клиенты отключались без уведомления
- **Решение:** Добавлен close message при завершении (частично)

## Результаты тестирования

### Go middleware tests: 7/7 PASS ✅
### FastAPI client tests: 9/9 PASS ✅
### Gateway tests: 5/5 PASS ✅
### WebSocket tests: Требуют запущенного сервера (manual)

## Запуск проекта

```bash
# Терминал 1 - Go сервис
cd go-service
go run main.go

# Терминал 2 - FastAPI сервис
cd py-service
uvicorn main:app --port 8000

# Терминал 3 - Gateway
cd go-gateway
export GO_SERVICE_URL=http://localhost:8080
export PY_SERVICE_URL=http://localhost:8000
go run main.go

# Терминал 4 - WebSocket клиент
cd py-service
python websocket_client.py