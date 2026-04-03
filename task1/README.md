# Task 1 (ЛР10): Простое API на Go (Gin)

Реализовано 3 эндпоинта:

- `GET /health` — проверка статуса сервиса.
- `GET /hello/:name` — приветствие по имени.
- `POST /sum` — сумма двух чисел из JSON.

## Запуск

```powershell
cd task1
go mod tidy
go run .
```

По умолчанию сервер стартует на `http://127.0.0.1:8082`.

## Примеры запросов

```powershell
Invoke-RestMethod -Uri "http://127.0.0.1:8082/health"
Invoke-RestMethod -Uri "http://127.0.0.1:8082/hello/Alice"
Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8082/sum" -ContentType "application/json" -Body '{"a":2.5,"b":7.5}'
```

## Внутренние тесты в `main.go`

```powershell
cd task1
$env:RUN_SELF_TESTS="1"
go run .
```

Ожидаемый вывод: `Task1 Go self-tests passed`.
