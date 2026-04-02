# Лабораторная работа №10 - Веб-разработка: FastAPI (Python) vs Gin (Go)

**Студент:** Кузьмищев Родион  
**Группа:** 221331  
**Вариант:** 8  

## Выполненные задания

### Средней сложности
1. ✅ **Middleware для логирования в Go** - JSON формат, логирует метод, путь, статус, время, IP
2. ✅ **FastAPI сервис, вызывающий Go через HTTP** - прокси с обработкой ошибок и таймаутов
3. ✅ **Swagger для FastAPI и OpenAPI для Gin** - интерактивная документация

### Повышенной сложности
1. ✅ **API-шлюз на Go** - маршрутизация /go/* и /py/*, rate limiting (10 req/sec)
2. ✅ **WebSocket чат** - Go сервер + Python клиент, комнаты, broadcast

## Быстрый старт

### Запуск всех сервисов

```bash
# Терминал 1 - Go сервис (порт 8080)
cd go-service
go run main.go

# Терминал 2 - FastAPI сервис (порт 8000)
cd py-service
uvicorn main:app --port 8000

# Терминал 3 - API Gateway (порт 8888)
cd go-gateway
export GO_SERVICE_URL=http://localhost:8080
export PY_SERVICE_URL=http://localhost:8000
go run main.go