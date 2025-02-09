package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Status struct {
	IP            string    `json:"ip"`
	PingTime      int64     `json:"ping_time"`
	LastSuccessAt time.Time `json:"last_success_at"`
}

var dbPool *pgxpool.Pool

func initDB(ctx context.Context, dbURL string) {
	var err error
	dbPool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS statuses (
		id SERIAL PRIMARY KEY,
		ip TEXT NOT NULL,
		ping_time BIGINT NOT NULL,
		last_success_at TIMESTAMP NOT NULL
	);
	`
	_, err = dbPool.Exec(ctx, createTable)
	if err != nil {
		log.Fatalf("Не удалось создать таблицу: %v", err)
	}
}

func getStatuses(c *gin.Context) {
	rows, err := dbPool.Query(context.Background(), "SELECT ip, ping_time, last_success_at FROM statuses")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var statuses []Status
	for rows.Next() {
		var s Status
		if err := rows.Scan(&s.IP, &s.PingTime, &s.LastSuccessAt); err != nil {
			continue
		}
		statuses = append(statuses, s)
	}
	c.JSON(http.StatusOK, statuses)
}

func addStatus(c *gin.Context) {
	var s Status
	if err := c.BindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}
	_, err := dbPool.Exec(context.Background(),
		"INSERT INTO statuses (ip, ping_time, last_success_at) VALUES ($1, $2, $3)",
		s.IP, s.PingTime, s.LastSuccessAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка записи в БД"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Данные добавлены"})
}

func main() {
	port := "8080"
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Переменная окружения DATABASE_URL не установлена")
	}

	ctx := context.Background()
	initDB(ctx, dbURL)

	router := gin.Default()
	router.GET("/statuses", getStatuses)
	router.POST("/statuses", addStatus)

	log.Printf("Backend-сервис запущен на порту %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
