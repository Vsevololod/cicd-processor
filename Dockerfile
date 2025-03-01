# Этап сборки
FROM golang:1.23.5-alpine3.21 AS builder

WORKDIR /app

# Копируем файлы зависимостей перед кодом, чтобы использовать кеш
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Передаем аргументы сборки
ARG VERSION=dev
ARG GIT_COMMIT=none
ARG BUILD_TIME=unknown

# Собираем бинарный файл с оптимизациями
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X goindex/handlers.Version=$VERSION -X goindex/handlers.GitCommit=$GIT_COMMIT -X goindex/handlers.BuildTime=$BUILD_TIME" -o /app/bin/app .

# Финальный минималистичный образ
FROM alpine:3.21.2

# Устанавливаем необходимые зависимости
RUN apk add --no-cache git docker-cli docker-compose bash

# Проверяем установки
RUN git --version && docker --version

WORKDIR /root/
COPY --from=builder /app/bin/app .

EXPOSE 8080

CMD ["./app"]
