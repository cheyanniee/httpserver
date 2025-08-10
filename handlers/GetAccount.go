package handlers

import (
	"database/sql"
	"fmt"
	"httpserver/models"
	"httpserver/utils"
	"net/http"
	"strconv"
	"strings"
)

// Helper function to be used for other handlers as well
func GetAccountByID(accountID int) (*models.Account, error) {

	// Store balance
	var balance float64

	// Query for account
	err := models.DB.QueryRow("SELECT balance FROM accounts WHERE account_id = $1", accountID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, err
	}

	// Store account to be returned
	acc := &models.Account{
		AccountID:      accountID,
		CurrentBalance: balance,
	}

	return acc, nil
}

// Handler to get account
func GetAccountHandler(w http.ResponseWriter, r *http.Request) {

	// Ensure usage of GET method
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Extract account ID and verify it is a number
	accountIDStr := strings.TrimPrefix(r.URL.Path, "/accounts/")
	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid account ID")
		return
	}

	// Query for account using helper function above
	acc, err := GetAccountByID(accountID)
	if err != nil {
		if err.Error() == "account not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// JSON response with account details if successful
	utils.WriteJSON(w, http.StatusOK, acc)
}
