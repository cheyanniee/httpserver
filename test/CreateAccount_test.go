package test

import (
	"bytes"
	"database/sql"
	"httpserver/handlers"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

/* Testcases for GetAccountByID */

// Success: Valid request
func TestCreateAccountHandler_Success(t *testing.T) {
	mock := setupMockDB(t)

	// Expect successful creation of account
	mock.ExpectExec("INSERT INTO accounts").
		WithArgs(1, 100.0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Valid account details in body
	body := []byte(`{"account_id": 1, "initial_balance": "100.00"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handlers.CreateAccountHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Expect empty response
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

// Fail: Invalid JSON
func TestCreateAccountHandler_InvalidJSON(t *testing.T) {
	// Invalid JSON in body
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBufferString("{invalid json"))
	w := httptest.NewRecorder()

	handlers.CreateAccountHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Expect bad request
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Fail: Invalid AccountID
func TestCreateAccountHandler_InvalidAccountID(t *testing.T) {
	// String for initial_balance instead of a number
	body := []byte(`{"account_id": "not-a-number", "initial_balance": "100.00"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handlers.CreateAccountHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Expect bad request
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Fail: Invalid Balance
func TestCreateAccountHandler_InvalidBalance(t *testing.T) {
	// String for initial_balance instead of a number
	body := []byte(`{"account_id": 1, "initial_balance": "not-a-number"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handlers.CreateAccountHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Expect bad request
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Fail: DB Error
func TestCreateAccountHandler_DBError(t *testing.T) {
	mock := setupMockDB(t)

	// Simulate a DB error
	mock.ExpectExec("INSERT INTO accounts").
		WithArgs(1, 100.0).
		WillReturnError(sql.ErrConnDone)

	// Valid account details in body
	body := []byte(`{"account_id": 1, "initial_balance": "100.00"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handlers.CreateAccountHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Expect server error
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}

// Fail: Invalid Method
func TestCreateAccountHandler_InvalidMethod(t *testing.T) {
	// Use GET instead of POST
	req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	w := httptest.NewRecorder()

	handlers.CreateAccountHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Expect invalid method error
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}
