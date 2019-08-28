package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

var (
	tpl             *template.Template
	cartservice     = mustMapEnv("CARTSERVICE")
	productservice  = mustMapEnv("PRODUCTSERVICE")
	checkoutservice = mustMapEnv("CHECKOUTSERVICE")
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
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func homePage(w http.ResponseWriter, r *http.Request) {

	// Set the sessionID in a cookie if there isn't already one set
	_, err := r.Cookie("sessionid")
	if err != nil {
		uuid, err := uuid.NewV4()
		if err != nil {
			log.Error(err)
		}
		cookie := http.Cookie{Name: "sessionid", Value: uuid.String(), Expires: time.Now().Add(1 * time.Hour)}
		http.SetCookie(w, &cookie)
	}

	// Get all products
	products, status, err := getProducts()
	// Render error page if something went wrong
	if status != 200 {
		log.Error(err)
		renderError(w, r, status, err)
		return
	}

	err = tpl.ExecuteTemplate(w, "home.html", products)
	if err != nil {
		log.Error(err)
	}
}

func productPage(w http.ResponseWriter, r *http.Request) {

	products, status, err := getProducts()

	// Render error page if something went wrong
	if status != 200 {
		log.Error(err)
		renderError(w, r, status, err)
		return
	}

	vars := mux.Vars(r)
	sku := vars["SKU"]

	for _, v := range products {
		if v.SKU == sku {
			err := tpl.ExecuteTemplate(w, "product.html", v)
			if err != nil {
				log.Error(err)
				return
			}
			return
		}
	}

}

func cartPage(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {

		// Get the form values and session ID
		r.ParseForm()
		sku := r.PostFormValue("sku")
		qtystr := r.PostFormValue("qty")
		cookie, err := r.Cookie("sessionid")
		if err != nil {
			log.Error(err)
		}
		sessionid := cookie.Value
		qty, _ := strconv.Atoi(qtystr)

		// Add the items to the cart
		status, err := addToCart(sessionid, sku, qty)

		// Render the error page if something went wrong
		if status != 201 {
			log.Error(err)
			renderError(w, r, status, err)
			return
		}
	}

	// Get the users shopping cart
	ck, err := r.Cookie("sessionid")
	if err != nil {
		log.Error(err)
		renderError(w, r, 0, err)
	}
	sessionid := ck.Value
	cart, status, err := getCart(sessionid)

	if err != nil {
		log.Error(err)
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

	products, status, err := getProducts()
	// Render error page if something went wrong
	if status != 200 {
		log.Error(err)
		renderError(w, r, status, err)
		return
	}

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
		log.Error(err)
	}
}

func checkoutPage(w http.ResponseWriter, r *http.Request) {
	// Get form values and sessionID
	r.ParseForm()
	address := r.PostFormValue("street_address")
	creditcard := r.PostFormValue("credit_card_number")
	email := r.PostFormValue("email")
	total := r.PostFormValue("total")
	cookie, _ := r.Cookie("sessionid")
	sessionid := cookie.Value

	// Prepare JSON payload
	payload := map[string]string{
		"SessionID":  sessionid,
		"Address":    address,
		"Email":      email,
		"Creditcard": creditcard,
	}
	jsonPayload, _ := json.Marshal(payload)

	// Check the user out by calling the checkoutservice
	url := fmt.Sprintf("%v/checkout", checkoutservice)
	log.Info("Calling service checkoutservice...")
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()

	// Read HTTP body
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
	}

	// Unmarshal response
	type CheckoutResponse struct {
		TransactionID string
		ShippingID    string
	}
	var cr CheckoutResponse
	json.Unmarshal(result, &cr)

	// Render error page if something went wrong
	if resp.StatusCode != 200 {
		log.Error(err)
		renderError(w, r, resp.StatusCode, errors.New(string(result)))
		return
	}

	// Empty the shopping cart and render the page if the checkout was succesful
	status, err := deleteCart(sessionid)
	if err != nil {
		log.Error(err)
		renderError(w, r, status, err)
		return
	}

	err = tpl.ExecuteTemplate(w, "checkout.html", map[string]interface{}{
		"response": cr,
		"total":    total,
	})
	if err != nil {
		log.Error(err)
	}
}

func emptyCart(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sessionid")
	if err != nil {
		log.Error(err)
		renderError(w, r, 0, err)
		return
	}
	sessionid := cookie.Value

	status, err := deleteCart(sessionid)
	if err != nil {
		log.Error(err)
		renderError(w, r, status, err)
		return
	}

	http.Redirect(w, r, "/", 301)
}

func getProducts() ([]ProductResponse, int, error) {

	// Get all products
	url := fmt.Sprintf("%v/product", productservice)

	log.Info("Calling service productservice...")
	resp, err := http.Get(url)
	if err != nil {
		log.Error(err)
		return []ProductResponse{}, 0, err
	}
	defer resp.Body.Close()

	// Read HTTP body
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return []ProductResponse{}, 0, err
	}

	if resp.StatusCode != 200 {
		return []ProductResponse{}, resp.StatusCode, errors.New(string(result))
	}

	var products []ProductResponse
	err = json.Unmarshal(result, &products)

	if err != nil {
		log.Error(err)
		return []ProductResponse{}, 0, err
	}

	return products, 200, nil
}

func addToCart(sessionid, sku string, qty int) (int, error) {
	// Add the items to our cart by calling the cartservice
	url := fmt.Sprintf("%v/cart/%v", cartservice, sessionid)
	log.Info("Calling service cartservice...")
	cart := Cart{
		Items: []Item{Item{Sku: sku, Qty: qty}},
	}
	jsonValue, err := json.Marshal(cart)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Error(err)
		return 0, err
	}
	defer resp.Body.Close()

	// Read HTTP body
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	if resp.StatusCode != 201 {
		return resp.StatusCode, errors.New(string(result))
	}

	return 201, nil
}

func getCart(sessionid string) (Cart, int, error) {
	url := fmt.Sprintf("%v/cart/%v", cartservice, sessionid)

	log.Info("Calling service cartservice...")
	resp, err := http.Get(url)
	if err != nil {
		log.Error(err)
		return Cart{}, 0, err
	}
	defer resp.Body.Close()

	// Read HTTP body
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return Cart{}, 0, err
	}

	if resp.StatusCode != 200 {
		log.Error(err)
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
		log.Error(err)
		return 0, err
	}

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return 0, err
	}
	defer resp.Body.Close()

	// Read Response Body
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
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

func mustMapEnv(envKey string) string {
	if os.Getenv(envKey) == "" {
		log.Panicf("Environment variable %v not set", envKey)
	}
	return os.Getenv(envKey)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/", homePage).Methods(http.MethodGet)
	r.HandleFunc("/product/{SKU}", productPage).Methods(http.MethodGet)
	r.HandleFunc("/cart", cartPage).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/cart/empty", emptyCart).Methods(http.MethodGet)
	r.HandleFunc("/checkout", checkoutPage).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/health", checkoutPage).Methods(http.MethodGet)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Info("Starting service frontend")

	// Run server
	http.ListenAndServe(":80", r)
}
