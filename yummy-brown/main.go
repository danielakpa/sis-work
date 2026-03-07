package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type OrderItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type Order struct {
	ID        string      `json:"id"`
	Items     []OrderItem `json:"items"`
	Total     float64     `json:"total"`
	Name      string      `json:"name"`
	CreatedAt time.Time   `json:"created_at"`
}

type OrderRequest struct {
	Name  string      `json:"name"`
	Items []OrderItem `json:"items"`
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
		"description": "A nutritious yummy brown, coating your tongue with a rich and a hint of sweetness, leaving a satisfying full-bodied feel.",
		"price":       2500 & 3500,
		"badge":       "Fan Favourite",
	},
	{
		"id":          "creamy-custard",
		"name":        "Creamy-Custard",
		"description": "Silky smooth banana and vanilla custard with a golden hue and melt-in-your-mouth texture. sweet and impossibly creamy.",
		"price":       7.49,
		"emoji":       "🍮",
		"badge":       "sweet spot",
	},
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
			CreatedAt: time.Now(),
		}
		orders = append(orders, order)
		ordersMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(order)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func frontendHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	http.HandleFunc("/", frontendHandler)
	http.HandleFunc("/api/menu", menuHandler)
	http.HandleFunc("/api/orders", orderHandler)

	fmt.Println("Gracy Foodie Production server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
