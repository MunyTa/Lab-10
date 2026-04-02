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

	mockPy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
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
	assert.Equal(t, "Test User", response["name"])
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

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "123", response["id"])
	assert.Equal(t, "Python User", response["name"])
}

func TestRateLimiting_ExceedsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockGo.Close()

	mockPy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer mockPy.Close()

	// Лимит: 2 запроса в секунду, burst 3
	router := setupRouter(mockGo.URL, mockPy.URL, 2, 3)

	successCount := 0
	limitCount := 0

	// Делаем 10 быстрых запросов подряд
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/go/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			successCount++
		} else if w.Code == http.StatusTooManyRequests {
			limitCount++
		}
	}

	assert.Greater(t, limitCount, 0, "Должны быть запросы, отклоненные rate limiter'ом")
	assert.LessOrEqual(t, successCount, 5, "Не более 5 успешных запросов (с учетом burst)")
}

func TestGatewayHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer mockGo.Close()

	mockPy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer mockPy.Close()

	router := setupRouter(mockGo.URL, mockPy.URL, 10, 20)

	req := httptest.NewRequest("GET", "/gateway/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "api-gateway", response["service"])
}

func TestNotFound(t *testing.T) {
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
