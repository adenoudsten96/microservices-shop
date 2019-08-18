package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	cartservice     = os.Getenv("CARTSERVICE")
	emailservice    = os.Getenv("EMAILSERVICE")
	paymentservice  = os.Getenv("PAYMENTSERVICE")
	shippingservice = os.Getenv("SHIPPINGSERVICE")
	productservice  = os.Getenv("PRODUCTSERVICE")
)

// Checkout represents the information required to perform a succesful checkout.
type Checkout struct {
	SessionID  string `json:"sessionid" binding:"required"`
	Address    string `json:"address" binding:"required"`
	Email      string `json:"email" binding:"required"`
	Creditcard string `json:"creditcard" binding:"required"`
}

// Cart represents the shopping cart model
type Cart struct {
	Items []Item `json:"items"`
}

// Item represents the items in a Cart
type Item struct {
	Sku string `json:"sku"`
	Qty int    `json:"qty"`
}

// Ship represents the payload to the shipping service
type Ship struct {
	Address string `json:"address"`
	Items   []Item `json:"items"`
}

// getShoppingCart calls cartservice to get a shopping cart. Returns type Cart, which may be empty.
func getShoppingCart(sessionid string) (Cart, error) {

	// Make the request to the cart service
	log.Println("Calling service cartservice...")
	url := fmt.Sprintf("%v/cart/%v", cartservice, sessionid)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err := errors.New("failed to get shopping cart")
		return Cart{}, err
	}

	// Read HTTP body
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	// Unmarshal the JSON into our Cart struct
	var cart Cart
	err = json.Unmarshal(result, &cart)
	if err != nil {
		log.Println(err)
	}

	// Return the shopping cart
	return cart, err
}

// payProduct calls paymentservice to charge a creditcard. Returns the transaction id as a string.
func payProduct(creditcard string, amount int) (string, error) {

	// Make the request to the payment service
	log.Println("Calling service paymentservice...")
	url := fmt.Sprintf("%v/payment", paymentservice)
	payload := []byte(fmt.Sprintf(`{"creditcard": "%v", "amount": %v}`, creditcard, amount))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err := errors.New("failed to process payment")
		return "", err
	}

	// Read HTTP body
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	// Unmarshal results
	type PaymentResponse struct {
		TransactionID string `json:"transactionid"`
	}
	var paymentresponse PaymentResponse
	json.Unmarshal(result, &paymentresponse)

	return paymentresponse.TransactionID, err
}

// shipProduct calls shipservice to ship the products to a user. Returns the shipping id as a string.
func shipProduct(address string, products []Item) (string, error) {

	// Convert products to our Ship struct
	ship := Ship{
		Address: address,
		Items:   products,
	}

	// Make the request to the shipping service
	log.Println("Calling service shippingservice...")
	url := fmt.Sprintf("%v/ship", shippingservice)
	payload, _ := json.Marshal(ship)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err := errors.New("failed to process shipment")
		return "", err
	}

	// Read HTTP body
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	// Unmarshal results
	type ShipResponse struct {
		ShippingID string `json:"shippingid"`
	}
	var shipresponse ShipResponse
	json.Unmarshal(result, &shipresponse)

	return shipresponse.ShippingID, err

}

// sendEmail calls the emailservice to send the user an order confirmation email.
func sendEmail(email string) error {
	// Make the request to the email service
	log.Println("Calling service emailservice...")
	url := fmt.Sprintf("%v/email", emailservice)
	payload := []byte(fmt.Sprintf(`{"email": "%v"}`, email))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err := errors.New("failed to send confirmation email")
		return err
	}

	return err
}

// checkout orchestrates the checkout process.
func checkout(c *gin.Context) {

	// Get the JSON data
	var checkout Checkout
	if err := c.ShouldBindJSON(&checkout); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get the users shopping cart
	cart, err := getShoppingCart(checkout.SessionID)
	if err != nil {
		fmt.Println(err)
	}

	// Count up the price and charge the users creditcard
	// Get price at productservice
	var total int
	for _, v := range cart.Items {

		// Get the price of the SKU by calling the Products service
		log.Println("Calling service productservice...")
		url := fmt.Sprintf("%v/product/%v", productservice, v.Sku)
		resp, err := http.Get(url)
		if err != nil {
			log.Println(err)
		}
		defer resp.Body.Close()

		// Read HTTP body
		result, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
		}

		// Unmarshal JSON into struct
		type ProductResponse struct {
			Price int `json:"price"`
		}
		var pr ProductResponse
		json.Unmarshal(result, &pr)

		// Add the price of this SKU to the total amount
		total = total + (pr.Price * v.Qty)
	}

	// Charge the user by calling the Payments service
	transactionid, err := payProduct(checkout.Creditcard, total)
	if err != nil {
		log.Println(err)
	}

	// Ship the products to the user
	shippingid, err := shipProduct(checkout.Address, cart.Items)
	if err != nil {
		log.Println(err)
	}

	// Send the user an email
	err = sendEmail(checkout.Email)
	if err != nil {
		log.Println(err)
	}

	// Return checkout success and ID's
	log.Printf("Succesfully checked out user with sessionid %v for a total of €%v", checkout.SessionID, total/100)
	c.JSON(
		http.StatusOK,
		gin.H{
			"transactionid": transactionid,
			"shippingid":    shippingid,
		},
	)
}

// setupRouter initializes our HTTP routes
func setupRouter() *gin.Engine {
	router := gin.New()

	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

		// Custom log format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	router.Use(gin.Recovery())
	router.POST("/checkout", checkout)
	return router
}

func main() {
	// Start HTTP server
	gin.SetMode(gin.ReleaseMode)
	r := setupRouter()
	log.Println("Server started. Now accepting connections...")
	r.Run(":8080")
}
