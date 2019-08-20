package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/health", bytes.NewBuffer([]byte("Test")))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestAddToCart(t *testing.T) {
	os.Setenv("REDIS_HOST", "localhost:32768")
	router := setupRouter()
	w := httptest.NewRecorder()

	items := Item{
		Sku: "test",
		Qty: 22,
	}
	cart := Cart{
		Items: []Item{
			items,
		},
	}

	jsonpayload, _ := json.Marshal(cart)
	req, _ := http.NewRequest("POST", "/cart/sessiontest", bytes.NewBuffer(jsonpayload))
	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
	assert.Equal(t, "{\"status\":\"ok\"}\n", w.Body.String())
}

func TestGetCart(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/cart/sessiontest", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "{\"items\":[{\"sku\":\"test\",\"qty\":22}]}\n", w.Body.String())
}

func TestDeleteCart(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("DELETE", "/cart/sessiontest", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "{\"status\":\"ok\"}\n", w.Body.String())
}
