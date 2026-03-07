package main

import (
	"encoding/json"
	"fmt"
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
	myEmail       = "graceakpa083@gmail.com" // your Gmail address
	myEmailPass   = "ojvs zkpd klba xynu"    // Gmail App Password (not your login password)
	notifyEmail   = "graceakpa083@gmail.com" // email to receive order notifications
	whatsappPhone = "2348102966386"          // your WhatsApp number (no + or spaces)
	ordersFile    = "orders.json"            // file to save orders
)

// ─────────────────────────────────────────────────────────────

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
	Location  Location    `json:"location"`
	CreatedAt time.Time   `json:"created_at"`
}

type OrderRequest struct {
	Name     string      `json:"name"`
	Phone    string      `json:"phone"`
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

// ── SAVE ORDERS TO FILE ───────────────────────────────────────
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
	log.Println("Orders saved to", ordersFile)
}

func loadOrdersFromFile() {
	data, err := os.ReadFile(ordersFile)
	if err != nil {
		return // file doesn't exist yet, that's fine
	}
	if err := json.Unmarshal(data, &orders); err != nil {
		log.Println("Error loading orders file:", err)
		return
	}
	counter = len(orders)
	log.Printf("Loaded %d orders from %s\n", len(orders), ordersFile)
}

// ── SEND EMAIL ────────────────────────────────────────────────
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
				"Phone   : %s\r\n"+
				"Address  : %s\r\n"+
				"Time     : %s\r\n\r\n"+
				"Items:\r\n%s\r\n\r\n"+
				"Total: ₦%.2f\r\n",
			order.ID, order.Name,
			order.ID,
			order.Name,
			order.Phone,
			order.Location.Address,
			order.CreatedAt.Format("2006-01-02 15:04:05"),
			strings.Join(itemLines, "\r\n"),
			order.Total,
		)

		err := smtp.SendMail(
			"smtp.gmail.com:587",
			auth,
			myEmail,
			[]string{notifyEmail},
			[]byte(body),
		)
		if err != nil {
			log.Println("Email error:", err)
		} else {
			log.Println("Email notification sent for", order.ID)
		}
	}()
}

// ── HANDLERS ──────────────────────────────────────────────────
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
			Location:  req.Location,
			CreatedAt: time.Now(),
		}
		orders = append(orders, order)
		ordersMu.Unlock()

		// 1. Save to file
		go saveOrdersToFile()

		// 2. Send email
		sendEmailNotification(order)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(order)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// whatsappHandler returns the WhatsApp phone number to the frontend
func whatsappHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"phone": whatsappPhone})
}

func frontendHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	// Load saved orders on startup
	loadOrdersFromFile()

	http.HandleFunc("/", frontendHandler)
	http.HandleFunc("/api/menu", menuHandler)
	http.HandleFunc("/api/orders", orderHandler)
	http.HandleFunc("/api/whatsapp", whatsappHandler)

	fmt.Println("🍮 Gracy Foodie Production server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
