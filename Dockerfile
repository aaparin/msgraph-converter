# Этап сборки
FROM golang:1.23.3-alpine AS builder

# Устанавливаем необходимые зависимости для сборки
RUN apk add --no-cache git make gcc musl-dev

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
# CGO_ENABLED=0 отключает CGO для статической компиляции
# -ldflags="-w -s" уменьшает размер бинарного файла
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o msgraph-converter .

# Этап финального образа
FROM alpine:3.19

# Добавляем пользователя для безопасности
RUN adduser -D -g '' appuser

# Создаем необходимые директории
RUN mkdir /app && chown appuser:appuser /app

# Копируем файлы конфигурации и бинарный файл
COPY --from=builder /app/msgraph-converter /app/
COPY --from=builder /app/.env /app/

# Устанавливаем рабочую директорию
WORKDIR /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Проверяем наличие переменной окружения SERVICE_PORT
ARG SERVICE_PORT
ENV SERVICE_PORT=${SERVICE_PORT:-8181}

# Объявляем порт, который будет использоваться приложением
EXPOSE ${SERVICE_PORT}

# Запускаем приложение
CMD ["./msgraph-converter"]

# Добавляем метаданные
LABEL maintainer="Your Name <your.email@example.com>"
LABEL version="1.0"
LABEL description="Microsoft Graph Document Converter Service"