package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateProduct(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()

	p := Product{
		SKU:         "SKU1",
		Name:        "test1",
		Price:       22,
		Description: "used for testing",
	}
	pjson, _ := json.Marshal(p)

	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(pjson))
	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
}

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
