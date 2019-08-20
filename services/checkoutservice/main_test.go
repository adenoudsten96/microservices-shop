package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckout(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()

	ck := Checkout{
		SessionID:  "sessiontest",
		Address:    "testlane 1",
		Email:      "test@test.com",
		Creditcard: "123-456-789cc",
	}
	ckjson, _ := json.Marshal(ck)

	req, _ := http.NewRequest("POST", "/checkout", bytes.NewBuffer(ckjson))
	router.ServeHTTP(w, req)

	type CheckoutResponse struct {
		ShippingID    string `json:"shippingid"`
		TransactionID string `json:"transactionid"`
	}
	var cr CheckoutResponse
	_ = json.Unmarshal([]byte(w.Body.String()), &cr)

	// Use RegEx to match a UUID
	match, _ := regexp.MatchString("[0-9a-fA-F]{8}\\-[0-9a-fA-F]{4}\\-[0-9a-fA-F]{4}\\-[0-9a-fA-F]{4}\\-[0-9a-fA-F]{12}", cr.ShippingID)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, true, match)
}
