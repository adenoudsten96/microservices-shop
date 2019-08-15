package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Product represents the product model
type Product struct {
	gorm.Model
	SKU         string `json:"sku" binding:"required" gorm:"unique"`
	Name        string `json:"name" binding:"required"`
	Price       int    `json:"price" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// getAllProducts fetches all products from the database and returns them as JSON.
func getAllProducts(c *gin.Context) {
	var allProducts []Product
	if err := db.Find(&allProducts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, allProducts)
}

// getProduct fetches a specific product from the database and returns it as JSON.
func getProduct(c *gin.Context) {

	// Get the SKU ID
	sku := c.Param("sku")

	// Check if there is a product with this SKU in the database
	var product Product
	if result := db.Where("sku = ?", sku).First(&product).RowsAffected; result == 0 {
		log.Println(sku)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not found",
		})
		return
	}

	// Return the product
	c.JSON(200, product)
}

// createProduct adds a new product to the database.
func createProduct(c *gin.Context) {

	// Get the JSON data
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if the product already exists by checking if there are rows affected
	if result := db.Where("sku = ?", product.SKU).First(&product).RowsAffected; result == 1 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "a product with this SKU already exists",
		})
		return
	}

	// Insert the product into the database
	if err := db.Create(&product).Error; err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return the created product
	c.JSON(
		http.StatusCreated,
		product,
	)
}

var db *gorm.DB

// init initializes our Postgres database
func init() {
	// Setup the database connection
	var err error
	username := "postgres"
	password := os.Getenv("DB_PASS")
	dbName := "products"
	dbHost := os.Getenv("DB_HOST")
	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, username, dbName, password)

	log.Println("Starting service productservice...")
	log.Printf("Connecting to database on host '%v'...", dbHost)
	db, err = gorm.Open("postgres", dbURI)
	if err != nil {
		// Retry a couple times
		counter := 3
		for counter > 0 {
			db, err = gorm.Open("postgres", dbURI)
			if err != nil {
				log.Println(err)
				log.Printf("Could not connect to database on host '%v', trying %v more time(s)", dbHost, counter)
				counter--
				time.Sleep(2 * time.Second)

				if counter == 0 {
					log.Panicf("Could not connect to database on host '%v'.", dbHost)
					break
				}
				continue
			}
			break
		}
	}
	log.Printf("Successfully connected to database on host '%v'...", dbHost)

	// Migrate the schema
	db.AutoMigrate(&Product{})

}

// setupRouter initializes our HTTP routes
func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/product", getAllProducts)
	router.GET("/product/:sku", getProduct)
	router.POST("/product", createProduct)
	return router
}

func main() {
	// Start HTTP server
	gin.SetMode(gin.ReleaseMode)
	r := setupRouter()
	log.Println("Server started. Now accepting connections...")
	r.Run(":8082")
}
