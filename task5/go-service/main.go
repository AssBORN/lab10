package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Geo struct {
	Lat float64 `json:"lat" binding:"required"`
	Lon float64 `json:"lon" binding:"required"`
}

type Address struct {
	City   string `json:"city" binding:"required"`
	Street string `json:"street" binding:"required"`
	Geo    Geo    `json:"geo" binding:"required"`
}

type Contact struct {
	Type  string `json:"type" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type Item struct {
	SKU      string   `json:"sku" binding:"required"`
	Name     string   `json:"name" binding:"required"`
	Qty      int      `json:"qty" binding:"required,min=1"`
	Price    float64  `json:"price" binding:"required,gt=0"`
	Tags     []string `json:"tags"`
	Metadata any      `json:"metadata"`
}

type Payment struct {
	Method   string         `json:"method" binding:"required"`
	Currency string         `json:"currency" binding:"required"`
	Paid     bool           `json:"paid"`
	Extra    map[string]any `json:"extra"`
}

type OrderRequest struct {
	RequestID string            `json:"request_id" binding:"required"`
	UserID    int               `json:"user_id" binding:"required,gt=0"`
	Name      string            `json:"name" binding:"required"`
	Address   Address           `json:"address" binding:"required"`
	Contacts  []Contact         `json:"contacts" binding:"required,min=1,dive"`
	Items     []Item            `json:"items" binding:"required,min=1,dive"`
	Payment   Payment           `json:"payment" binding:"required"`
	Flags     map[string]bool   `json:"flags"`
	Metadata  map[string]string `json:"metadata"`
}

func registerRoutes(r *gin.Engine) {
	r.POST("/process", func(c *gin.Context) {
		var req OrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		var total float64
		totalQty := 0
		for _, item := range req.Items {
			total += item.Price * float64(item.Qty)
			totalQty += item.Qty
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       "ok",
			"processed_by": "go-service",
			"request_id":   req.RequestID,
			"user_id":      req.UserID,
			"city":         req.Address.City,
			"items_count":  len(req.Items),
			"total_qty":    totalQty,
			"total_price":  total,
			"paid":         req.Payment.Paid,
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up"})
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

	validPayload := map[string]any{
		"request_id": "REQ-SELFTEST",
		"user_id":    7,
		"name":       "Self Test",
		"address": map[string]any{
			"city":   "Moscow",
			"street": "Tverskaya 1",
			"geo":    map[string]any{"lat": 55.75, "lon": 37.61},
		},
		"contacts": []map[string]string{{"type": "email", "value": "test@example.com"}},
		"items": []map[string]any{
			{"sku": "A", "name": "Book", "qty": 2, "price": 100.0},
			{"sku": "B", "name": "Pen", "qty": 3, "price": 50.0},
		},
		"payment": map[string]any{"method": "card", "currency": "RUB", "paid": true},
	}
	raw, err := json.Marshal(validPayload)
	if err != nil {
		panic(err)
	}
	processReq := httptest.NewRequest(http.MethodPost, "/process", bytes.NewReader(raw))
	processReq.Header.Set("Content-Type", "application/json")
	processRec := httptest.NewRecorder()
	r.ServeHTTP(processRec, processReq)
	must(processRec.Code == http.StatusOK, "process status must be 200")

	var processBody map[string]any
	if err := json.Unmarshal(processRec.Body.Bytes(), &processBody); err != nil {
		panic(err)
	}
	must(processBody["status"] == "ok", "process body status must be ok")
	must(processBody["total_qty"] == float64(5), "total_qty must be 5")
	must(processBody["total_price"] == 350.0, "total_price must be 350")

	invalidRaw := []byte(`{"request_id":"","user_id":0}`)
	invalidReq := httptest.NewRequest(http.MethodPost, "/process", bytes.NewReader(invalidRaw))
	invalidReq.Header.Set("Content-Type", "application/json")
	invalidRec := httptest.NewRecorder()
	r.ServeHTTP(invalidRec, invalidReq)
	must(invalidRec.Code == http.StatusBadRequest, "invalid process status must be 400")

	var invalidBody map[string]any
	if err := json.Unmarshal(invalidRec.Body.Bytes(), &invalidBody); err != nil {
		panic(err)
	}
	errStr, _ := invalidBody["error"].(string)
	must(strings.Contains(errStr, "Error:"), "invalid response must contain validation message")

	fmt.Println("Go self-tests passed")
}

func main() {
	if os.Getenv("RUN_SELF_TESTS") == "1" {
		runSelfTests()
		return
	}

	r := gin.Default()
	registerRoutes(r)
	_ = r.Run(":8080")
}
