package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func rateLimitMiddleware(limiter *rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Rate limit exceeded.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func proxyMiddleware(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		remote, err := url.Parse(targetURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid target URL",
			})
			return
		}

		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/go/") {
			path = strings.TrimPrefix(path, "/go")
		} else if strings.HasPrefix(path, "/py/") {
			path = strings.TrimPrefix(path, "/py")
		}
		if path == "" {
			path = "/"
		}

		proxyURL := *remote
		proxyURL.Path = path
		proxyURL.RawQuery = c.Request.URL.RawQuery

		var body io.Reader
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			body = bytes.NewReader(bodyBytes)
		}

		proxyReq, err := http.NewRequest(c.Request.Method, proxyURL.String(), body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		proxyReq.Header = c.Request.Header.Clone()
		proxyReq.Host = remote.Host

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(proxyReq)
		if err != nil {
			log.Printf("Proxy error: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "Bad gateway: " + err.Error()})
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}
		c.Status(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	}
}

func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		log.Printf("[GATEWAY] %s %s -> %d (%v)",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}

func setupRouter(goURL, pyURL string, rateLimit, burst int) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware())

	limiter := rate.NewLimiter(rate.Limit(rateLimit), burst)
	router.Use(rateLimitMiddleware(limiter))

	router.GET("/gateway/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "healthy",
			"service":    "api-gateway",
			"rate_limit": rateLimit,
			"burst":      burst,
		})
	})

	router.Any("/go/*path", proxyMiddleware(goURL))
	router.Any("/py/*path", proxyMiddleware(pyURL))

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Route not found",
			"path":  c.Request.URL.Path,
		})
	})

	return router
}

func main() {
	goURL := os.Getenv("GO_SERVICE_URL")
	if goURL == "" {
		goURL = "http://localhost:8080"
	}

	pyURL := os.Getenv("PY_SERVICE_URL")
	if pyURL == "" {
		pyURL = "http://localhost:8000"
	}

	rateLimit := 10
	burst := 20

	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	router := setupRouter(goURL, pyURL, rateLimit, burst)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("API Gateway starting on port %s", port)
		log.Printf("Go service: %s", goURL)
		log.Printf("Python service: %s", pyURL)
		log.Printf("Rate limit: %d req/sec, burst: %d", rateLimit, burst)
		log.Printf("Gateway health: http://localhost:%s/gateway/health", port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Gateway exited")
}
