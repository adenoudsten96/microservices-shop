package main

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetAllProducts(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/product", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestGetProduct(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/product/SKU1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCreateProduct(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()

	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyz")
	length := 8
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	p := Product{
		SKU:         b.String(),
		Name:        "test1",
		Price:       22,
		Description: "used for testing",
	}
	pjson, _ := json.Marshal(p)

	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(pjson))
	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
}
