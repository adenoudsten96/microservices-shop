package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

var (
	tpl             *template.Template
	cartservice     = os.Getenv("CARTSERVICE")
	emailservice    = os.Getenv("EMAILSERVICE")
	paymentservice  = os.Getenv("PAYMENTSERVICE")
	shippingservice = os.Getenv("SHIPPINGSERVICE")
	productservice  = os.Getenv("PRODUCTSERVICE")
)

// ProductResponse is the response that comes back from the productservice
type ProductResponse struct {
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
}

// Cart represents the shopping cart model
type Cart struct {
	Items []Item `json:"items" binding:"required"`
}

// Item represents the items in a Cart
type Item struct {
	Sku string `json:"sku" binding:"required"`
	Qty int    `json:"qty" binding:"required"`
}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func cartDelete(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sessionid")
	if err != nil {
		log.Println(err)
		renderError(w, r, 0, err)
		return
	}
	sessionid := cookie.Value

	status, err := deleteCart(sessionid)
	if err != nil {
		log.Println(err)
		renderError(w, r, status, err)
		return
	}

	http.Redirect(w, r, "/", 301)
}

func cart(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {

		// Get the form values and session ID
		r.ParseForm()
		sku := r.PostFormValue("sku")
		qtystr := r.PostFormValue("qty")
		cookie, err := r.Cookie("sessionid")
		if err != nil {
			log.Println(err)
		}
		sessionid := cookie.Value
		qty, _ := strconv.Atoi(qtystr)

		// Add the items to the cart
		status, err := addToCart(sessionid, sku, qty)

		// Render the error page if something went wrong
		if status != 201 {
			renderError(w, r, status, err)
			return
		}
	}

	// Get the users shopping cart
	ck, err := r.Cookie("sessionid")
	if err != nil {
		log.Println(err)
		renderError(w, r, 0, err)
	}
	sessionid := ck.Value
	cart, status, err := getCart(sessionid)

	if err != nil {
		renderError(w, r, status, err)
		return
	}

	// Match the items in the shopping cart to products
	// Prepare a struct to render the page with
	// Count up total money owed
	type ItemRow struct {
		Sku      string
		Name     string
		Price    int
		Quantity int
	}

	var irs []ItemRow
	var total int
	products := getProducts()
	for _, v := range cart.Items {
		var ir ItemRow
		for _, c := range products {
			if v.Sku == c.SKU {
				ir.Sku = c.SKU
				ir.Name = c.Name
				ir.Price = c.Price
			}
		}
		ir.Quantity = v.Qty
		irs = append(irs, ir)
		total = total + (v.Qty * ir.Price)
	}

	// Render template
	err = tpl.ExecuteTemplate(w, "cart.html", map[string]interface{}{
		"items": irs,
		"total": total})
	if err != nil {
		log.Println(err)
	}
}

func product(w http.ResponseWriter, r *http.Request) {

	products := getProducts()

	vars := mux.Vars(r)
	sku := vars["SKU"]

	for _, v := range products {
		if v.SKU == sku {
			err := tpl.ExecuteTemplate(w, "product.html", v)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func homepage(w http.ResponseWriter, r *http.Request) {

	// Set the sessionID in a cookie if there isn't already one set
	_, err := r.Cookie("sessionid")
	if err != nil {
		uuid, err := uuid.NewV4()
		if err != nil {
			log.Println(err)
		}
		cookie := http.Cookie{Name: "sessionid", Value: uuid.String(), Expires: time.Now().Add(1 * time.Hour)}
		http.SetCookie(w, &cookie)
	}

	// Get all products and display them
	products := getProducts()
	err = tpl.ExecuteTemplate(w, "home.html", products)
	if err != nil {
		log.Println(err)
	}
}

func getProducts() []ProductResponse {

	// Get all products
	url := fmt.Sprintf("%v/product", productservice)

	log.Println("Calling service productservice...")
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

	var products []ProductResponse
	json.Unmarshal(result, &products)

	return products
}

func addToCart(sessionid, sku string, qty int) (int, error) {
	// Add the items to our cart by calling the cartservice
	url := fmt.Sprintf("%v/cart/%v", cartservice, sessionid)
	log.Println("Calling service cartservice...")
	cart := Cart{
		Items: []Item{Item{Sku: sku, Qty: qty}},
	}
	jsonValue, err := json.Marshal(cart)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	// Read HTTP body
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	if resp.StatusCode != 201 {
		return resp.StatusCode, errors.New(string(result))
	}

	return 201, nil
}

func getCart(sessionid string) (Cart, int, error) {
	url := fmt.Sprintf("%v/cart/%v", cartservice, sessionid)

	log.Println("Calling service cartservice...")
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

	if resp.StatusCode != 200 {
		return Cart{}, resp.StatusCode, errors.New(string(result))
	}

	var cart Cart
	json.Unmarshal(result, &cart)

	return cart, 200, nil
}

func deleteCart(sessionid string) (int, error) {
	url := fmt.Sprintf("%v/cart/%v", cartservice, sessionid)

	// Create request
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	defer resp.Body.Close()

	// Read Response Body
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	if resp.StatusCode != 200 {
		return resp.StatusCode, errors.New(string(result))
	}

	return 200, nil
}

func renderError(w http.ResponseWriter, r *http.Request, code int, err error) {
	errMsg := fmt.Sprintf("%+v", err)
	w.WriteHeader(code)
	_ = tpl.ExecuteTemplate(w, "error", map[string]interface{}{
		"error":       errMsg,
		"status_code": code})
}

func main() {
	fmt.Println("Hello world!")

	r := mux.NewRouter()
	r.HandleFunc("/", homepage)
	r.HandleFunc("/product/{SKU}", product)
	r.HandleFunc("/cart", cart)
	r.HandleFunc("/cart/empty", cartDelete)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Run server
	http.ListenAndServe(":80", r)
}
