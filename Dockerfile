# Используем многоэтапную сборку
# Этап 1: Сборка приложения
FROM golang:1.25-alpine AS builder

# Устанавливаем ТОЛЬКО необходимые зависимости для Go
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN GOOS=linux go build -o /app/bot ./cmd/bot/main.go

# Этап 2: Финальный образ
FROM alpine:3.20

# Устанавливаем необходимые зависимости
RUN apk add --no-cache \
    ca-certificates \
    tzdata

WORKDIR /app

# Копируем скомпилированный бинарник
COPY --from=builder /app/bot .

# Копируем JSON файлы
COPY --from=builder /app/jsons ./jsons

# Копируем Prisma schema и миграции (если нужны)
COPY --from=builder /app/prisma ./prisma

# Создаем пользователя для безопасности
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup && \
    chown -R appuser:appgroup /app

USER appuser

# Команда для запуска
CMD ["./bot"]