FROM golang:1.23.6-alpine AS builder
WORKDIR /app
COPY go.mod main.go ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o pinger

FROM scratch
COPY --from=builder /app/pinger /app/pinger
ENTRYPOINT ["/app/pinger"]
