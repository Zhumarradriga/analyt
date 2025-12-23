# Dockerfile
# Сервис аналитики → ClickHouse → порт 8100

# ---- Stage 1: сборка ----
    FROM golang:1.25.3-alpine AS builder

    # Установим git (для go mod)
    RUN apk add --no-cache git
    
    WORKDIR /app
    
    # Кэшируем зависимости
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Копируем исходный код
    COPY . .
    
    # Собираем бинарь (статически, без CGO)
    RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o analytics_service ./cmd/api/main.go
    
    # ---- Stage 2: финальный образ ----
    FROM alpine:3.20
    
    # Установим только необходимое
    RUN apk --no-cache add ca-certificates tzdata
    
    # Создадим непривилегированного пользователя
    RUN adduser -D -s /bin/sh appuser
    
    WORKDIR /app
    
    # Копируем бинарь
    COPY --from=builder /app/analytics_service .
    
    # Меняем владельца
    RUN chown appuser:appuser ./analytics_service
    
    # Переключаемся на непривилегированного пользователя
    USER appuser
    
    # Порт сервиса
    EXPOSE 8100
    
    # Запуск
    CMD ["./analytics_service"]