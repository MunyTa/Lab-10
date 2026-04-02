package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MunyTa/Lab-10/go-service/middleware"
	ws "github.com/MunyTa/Lab-10/go-service/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var chatHub = ws.NewHub()

func main() {
	// Запускаем WebSocket hub
	go chatHub.Run()

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
			Name string `json:"name" binding:"required"`
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

	// WebSocket chat endpoint
	router.GET("/ws/chat", func(c *gin.Context) {
		room := c.Query("room")
		user := c.Query("user")

		if room == "" || user == "" {
			c.JSON(400, gin.H{"error": "room and user parameters required"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		ws.HandleWebSocket(chatHub, conn, room, user)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Запуск сервера
	go func() {
		log.Printf("Go service starting on port %s", port)
		log.Printf("WebSocket chat available at ws://localhost:%s/ws/chat?room=test&user=alice", port)
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
