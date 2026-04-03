package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SumRequest struct {
	A float64 `json:"a" binding:"required"`
	B float64 `json:"b" binding:"required"`
}

func registerRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "up",
			"service": "task1-go-api",
		})
	})

	r.GET("/hello/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, " + name + "!",
		})
	})

	r.POST("/sum", func(c *gin.Context) {
		var req SumRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"a":      req.A,
			"b":      req.B,
			"result": req.A + req.B,
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

	helloReq := httptest.NewRequest(http.MethodGet, "/hello/Alice", nil)
	helloRec := httptest.NewRecorder()
	r.ServeHTTP(helloRec, helloReq)
	must(helloRec.Code == http.StatusOK, "hello status must be 200")

	var helloBody map[string]any
	if err := json.Unmarshal(helloRec.Body.Bytes(), &helloBody); err != nil {
		panic(err)
	}
	must(helloBody["message"] == "Hello, Alice!", "hello message mismatch")

	sumReqBody := []byte(`{"a":2.5,"b":7.5}`)
	sumReq := httptest.NewRequest(http.MethodPost, "/sum", bytes.NewReader(sumReqBody))
	sumReq.Header.Set("Content-Type", "application/json")
	sumRec := httptest.NewRecorder()
	r.ServeHTTP(sumRec, sumReq)
	must(sumRec.Code == http.StatusOK, "sum status must be 200")

	var sumBody map[string]any
	if err := json.Unmarshal(sumRec.Body.Bytes(), &sumBody); err != nil {
		panic(err)
	}
	must(sumBody["status"] == "ok", "sum status field must be ok")
	must(sumBody["result"] == 10.0, "sum result must be 10")

	badReq := httptest.NewRequest(http.MethodPost, "/sum", bytes.NewReader([]byte(`{"a":"x","b":2}`)))
	badReq.Header.Set("Content-Type", "application/json")
	badRec := httptest.NewRecorder()
	r.ServeHTTP(badRec, badReq)
	must(badRec.Code == http.StatusBadRequest, "bad sum input must return 400")

	fmt.Println("Task1 Go self-tests passed")
}

func main() {
	if os.Getenv("RUN_SELF_TESTS") == "1" {
		runSelfTests()
		return
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	registerRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	if _, err := strconv.Atoi(port); err != nil {
		panic("PORT must be numeric")
	}

	_ = r.Run(":" + port)
}
