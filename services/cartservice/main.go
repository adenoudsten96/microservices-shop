package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

// Cart represents the shopping cart model
type Cart struct {
	Items []Item `json:"items" binding:"required"`
}

// Item represents the items in a Cart
type Item struct {
	Sku string `json:"sku" binding:"required"`
	Qty int    `json:"qty" binding:"required"`
}

// addToCart adds an item or items to a shopping cart in Redis.
// Shopping carts are identified by session IDs and contain Items. Each item contains the product SKU and quantity.
// So, our Redis carts look like this:
// sessionid: [
// 	"sku1": 2,
// 	"sku2": 3 ]
func addToCart(c *gin.Context) {

	// Get the session ID
	sessionid := c.Param("sessionid")

	// Unmarshal the JSON data from the body
	var cart Cart
	if err := c.ShouldBindJSON(&cart); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Add all the Items in the Cart to Redis
	for _, i := range cart.Items {
		err := rclient.HSet(sessionid, i.Sku, i.Qty).Err()
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	// Return status created
	c.JSON(
		http.StatusCreated,
		gin.H{
			"status": "added items to cart",
		},
	)
}

// getCart gets all items from a shopping cart and returns them as JSON.
func getCart(c *gin.Context) {

	// Get the session ID
	sessionid := c.Param("sessionid")

	// Get all items in the shopping cart by session ID
	result, err := rclient.HGetAll(sessionid).Result()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Convert to our Cart struct and marshal to JSON
	var items []Item
	for k, v := range result {
		qty, _ := strconv.Atoi(v)
		items = append(items, Item{k, qty})
	}
	cart := Cart{
		Items: items,
	}

	// Return the data
	c.JSON(
		http.StatusOK,
		cart,
	)

}

// emptyCart empties a shopping cart by deleting the key in Redis.
func emptyCart(c *gin.Context) {

	// Get the session ID
	sessionid := c.Param("sessionid")

	// Delete the cart in Redis
	if err := rclient.Del(sessionid).Err(); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return cart emptied message
	c.JSON(
		http.StatusOK,
		gin.H{"message": "cart emptied"},
	)
}

// setupRouter initializes our HTTP routes
func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/cart/:sessionid", getCart)
	router.POST("/cart/:sessionid", addToCart)
	router.DELETE("/cart/:sessionid", emptyCart)
	return router
}

var rclient *redis.Client

// init initializes our Redis database
func init() {
	redisHost := os.Getenv("REDIS_HOST")
	rclient = redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := rclient.Ping().Err()
	if err != nil {
		log.Panicln("Could not connect to Redis on host", redisHost)
	}
}

func main() {
	// Start HTTP server
	r := setupRouter()
	log.Println("Server started. Now accepting connections...")
	r.Run(":8080")
}
