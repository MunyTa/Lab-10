package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MunyTa/Lab-10/go-service/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.DebugMode)

	router := gin.New()
	router.Use(middleware.LoggingMiddleware())
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "go-service",
		})
	})

	// GET user
	router.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"id":   id,
			"name": "User " + id,
		})
	})

	// POST user
	router.POST("/users", func(c *gin.Context) {
		var user struct {
			Name string `json:"name"`
		}
		if err := c.BindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}
		c.JSON(201, gin.H{
			"id":   "123",
			"name": user.Name,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Запуск сервера
	go func() {
		log.Printf("Go service starting on port %s", port)
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	time.Sleep(5 * time.Second)
	log.Println("Server exited")
}
