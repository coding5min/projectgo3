package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Person struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Файл .env не найден, читаем переменные окружения из системы")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL не задан")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	dbpool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatal("ошибка подключений PosgreSQL", err)
	}
	defer dbpool.Close()

	err = dbpool.Ping(context.Background())
	if err != nil {
		log.Fatal("PostgreSQL не отвечает:", err)
	}

	log.Println("PostgreSQL подключен")

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"mesage": "Go Gin Api работает",
		})
	})

	r.GET("/persons", func(c *gin.Context) {
		rows, err := dbpool.Query(
			context.Background(),
			`SELECT id, name, age FROM persons ORDER BY id`,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer rows.Close()

		persons := []Person{}

		for rows.Next() {
			var p Person

			err := rows.Scan(&p.ID, &p.Name, &p.Age)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			persons = append(persons, p)
		}

		c.JSON(http.StatusOK, persons)
	})

	log.Println("Сервер запущен на порту :", port)

	err = r.Run(":" + port)
	if err != nil {
		log.Fatal("Oшибка запуска сервера", err)
	}
}
