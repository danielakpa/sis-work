package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"
)

// ── CONFIG — change these values ──────────────────────────────
const (
	myEmail       = "graceakpa083@gmail.com"
	myEmailPass   = "ojvs zkpd klba xynu"
	notifyEmail   = "graceakpa083@gmail.com"
	whatsappPhone = "2348102966386"
	ordersFile    = "orders.json"
)

// ── FLUTTERWAVE KEYS ──────────────────────────────────────────
// Option 1: Set as environment variables on Render:
//
//	FLW_PUBLIC_KEY = FLWPUBK_TEST-xxxxxxxxxxxx
//	FLW_SECRET_KEY = FLWSECK_TEST-xxxxxxxxxxxx
//
// Option 2: Paste directly below (not recommended for production)
func getFlwPublicKey() string {
	if k := os.Getenv("FLW_PUBLIC_KEY"); k != "" {
		return k
	}
	return "FLWPUBK_TEST-bd98a6d3b88e2282cc7e3321bccc1309-X" // ← PASTE PUBLIC KEY HERE
}
func getFlwSecretKey() string {
	if k := os.Getenv("FLW_SECRET_KEY"); k != "" {
		return k
	}
	return "FLWSECK_TEST-dcae97052935dd3a476200260f124d0b-X" // ← PASTE SECRET KEY HERE
}

// ─────────────────────────────────────────────────────────────

type OrderStatus string

const (
	StatusPending   OrderStatus = "Pending"
	StatusPaid      OrderStatus = "Paid"
	StatusPreparing OrderStatus = "Preparing"
	StatusDelivered OrderStatus = "Delivered"
)

type OrderItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type Location struct {
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

type Order struct {
	ID        string      `json:"id"`
	Items     []OrderItem `json:"items"`
	Total     float64     `json:"total"`
	Name      string      `json:"name"`
	Phone     string      `json:"phone"`
	Email     string      `json:"email"`
	Location  Location    `json:"location"`
	CreatedAt time.Time   `json:"created_at"`
	Status    OrderStatus `json:"status"`
	PaymentID string      `json:"payment_id"`
}

type OrderRequest struct {
	Name     string      `json:"name"`
	Phone    string      `json:"phone"`
	Email    string      `json:"email"`
	Items    []OrderItem `json:"items"`
	Location Location    `json:"location"`
}

var (
	orders   []Order
	ordersMu sync.Mutex
	counter  int
)

var menu = []map[string]interface{}{
	{
		"id":          "yummy-brown",
		"name":        "Yummy-Brown",
		"description": "A rich and nutritious delight with slightly sweet taste, leaving a satisfying, full-bodied feel.",
		"priceSmall":  2500.00,
		"priceBig":    3500.00,
		"emoji":       "🥣🟤",
		"badge":       "Fan Favourite",
	},
	{
		"id":          "creamy-custard",
		"name":        "Creamy-Custard",
		"description": "Silky smooth banana and vanilla flavour with a golden hue and melt-in-your-mouth texture. Cool, sweet, and impossibly creamy.",
		"priceSmall":  1500.00,
		"priceBig":    2000.00,
		"emoji":       "🍮",
		"badge":       "Fan Delight",
	},
}

func saveOrdersToFile() {
	ordersMu.Lock()
	ordersCopy := make([]Order, len(orders))
	copy(ordersCopy, orders)
	ordersMu.Unlock()
	data, err := json.MarshalIndent(ordersCopy, "", "  ")
	if err != nil {
		log.Println("Error saving orders:", err)
		return
	}
	if err := os.WriteFile(ordersFile, data, 0644); err != nil {
		log.Println("Error writing orders file:", err)
		return
	}
}

func loadOrdersFromFile() {
	data, err := os.ReadFile(ordersFile)
	if err != nil {
		return
	}
	if err := json.Unmarshal(data, &orders); err != nil {
		log.Println("Error loading orders file:", err)
		return
	}
	counter = len(orders)
	log.Printf("Loaded %d orders from %s\n", len(orders), ordersFile)
}

func sendEmailNotification(order Order) {
	go func() {
		auth := smtp.PlainAuth("", myEmail, myEmailPass, "smtp.gmail.com")
		var itemLines []string
		for _, item := range order.Items {
			itemLines = append(itemLines, fmt.Sprintf("  - %s x%d = ₦%.2f", item.Name, item.Quantity, item.Price*float64(item.Quantity)))
		}
		body := fmt.Sprintf(
			"Subject: 🍮 New Order %s from %s\r\n\r\n"+
				"New order received!\r\n\r\n"+
				"Order ID : %s\r\n"+
				"Customer : %s\r\n"+
				"Phone    : %s\r\n"+
				"Email    : %s\r\n"+
				"Address  : %s\r\n"+
				"Status   : %s\r\n"+
				"Time     : %s\r\n\r\n"+
				"Items:\r\n%s\r\n\r\n"+
				"Total: ₦%.2f\r\n",
			order.ID, order.Name,
			order.ID, order.Name, order.Phone, order.Email,
			order.Location.Address, string(order.Status),
			order.CreatedAt.Format("2006-01-02 15:04:05"),
			strings.Join(itemLines, "\r\n"),
			order.Total,
		)
		err := smtp.SendMail("smtp.gmail.com:587", auth, myEmail, []string{notifyEmail}, []byte(body))
		if err != nil {
			log.Println("Email error:", err)
		}
	}()
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func menuHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menu)
}

func orderHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodGet {
		ordersMu.Lock()
		defer ordersMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)
		return
	}

	if r.Method == http.MethodPost {
		var req OrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		if req.Location.Address == "" {
			http.Error(w, "Delivery address is required", http.StatusBadRequest)
			return
		}
		if len(req.Items) == 0 {
			http.Error(w, "At least one item required", http.StatusBadRequest)
			return
		}

		var total float64
		for _, item := range req.Items {
			total += item.Price * float64(item.Quantity)
		}

		ordersMu.Lock()
		counter++
		order := Order{
			ID:        fmt.Sprintf("ORD-%04d", counter),
			Items:     req.Items,
			Total:     total,
			Name:      req.Name,
			Phone:     req.Phone,
			Email:     req.Email,
			Location:  req.Location,
			CreatedAt: time.Now(),
			Status:    StatusPending,
		}
		orders = append(orders, order)
		ordersMu.Unlock()

		go saveOrdersToFile()
		sendEmailNotification(order)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(order)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// ── VERIFY FLUTTERWAVE PAYMENT ────────────────────────────────
func verifyPaymentHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TransactionID string `json:"transaction_id"`
		OrderID       string `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	flwURL := fmt.Sprintf("https://api.flutterwave.com/v3/transactions/%s/verify", req.TransactionID)
	httpReq, _ := http.NewRequest("GET", flwURL, nil)
	httpReq.Header.Set("Authorization", "Bearer "+getFlwSecretKey())
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Println("Flutterwave verify error:", err)
		http.Error(w, "Payment verification failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var flwResp map[string]interface{}
	json.Unmarshal(body, &flwResp)

	data, ok := flwResp["data"].(map[string]interface{})
	if !ok {
		http.Error(w, "Invalid Flutterwave response", http.StatusInternalServerError)
		return
	}

	status, _ := data["status"].(string)
	if status != "successful" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Payment not successful",
		})
		return
	}

	// Update order to Paid + Preparing
	ordersMu.Lock()
	for i, o := range orders {
		if o.ID == req.OrderID {
			orders[i].Status = StatusPreparing
			orders[i].PaymentID = req.TransactionID
			break
		}
	}
	ordersMu.Unlock()
	go saveOrdersToFile()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Payment verified! Your order is now being prepared.",
	})
}

// ── UPDATE ORDER STATUS (admin use) ──────────────────────────
func updateStatusHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID     string      `json:"id"`
		Status OrderStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	validStatuses := map[OrderStatus]bool{
		StatusPending: true, StatusPaid: true,
		StatusPreparing: true, StatusDelivered: true,
	}
	if !validStatuses[req.Status] {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	ordersMu.Lock()
	found := false
	for i, o := range orders {
		if o.ID == req.ID {
			orders[i].Status = req.Status
			found = true
			break
		}
	}
	ordersMu.Unlock()

	if !found {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	go saveOrdersToFile()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// ── SERVE FLUTTERWAVE PUBLIC KEY TO FRONTEND ──────────────────
func flwConfigHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"public_key": getFlwPublicKey(),
	})
}

func whatsappHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"phone": whatsappPhone})
}

func deleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ordersMu.Lock()http://localhost:8080
	newOrders := []Order{}
	for _, o := range orders {
		if o.ID != req.ID {
			newOrders = append(newOrders, o)
		}
	}
	orders = newOrders
	ordersMu.Unlock()

	go saveOrdersToFile()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func frontendHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	loadOrdersFromFile()

	http.HandleFunc("/", frontendHandler)
	http.HandleFunc("/api/menu", menuHandler)
	http.HandleFunc("/api/orders", orderHandler)
	http.HandleFunc("/api/orders/delete", deleteOrderHandler)
	http.HandleFunc("/api/orders/status", updateStatusHandler)
	http.HandleFunc("/api/payment/verify", verifyPaymentHandler)
	http.HandleFunc("/api/flw-config", flwConfigHandler)
	http.HandleFunc("/api/whatsapp", whatsappHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("🍮 Gracy Foodie Production server running at http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
