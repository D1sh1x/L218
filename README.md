# Календарь событий - HTTP сервер

HTTP-сервер для работы с календарем событий, реализованный на Go.

## Возможности

- CRUD операции для событий календаря
- Получение событий по дням, неделям и месяцам
- Поддержка JSON и form-encoded запросов
- Логирование всех HTTP запросов
- Конфигурируемый порт сервера

## API Endpoints

### Создание события
```
POST /create_event
Content-Type: application/json
{
  "user_id": "1",
  "date": "2023-12-31",
  "event": "Новогодняя вечеринка"
}
```

### Обновление события
```
POST /update_event
Content-Type: application/json
{
  "id": "event-uuid",
  "user_id": "1",
  "date": "2023-12-31",
  "event": "Обновленное событие"
}
```

### Удаление события
```
POST /delete_event
Content-Type: application/x-www-form-urlencoded
id=event-uuid&user_id=1
```

### Получение событий на день
```
GET /events_for_day?user_id=1&date=2023-12-31
```

### Получение событий на неделю
```
GET /events_for_week?user_id=1&date=2023-12-31
```

### Получение событий на месяц
```
GET /events_for_month?user_id=1&date=2023-12-31
```

## Установка и запуск

### Требования
- Go 1.24.2 или выше

### Запуск
```bash
# Клонирование репозитория
git clone <repository-url>
cd L1

# Установка зависимостей
go mod tidy

# Запуск сервера
go run ./cmd/calendar

# Запуск на другом порту
go run ./cmd/calendar -port=8081

# Через переменную окружения
PORT=8081 go run ./cmd/calendar
```

## Структура проекта

```
├── cmd/
│   └── calendar/
│       └── main.go          # Точка входа приложения
├── internal/
│   ├── calendar/
│   │   ├── calendar.go      # Бизнес-логика календаря
│   │   └── calendar_test.go # Тесты
│   └── httpserver/
│       ├── server.go        # HTTP обработчики
│       └── middleware.go    # Логирование
├── go.mod
└── README.md
```

## Форматы данных

### Входные данные
- `user_id` - ID пользователя (положительное целое число)
- `date` - Дата в формате YYYY-MM-DD
- `event` - Текст события (не пустая строка)
- `id` - UUID события (для обновления/удаления)

### Ответы
Успешный ответ:
```json
{
  "result": {
    "id": "uuid",
    "user_id": 1,
    "date": "2023-12-31T00:00:00Z",
    "event": "Текст события"
  }
}
```

Ошибка:
```json
{
  "error": "описание ошибки"
}
```

## HTTP статус коды

- `200 OK` - успешное выполнение
- `400 Bad Request` - ошибки ввода (некорректные параметры)
- `503 Service Unavailable` - ошибки бизнес-логики
- `500 Internal Server Error` - внутренние ошибки сервера

## Тестирование

```bash
# Запуск всех тестов
go test ./...

# Запуск тестов с покрытием
go test -cover ./...

# Проверка кода
go vet ./...
```

## Примеры использования

### Создание события
```bash
curl -X POST http://localhost:8080/create_event \
  -H "Content-Type: application/json" \
  -d '{"user_id":"1","date":"2023-12-31","event":"Новогодняя вечеринка"}'
```

### Получение событий на день
```bash
curl "http://localhost:8080/events_for_day?user_id=1&date=2023-12-31"
```

### Обновление события
```bash
curl -X POST http://localhost:8080/update_event \
  -H "Content-Type: application/json" \
  -d '{"id":"event-uuid","user_id":"1","date":"2023-12-31","event":"Обновленное событие"}'
```

### Удаление события
```bash
curl -X POST http://localhost:8080/delete_event \
  -d "id=event-uuid&user_id=1"
```

## Особенности реализации

- События хранятся в памяти (in-memory storage)
- Потокобезопасность обеспечивается через RWMutex
- Даты нормализуются до полуночи
- Неделя считается с понедельника по воскресенье
- События сортируются по дате и ID
- Поддержка как JSON, так и form-encoded запросов
- Логирование всех HTTP запросов с временем выполнения
