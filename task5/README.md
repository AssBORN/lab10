# Task 5 (ЛР10): Передача сложного JSON между сервисами

В папке находятся 2 сервиса:

- `go-service` (Gin): принимает сложный JSON на `POST /process`, валидирует и возвращает агрегированный результат.
- `python-service` (FastAPI): принимает такой же JSON на `POST /forward` и пересылает его в Go-сервис.

## 1) Запуск Go-сервиса

```powershell
cd task5\go-service
go mod tidy
go run .
```

Сервис слушает `http://127.0.0.1:8080`.

## 2) Запуск Python-сервиса

```powershell
cd task5\python-service
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -r requirements.txt
uvicorn main:app --reload --port 8000
```

Сервис слушает `http://127.0.0.1:8000`.

## 3) Проверка передачи сложного JSON

Из корня `task5`:

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "http://127.0.0.1:8000/forward" `
  -ContentType "application/json" `
  -InFile ".\sample_payload.json"
```

Ожидаемый результат: ответ от Python-сервиса с полем `go_response`, которое содержит результат обработки на стороне Go.

## 4) Тесты внутри main-файлов

**Go (`main.go`)**:

```powershell
cd task5\go-service
$env:RUN_SELF_TESTS="1"
go run .
```

**Python (`main.py`)**:

```powershell
cd task5\python-service
py -3 main.py
```

Оба запуска выполняют встроенные self-tests напрямую из `main.go` и `main.py`.
