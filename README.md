# Go API Service

A lightweight API service built in **Go** using the [chi router](https://github.com/go-chi/chi), with authentication middleware, structured logging, and a mock database.
This project demonstrates clean structuring of a Go backend, focusing on **routing, middleware, error handling, and JSON APIs**.

---

## 🚀 Features

* **Routing** with `chi.Router` (`/account/coins` endpoint).
* **Middleware** for:

  * Authentication (username + token).
  * Normalizing routes (strip trailing slashes).
* **Mock Database** simulating login details and coin balances.
* **Error Handling** with consistent JSON responses.
* **Structured Logging** using [logrus](https://github.com/sirupsen/logrus).

---

## 📂 Project Structure

```
.
├── api/                  # API response types & error helpers
│   └── api.go
├── cmd/
│   └── api/              # Entry point
│       └── main.go
├── internal/
│   ├── handlers/         # Route handlers
│   │   ├── api.go        # Router setup
│   │   └── get_coin_balance.go
│   ├── middleware/       # Custom middleware
│   │   └── authorization.go
│   └── tools/            # Database abstraction + mock DB
│       ├── database.go
│       └── mockdb.go
└── go.mod                # Go module definition
```

---

## ⚙️ Setup & Run

### 1. Clone the repository

```bash
git clone https://github.com/<your-username>/<your-repo>.git
cd <your-repo>
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Run the API

```bash
go run cmd/api/main.go
```

### 4. API will be available at:

```
http://localhost:3000
```

---

## 📡 API Usage

### 🔐 Authentication

Every request to `/account/...` requires:

* `username` as a query parameter.
* `Authorization` token in the request header.

Valid usernames/tokens (from mock DB):

```json
{
  "alex":   "123ABC",
  "jason":  "456DEF",
  "marie":  "789GHI"
}
```

---

### 🪙 Get Coin Balance

**Request:**

```http
GET /account/coins?username=alex
Authorization: 123ABC
```

**Response:**

```json
{
  "Code": 200,
  "Balance": 100
}
```

**Unauthorized example:**

```json
{
  "Code": 400,
  "Message": "Invalid username or token"
}
```

---

## 🛠️ Tech Stack

* **Go** (1.21+)
* [chi](https://github.com/go-chi/chi) — HTTP router.
* [gorilla/schema](https://github.com/gorilla/schema) — Query param decoding.
* [logrus](https://github.com/sirupsen/logrus) — Structured logging.

---

## 📖 What this project demonstrates

* Structuring a Go backend with **cmd/**, **internal/**, and **api/** packages.
* Writing **middleware** to enforce authentication.
* Clean error handling with **consistent JSON responses**.
* Designing for extensibility with **interfaces** and a mock DB.
* Using **pointers** in Go to handle `nil` checks and avoid unnecessary copies.
