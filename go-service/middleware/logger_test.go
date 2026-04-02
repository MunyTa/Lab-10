package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoggingMiddleware_LogsRequestDetails проверяет наличие метода, пути, статуса в логах
func TestLoggingMiddleware_LogsRequestDetails(t *testing.T) {
	// Подменяем stdout для захвата логов
	var buf bytes.Buffer
	gin.DefaultWriter = &buf
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что лог содержит JSON с нужными полями
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Contains(t, logEntry, "method")
	assert.Contains(t, logEntry, "path")
	assert.Contains(t, logEntry, "status")
	assert.Contains(t, logEntry, "duration_ms")
	assert.Contains(t, logEntry, "client_ip")
	assert.Contains(t, logEntry, "user_agent")

	assert.Equal(t, "GET", logEntry["method"])
	assert.Equal(t, "/test", logEntry["path"])
	assert.Equal(t, float64(http.StatusOK), logEntry["status"])
}

// TestLoggingMiddleware_LogsExecutionTime проверяет, что время выполнения > 0
func TestLoggingMiddleware_LogsExecutionTime(t *testing.T) {
	var buf bytes.Buffer
	gin.DefaultWriter = &buf
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/slow", func(c *gin.Context) {
		time.Sleep(100 * time.Millisecond)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	durationMs, ok := logEntry["duration_ms"].(float64)
	assert.True(t, ok)
	assert.Greater(t, durationMs, 90.0) // Должно быть около 100ms
	assert.Less(t, durationMs, 150.0)
}

// TestLoggingMiddleware_LogsClientIP проверяет логирование IP клиента
func TestLoggingMiddleware_LogsClientIP(t *testing.T) {
	var buf bytes.Buffer
	gin.DefaultWriter = &buf
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/ip", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/ip", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "192.168.1.100", logEntry["client_ip"])
}

// TestLoggingMiddleware_LogsUserAgent проверяет логирование User-Agent
func TestLoggingMiddleware_LogsUserAgent(t *testing.T) {
	var buf bytes.Buffer
	gin.DefaultWriter = &buf
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/ua", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/ua", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "TestAgent/1.0", logEntry["user_agent"])
}

// TestLoggingMiddleware_AppliedToAllRoutes проверяет, что middleware вызывается для всех маршрутов
func TestLoggingMiddleware_AppliedToAllRoutes(t *testing.T) {
	var buf bytes.Buffer
	gin.DefaultWriter = &buf
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/route1", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.POST("/route2", func(c *gin.Context) { c.Status(http.StatusCreated) })
	router.PUT("/route3", func(c *gin.Context) { c.Status(http.StatusAccepted) })

	testCases := []struct {
		method string
		path   string
	}{
		{"GET", "/route1"},
		{"POST", "/route2"},
		{"PUT", "/route3"},
	}

	for _, tc := range testCases {
		buf.Reset()
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, tc.method, logEntry["method"])
		assert.Equal(t, tc.path, logEntry["path"])
	}
}

// TestLoggingMiddleware_JSONFormat проверяет валидный JSON вывод
func TestLoggingMiddleware_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	gin.DefaultWriter = &buf
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/json-test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/json-test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Проверяем, что вывод - валидный JSON
	var logEntry interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)
	assert.NotNil(t, logEntry)

	// Проверяем, что каждая запись заканчивается новой строкой
	logLines := bytes.Split(buf.Bytes(), []byte("\n"))
	for _, line := range logLines {
		if len(line) > 0 {
			assert.True(t, json.Valid(line), "Каждая строка должна быть валидным JSON")
		}
	}
}

// TestLoggingMiddleware_ErrorResponse проверяет логирование ошибок (статус 4xx, 5xx)
func TestLoggingMiddleware_ErrorResponse(t *testing.T) {
	var buf bytes.Buffer
	gin.DefaultWriter = &buf
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/notfound", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
	router.GET("/servererror", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	})

	testCases := []struct {
		path   string
		status int
	}{
		{"/notfound", http.StatusNotFound},
		{"/servererror", http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		buf.Reset()
		req := httptest.NewRequest("GET", tc.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, float64(tc.status), logEntry["status"])
		assert.Contains(t, logEntry, "duration_ms")
	}
}
