package test

import (
	"database/sql"
	"encoding/json"
	"httpserver/handlers"
	"httpserver/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

/* Testcases for GetAccountByID */

// Success: Valid request
func TestGetAccountByID_Success(t *testing.T) {
	mock := setupMockDB(t)

	// Expect successfully getting account
	mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id =").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(150.75))

	account, err := handlers.GetAccountByID(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if account.AccountID != 1 || account.CurrentBalance != 150.75 {
		t.Errorf("unexpected account: %+v", account)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Fail: DB Error
func TestGetAccountByID_DBError(t *testing.T) {
	mock := setupMockDB(t)

	// Simulate a DB error
	mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id =").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	_, err := handlers.GetAccountByID(1)
	if err == nil {
		t.Errorf("expected DB error, got nil")
	}
}

/* Testcases for CreateAccountHandler */

// Success: Valid request
func TestGetAccountHandler_Success(t *testing.T) {
	mock := setupMockDB(t)

	// Expect successful creation of account
	mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id =").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(200.50))

	// Valid account
	req := httptest.NewRequest(http.MethodGet, "/accounts/1", nil)
	w := httptest.NewRecorder()

	handlers.GetAccountHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Checking for status 200
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var acc models.Account
	if err := json.NewDecoder(resp.Body).Decode(&acc); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	// Checking for correct account details
	if acc.AccountID != 1 || acc.CurrentBalance != 200.50 {
		t.Errorf("unexpected account data: %+v", acc)
	}
}

// Fail: Invalid AccountID
func TestGetAccountHandler_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/accounts/abc", nil)
	w := httptest.NewRecorder()

	handlers.GetAccountHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Expect bad request
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Fail: Account not found
func TestGetAccountHandler_NotFound(t *testing.T) {
	mock := setupMockDB(t)

	// Simulate account not found
	mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id =").
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/accounts/999", nil)
	w := httptest.NewRecorder()

	handlers.GetAccountHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Expect not found error
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// Fail: DB Error
func TestGetAccountHandler_DBError(t *testing.T) {
	mock := setupMockDB(t)

	// Simulate a DB error
	mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id =").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest(http.MethodGet, "/accounts/1", nil)
	w := httptest.NewRecorder()

	handlers.GetAccountHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Expect server error
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}
