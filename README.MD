
```markdown
# Load Balancer (Go Implementation)



Простой HTTP балансировщик нагрузки с:
- Weighted Round-Robin алгоритмом
- Проверкой здоровья бэкендов
- Ограничением запросов (Rate Limiting)
- REST API для управления

## Быстрый старт

1. Установите зависимости:
```bash
go mod download
```

2. Создайте конфиг (`config.yaml`):
```yaml
port: 8080
backends:
  - url: "http://localhost:8001"
    weight: 100
  - url: "http://localhost:8002"
    weight: 200
health_tick: 12
rate_limiting:
  default_capacity: 10  # Макс. количество запросов
  default_rate: 1       # 10 запросов в секунду
database:
  postgres:
    host: "localhost"
    port: 8010
    user: "postgres"
    password: "postgres"
    dbname: "postgres"
    sslmode: "disable"
```

3. Запустите бэкенды (в новых терминалах):
```bash
# Бэкенд 1 (порт 8081)
go run test_servers/backend1.go

# Бэкенд 2 (порт 8082)
go run test_servers/backend2.go
```

4. Запустите балансировщик:
```bash
go run .
```

## API Endpoints

| Метод  | Путь               | Тело запроса                              |
|--------|--------------------|-------------------------------------------|
| POST   | `/clients`         | `{"id":"test","capacity":100,"rate":10}`  |
| GET    | `/clients/{id}`    | -                                         |
| PUT    | `/clients`         | `{"id":"test","capacity":200,"rate":20}`  |
| DELETE | `/clients/{id}`    | -                                         |

## Примеры запросов

Создать клиента:
```bash
curl -X POST http://localhost:8080/clients \
  -H "Content-Type: application/json" \
  -d '{"id":"test1","capacity":100,"rate_per_second":10}'
```

Проверить балансировку:
```bash
for i in {1..5}; do curl http://localhost:8080; done
```

## Структура проекта

```
.
├── cmd/           # Главный исполняемый файл
├── configs/       # Конфигурация
├── internal/      # Основная логика
│   ├── balancer/  # Логика балансировки
│   ├── server/    # HTTP обработчики
│   └── ...        # Другие модули
├── test_servers/  # Тестовые бэкенды
└── README.md
```

## Тестирование

Запуск тестов:
```bash
go test ./...
```

Тестовый клиент:
```bash
go run test/api_client/main.go
```
---
```

Для работы вам понадобятся:
1. Файл `test_servers/backend1.go`:
```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Backend 1 response")
	})
	http.ListenAndServe(":8081", nil)
}
```

2. Файл `test_servers/backend2.go`:
```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Backend 2 response")
	})
	http.ListenAndServe(":8082", nil)
}
```
