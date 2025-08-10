package test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"httpserver/handlers"
	"httpserver/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// Success: Valid request
func TestTransferCurrency_Success(t *testing.T) {
	mock := setupMockDB(t)

	source := models.Account{AccountID: 1, CurrentBalance: 100.0}
	dest := models.Account{AccountID: 2, CurrentBalance: 50.0}
	amount := 20.0

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE accounts SET balance =").
		WithArgs(80.0, 1).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

	mock.ExpectExec("UPDATE accounts SET balance =").
		WithArgs(70.0, 2).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

	mock.ExpectCommit()

	err := handlers.TransferCurrency(source, dest, amount)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Fail: Source account not found
func TestTransferCurrency_SourceAccountNotFound(t *testing.T) {
	mock := setupMockDB(t)

	source := models.Account{AccountID: 1, CurrentBalance: 100.0}
	dest := models.Account{AccountID: 2, CurrentBalance: 50.0}
	amount := 20.0

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE accounts SET balance =").
		WithArgs(80.0, 1).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
	mock.ExpectRollback()

	err := handlers.TransferCurrency(source, dest, amount)
	if err == nil || err.Error() != "source account not found" {
		t.Errorf("expected source account not found error, got %v", err)
	}
}

// Fail: Destination account not found
func TestTransferCurrency_DestinationAccountNotFound(t *testing.T) {
	mock := setupMockDB(t)

	source := models.Account{AccountID: 1, CurrentBalance: 100.0}
	dest := models.Account{AccountID: 2, CurrentBalance: 50.0}
	amount := 20.0

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE accounts SET balance =").
		WithArgs(80.0, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("UPDATE accounts SET balance =").
		WithArgs(70.0, 2).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
	mock.ExpectRollback()

	err := handlers.TransferCurrency(source, dest, amount)
	if err == nil || err.Error() != "destination account not found" {
		t.Errorf("expected destination account not found error, got %v", err)
	}
}

// Fail: DB Error (Beginning)
func TestTransferCurrency_DBErrorOnBegin(t *testing.T) {
	mock := setupMockDB(t)

	mock.ExpectBegin().WillReturnError(errors.New("db begin error"))

	source := models.Account{AccountID: 1, CurrentBalance: 100.0}
	dest := models.Account{AccountID: 2, CurrentBalance: 50.0}

	err := handlers.TransferCurrency(source, dest, 20.0)
	if err == nil || err.Error() != "db begin error" {
		t.Errorf("expected db begin error, got %v", err)
	}
}

// Fail: DB Error (Commit)
func TestTransferCurrency_CommitError(t *testing.T) {
	mock := setupMockDB(t)

	source := models.Account{AccountID: 1, CurrentBalance: 100.0}
	dest := models.Account{AccountID: 2, CurrentBalance: 50.0}
	amount := 20.0

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE accounts SET balance =").
		WithArgs(80.0, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("UPDATE accounts SET balance =").
		WithArgs(70.0, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	err := handlers.TransferCurrency(source, dest, amount)
	if err == nil || err.Error() != "commit error" {
		t.Errorf("expected commit error, got %v", err)
	}
}

func TestTransactionHandler_Success(t *testing.T) {
	mock := setupMockDB(t)

	// Expect existance of source account
	mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id =").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.0))

	// Expect existance of destination account
	mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id =").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(50.0))

	mock.ExpectBegin()

	// Expect updating of source account
	mock.ExpectExec("UPDATE accounts SET balance =").
		WithArgs(80.0, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect updating of destination account
	mock.ExpectExec("UPDATE accounts SET balance =").
		WithArgs(70.0, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// Expect updated source account
	mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id =").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(80.0))

	// Expect updated destination account
	mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id =").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(70.0))

	// Valid request
	body := []byte(`{"source_account_id": 1, "destination_account_id": 2, "amount": "20.00"}`)
	req := httptest.NewRequest(http.MethodPost, "/transfer", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handlers.TransactionHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	// Decode JSON response
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if data["source_balance"] != 80.0 {
		t.Errorf("expected source_balance=80.0, got %v", data["source_balance"])
	}
	if data["dest_balance"] != 70.0 {
		t.Errorf("expected dest_balance=70.0, got %v", data["dest_balance"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Fail: Destination account not found
func TestTransactionHandler_SourceNotFound(t *testing.T) {
	mock := setupMockDB(t)

	// Expect existance of source account
	mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id =").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	body := `{"source_account_id":1,"destination_account_id":2,"amount":"10.00"}`
	req := httptest.NewRequest(http.MethodPost, "/transfer", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	handlers.TransactionHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// Fail: Destination account not found
func TestTransactionHandler_DestinationNotFound(t *testing.T) {
	mock := setupMockDB(t)

	// Expect existance of source account
	mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id =").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.0))

	body := `{"source_account_id":1,"destination_account_id":2,"amount":"10.00"}`
	req := httptest.NewRequest(http.MethodPost, "/transfer", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	handlers.TransactionHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// Fail: Source balance insufficient
func TestTransactionHandler_SourceBalanceInsufficient(t *testing.T) {
	mock := setupMockDB(t)

	// Expect source account SELECT with low balance
	mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id =").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(10.0))

	// Expect destination account SELECT with some balance
	mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id =").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(50.0))

	// No transaction or update calls expected here because balance check fails before.

	body := []byte(`{"source_account_id": 1, "destination_account_id": 2, "amount": "20.00"}`)
	req := httptest.NewRequest(http.MethodPost, "/transfer", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handlers.TransactionHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
