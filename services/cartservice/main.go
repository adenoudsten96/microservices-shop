package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
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
		log.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Add all the Items in the Cart to Redis
	for _, i := range cart.Items {
		err := rclient.HSet(sessionid, i.Sku, i.Qty).Err()
		if err != nil {
			log.WithFields(log.Fields{
				"item":      i,
				"sessionid": sessionid,
			}).Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	// Return status created
	log.WithFields(log.Fields{
		"cart":      cart,
		"sessionid": sessionid,
	}).Info("Added item(s) to cart")
	c.JSON(
		http.StatusCreated,
		gin.H{
			"status": "ok",
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
		log.WithFields(log.Fields{
			"sessionid": sessionid,
		}).Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Convert to our Cart struct
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
		log.WithFields(log.Fields{
			"sessionid": sessionid,
		}).Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return cart emptied message
	log.WithFields(log.Fields{
		"sessionid": sessionid,
	}).Info("Deleted item(s) from cart")
	c.JSON(
		http.StatusOK,
		gin.H{"status": "ok"},
	)
}

// setupRouter initializes our HTTP routes
func setupRouter() *gin.Engine {
	router := gin.New()

	// router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

	// 	// Custom log format
	// 	return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
	// 		param.ClientIP,
	// 		param.TimeStamp.Format(time.RFC1123),
	// 		param.Method,
	// 		param.Path,
	// 		param.Request.Proto,
	// 		param.StatusCode,
	// 		param.Latency,
	// 		param.Request.UserAgent(),
	// 		param.ErrorMessage,
	// 	)
	// }))
	logger := logrus.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	router.Use(ginlogrus.Logger(logger), gin.Recovery())

	router.GET("/cart/:sessionid", getCart)
	router.POST("/cart/:sessionid", addToCart)
	router.DELETE("/cart/:sessionid", emptyCart)
	router.GET("/health", healthCheck)
	return router
}

func healthCheck(c *gin.Context) {
	c.String(200, "OK")
}

var rclient *redis.Client

// init initializes our Redis database
func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	redisHost := mustMapEnv("REDIS_HOST")
	rclient = redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	log.Println("Starting service userservice...")
	log.Printf("Connecting to Redis on host '%v'...", redisHost)
	err := rclient.Ping().Err()
	if err != nil {
		// Retry a couple times
		counter := 3
		for counter > 0 {
			err := rclient.Ping().Err()
			if err != nil {
				log.Printf("Could not connect to Redis on host '%v', trying %v more time(s)", redisHost, counter)
				counter--
				time.Sleep(time.Second * 1)
				if counter == 0 {
					log.Panicf("Could not connect to Redis on host '%v'.", redisHost)
					break
				}
				continue
			}
			break
		}
	}
	log.Printf("Successfully connected to Redis on host '%v'...", redisHost)
}

func main() {
	// Start HTTP server
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	r := setupRouter()
	log.Info("Service cartservice started. Now accepting connections...")
	r.Run(":8081")
}

func mustMapEnv(envKey string) string {
	if os.Getenv(envKey) == "" {
		log.WithFields(log.Fields{
			"envkey": envKey,
		}).Panic("Could not bind environment variable")
	}
	return os.Getenv(envKey)
}
