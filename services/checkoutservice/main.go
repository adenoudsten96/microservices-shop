package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Checkout represents the information required to perform a succesfull checkout.
type Checkout struct {
	SessionID  string `json:"sessionid" binding:"required"`
	Address    string `json:"address" binding:"required"`
	Email      string `json:"email" binding:"required"`
	Creditcard string `json:"creditcard" binding:"required"`
}

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

	getShoppingCart()
	payProducts()
	shipProducts()
	sendEmail()

	return
}

// setupRouter initializes our HTTP routes
func setupRouter() *gin.Engine {
	router := gin.Default()
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
