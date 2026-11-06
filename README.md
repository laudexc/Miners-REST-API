# Miners REST API

Симулятор угольного предприятия на Go. В браузере можно нанимать шахтёров, покупать оборудование и следить за добычей.

## Запуск через Docker

1. Запустите Docker Desktop.
2. В папке проекта выполните:

   ```sh
   docker compose up --build -d
   ```
3. Откройте [веб-интерфейс](http://localhost:8081). Симуляция начнётся автоматически.

Для мониторинга доступны [Grafana](http://localhost:3000) (логин и пароль `admin`) и [Prometheus](http://localhost:9090).

Чтобы остановить все сервисы, выполните `docker compose down`.

## Запуск без Docker

Для запуска приложения с веб-интерфейсом установите Go 1.25.7 или новее. В папке проекта выполните:

```sh
go run ./cmd/app
```

Откройте [веб-интерфейс](http://localhost:8080). Чтобы остановить приложение, нажмите `Ctrl+C` в терминале.

## Запросы к API

Команды запуска выше одинаковы для PowerShell, Bash, zsh и CMD. В примерах API используется порт `8081` (Docker). Для запуска через Go замените его на `8080`.

### Bash и zsh

```bash
curl "http://localhost:8081/enterprise/summary"
curl -X POST "http://localhost:8081/miners/hire?class=weak&count=1"
```

В Windows CMD используйте те же команды с `curl.exe`.

### PowerShell

```powershell
Invoke-RestMethod "http://localhost:8081/enterprise/summary"
Invoke-RestMethod -Method Post -Uri "http://localhost:8081/miners/hire?class=weak&count=1"
```

Для найма одного шахтёра класса `weak` дождитесь, пока накопится 5 угля. В браузере можно сразу открыть [сводку предприятия](http://localhost:8081/enterprise/summary). В Postman выберите метод из таблицы, добавьте путь к `http://localhost:8081` и отправьте запрос. Для этих запросов тело не требуется.

### Справочник API

| Метод | Путь | Что делает |
| --- | --- | --- |
| GET | `/health` | Проверяет доступность сервера |
| GET | `/enterprise/summary` | Показывает краткую сводку |
| GET | `/enterprise/status` | Показывает полное состояние предприятия |
| GET | `/miners/prices` | Показывает цены и характеристики шахтёров |
| POST | `/miners/hire?class=weak&count=1` | Нанимает шахтёров класса `weak`, `normal` или `strong`. Количество от 1 |
| GET | `/miners/active?limit=10` | Показывает до 10 активных шахтёров |
| GET | `/equipment/prices` | Показывает цены оборудования |
| GET | `/equipment` | Показывает оборудование и следующую доступную цель |
| POST | `/equipment/{type}/buy` | Покупает следующую цель, если хватает угля. Подставьте тип из `/equipment`, например `headlamp` |
| POST | `/enterprise/shutdown` | Останавливает симуляцию и возвращает итоговый отчёт |
| POST | `/enterprise/start` | Запускает новую симуляцию после остановки |
| GET | `/events` | Открывает поток обновлений SSE |
| GET | `/metrics` | Возвращает метрики Prometheus |
| PUT | `/app/close` | Завершает приложение |

## Тесты

```sh
go test ./...
```
