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

func GetAccountByID(accountID int) (*models.Account, error) {
	var balance float64

	err := models.DB.QueryRow("SELECT balance FROM accounts WHERE account_id = $1", accountID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, err
	}

	acc := &models.Account{
		AccountID:      accountID,
		CurrentBalance: balance,
	}

	return acc, nil
}

func GetAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	accountIDStr := strings.TrimPrefix(r.URL.Path, "/accounts/")
	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid account ID")
		return
	}

	acc, err := GetAccountByID(accountID)
	if err != nil {
		if err.Error() == "account not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, acc)
}
