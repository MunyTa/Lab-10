package middleware

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
)

// LogEntry представляет структуру лога
type LogEntry struct {
	Timestamp    string  `json:"timestamp"`
	Method       string  `json:"method"`
	Path         string  `json:"path"`
	Status       int     `json:"status"`
	DurationMs   float64 `json:"duration_ms"`
	ClientIP     string  `json:"client_ip"`
	UserAgent    string  `json:"user_agent"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

// LoggingMiddleware создает middleware для логирования запросов в JSON формате
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Засекаем время начала запроса
		startTime := time.Now()

		// Обрабатываем запрос
		c.Next()

		// Вычисляем длительность выполнения
		duration := time.Since(startTime)
		durationMs := float64(duration.Nanoseconds()) / 1e6

		// Получаем IP клиента (учитывая прокси)
		clientIP := c.ClientIP()

		// Получаем User-Agent
		userAgent := c.Request.UserAgent()

		// Создаем запись лога
		logEntry := LogEntry{
			Timestamp:  time.Now().Format(time.RFC3339),
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			Status:     c.Writer.Status(),
			DurationMs: durationMs,
			ClientIP:   clientIP,
			UserAgent:  userAgent,
		}

		// Если есть ошибки в контексте Gin, добавляем их в лог
		if len(c.Errors) > 0 {
			logEntry.ErrorMessage = c.Errors.String()
		}

		// Сериализуем в JSON и выводим в stdout
		jsonLog, err := json.Marshal(logEntry)
		if err != nil {
			// В случае ошибки сериализации логируем как есть
			gin.DefaultWriter.Write([]byte("failed to marshal log entry: " + err.Error() + "\n"))
			return
		}

		// Добавляем новую строку для каждого лога
		gin.DefaultWriter.Write(append(jsonLog, '\n'))
	}
}
