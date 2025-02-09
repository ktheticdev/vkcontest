# Строим бинарник в golang контейнере
FROM golang:1.23.6-alpine AS builder
WORKDIR /app
COPY go.mod main.go ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main

# Используем scratch для минимального образа
FROM scratch
COPY --from=builder /app/main /app/main
EXPOSE 8080
ENTRYPOINT ["/app/main"]
