package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func registerRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up", "service": "go"})
	})

	r.GET("/work", func(c *gin.Context) {
		secondsRaw := c.DefaultQuery("seconds", "10")
		seconds, err := strconv.Atoi(secondsRaw)
		if err != nil || seconds < 0 || seconds > 120 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "seconds must be integer in range 0..120"})
			return
		}

		// Simulate long-running work and allow graceful completion.
		time.Sleep(time.Duration(seconds) * time.Second)
		c.JSON(http.StatusOK, gin.H{
			"status":         "done",
			"service":        "go",
			"work_seconds":   seconds,
			"finished_at_utc": time.Now().UTC().Format(time.RFC3339),
		})
	})
}

func must(ok bool, message string) {
	if !ok {
		panic(message)
	}
}

func runSelfTests() {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerRoutes(r)

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthRec := httptest.NewRecorder()
	r.ServeHTTP(healthRec, healthReq)
	must(healthRec.Code == http.StatusOK, "health status must be 200")

	var healthBody map[string]any
	if err := json.Unmarshal(healthRec.Body.Bytes(), &healthBody); err != nil {
		panic(err)
	}
	must(healthBody["status"] == "up", "health status field must be up")

	workReq := httptest.NewRequest(http.MethodGet, "/work?seconds=0", nil)
	workRec := httptest.NewRecorder()
	r.ServeHTTP(workRec, workReq)
	must(workRec.Code == http.StatusOK, "work status must be 200")

	var workBody map[string]any
	if err := json.Unmarshal(workRec.Body.Bytes(), &workBody); err != nil {
		panic(err)
	}
	must(workBody["status"] == "done", "work status field must be done")
	must(workBody["service"] == "go", "work service field must be go")

	invalidReq := httptest.NewRequest(http.MethodGet, "/work?seconds=abc", nil)
	invalidRec := httptest.NewRecorder()
	r.ServeHTTP(invalidRec, invalidReq)
	must(invalidRec.Code == http.StatusBadRequest, "invalid seconds must return 400")

	var invalidBody map[string]any
	if err := json.Unmarshal(invalidRec.Body.Bytes(), &invalidBody); err != nil {
		panic(err)
	}
	errMsg, _ := invalidBody["error"].(string)
	must(strings.Contains(errMsg, "seconds"), "invalid error message must mention seconds")

	fmt.Println("Go task7 self-tests passed")
}

func main() {
	if os.Getenv("RUN_SELF_TESTS") == "1" {
		runSelfTests()
		return
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	registerRoutes(router)

	srv := &http.Server{
		Addr:         ":8081",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 130 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("go-service listening on :8081")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutdown signal received, stopping go-service gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		_ = srv.Close()
	}
	log.Println("go-service stopped")
}
