# 🍮 Gracy Foodie Production

A full-stack food ordering web app built with **Go** (backend) and **HTML/CSS/JavaScript** (frontend). Customers can browse the menu, select sizes and flavours, place orders, and get notified via WhatsApp. You receive order notifications by email and can reply directly to customers.

---

## 📁 Project Structure

```
Gracy/
├── main.go        # Go backend server
├── index.html      # Frontend (served by Go)
├── go.mod         # Go module file
└── orders.json    # Auto-created, stores all orders
```

---

## 🚀 Getting Started

### Requirements
- [Go 1.22+](https://golang.org/dl/)
- A Gmail account with an **App Password** enabled
- 

### Run the server
```bash
cd yummy-brown
go run main.go
```

Then open your browser and go to:
```
http://127.0.0.1:8080
```

---

## ⚙️ Configuration

Open `main.go` and update these values at the top:

```go
const (
    myEmail       = "your@gmail.com"        // Your Gmail address
    myEmailPass   = "xxxx xxxx xxxx xxxx"   // Gmail App Password (16 chars)
    notifyEmail   = "your@gmail.com"        // Email to receive order notifications
    whatsappPhone = "2348102966386"         // Your WhatsApp number (no + or spaces)
    ordersFile    = "orders.json"           // File to save orders
)
```

### How to get a Gmail App Password
1. Go to [myaccount.google.com](https://myaccount.google.com)
2. Click **Security** → **2-Step Verification** (enable it if not already)
3. Scroll down to **App Passwords**
4. Create a new App Password for "Mail"
5. Copy the 16-character password into `myEmailPass`

---

## 🍽️ Menu Items

| Item | Small | Big |
|------|-------|-----|
| 🥣 Yummy-Brown | ₦2,500 | ₦3,500 |
| 🍮 Creamy-Custard | ₦1,500 | ₦2,000 |

Creamy-Custard supports **Vanilla** and **Banana** flavour selection.

---

## ✨ Features

- Browse menu with size picker (Small / Big)
- Flavour quantity selector for Creamy-Custard (Vanilla & Banana)
- Live order summary with ₦ pricing
- Customer details: Name, Phone, Email, Delivery Address
- GPS location detection + OpenStreetMap preview
- Order saved to `orders.json` (persists across restarts)
- Email notification sent to you on every new order
- WhatsApp message opened automatically after order
- Recent orders list with auto-refresh every 10 seconds
- ✔ Selected badge and gold border on chosen menu cards
- 🗑 Mark Delivered button to remove completed orders

---

## 🔌 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Serves the frontend |
| GET | `/api/menu` | Returns menu items |
| GET | `/api/orders` | Returns all orders |
| POST | `/api/orders` | Place a new order |
| POST | `/api/orders/delete` | Delete an order by ID |
| GET | `/api/whatsapp` | Returns WhatsApp phone number |


---

## 📞 Contact Numbers

Displayed at the bottom of the site:
- +234 8102966386
- +234 8083574564

---

## 🛠️ Built With

- **Go** — Backend server & REST API
- **HTML / CSS / JavaScript** — Frontend (single file)
- **Gmail SMTP** — Email notifications
- **WhatsApp API** — `wa.me` link for customer messaging
- **OpenStreetMap + Nominatim** — GPS reverse geocoding & map preview

---

 **Gracy Foodie Production** — Comfort Powered & Savor the Sweetness
