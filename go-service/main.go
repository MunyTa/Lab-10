package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MunyTa/Lab-10/go-service/middleware"
	"github.com/gin-gonic/gin"

	// Swagger imports
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Документация (будет сгенерирована)
	_ "github.com/MunyTa/Lab-10/go-service/docs"
)

// @title           Go Service API
// @version         1.0
// @description     Микросервис на Go (Gin) для управления пользователями
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@lab10.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @schemes   http
func main() {
	gin.SetMode(gin.DebugMode)

	router := gin.New()
	router.Use(middleware.LoggingMiddleware())
	router.Use(gin.Recovery())

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// @Summary     Health check
	// @Description Проверка работоспособности сервиса
	// @Tags        health
	// @Produce     json
	// @Success     200 {object} map[string]string "status ok"
	// @Router      /health [get]
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "go-service",
		})
	})

	// @Summary     Get user by ID
	// @Description Возвращает пользователя по указанному ID
	// @Tags        users
	// @Produce     json
	// @Param       id   path      string true "User ID"
	// @Success     200  {object} map[string]string "user data"
	// @Failure     404  {object} map[string]string "user not found"
	// @Router      /users/{id} [get]
	router.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"id":   id,
			"name": "User " + id,
		})
	})

	// @Summary     Create user
	// @Description Создаёт нового пользователя
	// @Tags        users
	// @Accept      json
	// @Produce     json
	// @Param       request body     map[string]string true "User data"
	// @Success     201    {object} map[string]string "created user"
	// @Failure     400    {object} map[string]string "bad request"
	// @Router      /users [post]
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Запуск сервера
	go func() {
		log.Printf("Go service starting on port %s", port)
		log.Printf("Swagger UI available at http://localhost:%s/swagger/index.html", port)
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
