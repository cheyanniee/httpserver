# Internal Transfers HTTP Server

An internal transfers application that facilitates financial transactions between accounts using golang. 
This application provides HTTP endpoints for submitting transaction details and querying account balances.
A postgres database is used to maintain transaction logs and account states

---

## 📦 Installation

### 1. Clone the repository
```bash
git clone https://github.com/cheyanniee/httpserver.git
cd httpserver
```

### 2. Install dependencies
Make sure you have the following installed **before** continuing:

- [**Go**](https://go.dev/) **v1.20+**  
- [**PostgreSQL**](https://www.postgresql.org/)  

Once installed, download the project dependencies:

```bash
go mod tidy
```


### 3. Set up PostgreSQL
Create a database:
```sql
CREATE DATABASE transfers_db;
```


### 4. Configure environment variables
Configure the config variable in main.go with your own details
```go
var config = models.Config{
	DBUser:     "postgres",
	DBPassword: "12345",
	DBName:     "postgres",
	DBHost:     "localhost",
	DBPort:     5432,
	ServerPort: ":3333",
}
```
---

## ▶️ Running the Application

```bash
go run main.go
```

The server will start on `http://localhost:3333` unless changed in the config.

---

## 📡 API Endpoints

### **1. Create Account**
**POST** `/accounts`  
**Request Body:**
```json
{
  "account_id": 123,
  "initial_balance": "100.23344"
}
```
**Response:**  
Expected response is either an error or an empty response

---

### **2. Get Account Balance**
**GET** `/accounts/{account_id}`  
**Response:**
```json
{
  "account_id": 123,
  "balance": "100.23344"
}
```

---

### **3. Submit Transaction**
**POST** `/transactions`  
**Request Body:**
```json
{
  "source_account_id": 123,
  "destination_account_id": 456,
  "amount": "100.12345"
}
```
**Response:**  
Expected response is either an error or the updated balances

---

## 🛠 Assumptions

1. All accounts use the same currency.
2. No authentication/authorization is implemented.
3. AccountIDs are all numbers 
4. Balances are accurate up to 5 dp (can be altered if needed)

---

## 🧪 Running Tests
```bash
cd test
go test ./...
```
