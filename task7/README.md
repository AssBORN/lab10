# Task 7 (ЛР10): Graceful shutdown в Go и Python

В папке два независимых сервиса:

- `go-service` (Gin + `http.Server`): корректно завершает работу по `SIGINT/SIGTERM`.
- `python-service` (FastAPI + Uvicorn): обрабатывает `SIGINT/SIGTERM` и дает запросам завершиться.

## 1) Запуск Go-сервиса

```powershell
cd task7\go-service
go mod tidy
go run .
```

Go слушает `http://127.0.0.1:8081`.

## 2) Запуск Python-сервиса

```powershell
cd task7\python-service
py -3 -m pip install -r requirements.txt
py -3 main.py
```

Python слушает `http://127.0.0.1:8001`.

## 3) Проверка graceful shutdown

1. Запусти долгий запрос (в отдельном терминале):

```powershell
Invoke-RestMethod -Uri "http://127.0.0.1:8081/work?seconds=10"
```

или

```powershell
Invoke-RestMethod -Uri "http://127.0.0.1:8001/work?seconds=10"
```

2. Пока запрос выполняется, отправь `Ctrl+C` в терминал сервиса.
3. Ожидаемое поведение:
   - сервис прекращает принимать новые соединения;
   - текущий запрос корректно завершается;
   - процесс завершает работу без "жесткого" обрыва.

## 4) Встроенные тесты внутри main-файлов

**Go (`go-service/main.go`)**:

```powershell
cd task7\go-service
$env:RUN_SELF_TESTS="1"
go run .
```

**Python (`python-service/main.py`)**:

```powershell
cd task7\python-service
$env:RUN_SELF_TESTS="1"
py -3 main.py
```
