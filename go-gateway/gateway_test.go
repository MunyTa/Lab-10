package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouting_GoService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем мок Go сервиса
	mockGo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/123", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "123", "name": "Test User"})
	}))
	defer mockGo.Close()

	// Создаем мок Python сервиса
	mockPy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer mockPy.Close()

	router := setupRouter(mockGo.URL, mockPy.URL, 10, 20)

	req := httptest.NewRequest("GET", "/go/users/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "123", response["id"])
}

func TestRouting_PyService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer mockGo.Close()

	mockPy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/users/123", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "123", "name": "Python User"})
	}))
	defer mockPy.Close()

	router := setupRouter(mockGo.URL, mockPy.URL, 10, 20)

	req := httptest.NewRequest("GET", "/py/api/users/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouting_404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer mockGo.Close()

	mockPy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer mockPy.Close()

	router := setupRouter(mockGo.URL, mockPy.URL, 10, 20)

	req := httptest.NewRequest("GET", "/invalid/path", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRateLimiting_ExceedsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockGo.Close()

	mockPy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer mockPy.Close()

	router := setupRouter(mockGo.URL, mockPy.URL, 2, 3) // Лимит: 2 запроса в секунду

	// Делаем 5 запросов подряд
	limitCount := 0

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/go/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusTooManyRequests {
			limitCount++
		}
	}

	assert.Greater(t, limitCount, 0, "Должны быть запросы, отклоненные rate limiter'ом")
}

func TestGateway_LoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockGo.Close()

	mockPy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer mockPy.Close()

	router := setupRouter(mockGo.URL, mockPy.URL, 10, 20)

	req := httptest.NewRequest("GET", "/go/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
