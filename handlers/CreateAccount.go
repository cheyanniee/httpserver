package handlers

import (
	"encoding/json"
	"httpserver/models"
	"httpserver/utils"
	"net/http"
	"strconv"
)

// Handler to create account
func CreateAccountHandler(w http.ResponseWriter, r *http.Request) {

	// Ensure usage of POST method
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Input structure
	var input struct {
		AccountID      int    `json:"account_id"`
		InitialBalance string `json:"initial_balance"`
	}

	// Verify JSON is valid
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Verify initial_balance is a number
	initialBalance, err := strconv.ParseFloat(input.InitialBalance, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "initial_balance must be a number")
		return
	}

	// Create new account with input details
	_, err = models.DB.Exec("INSERT INTO accounts (account_id, balance) VALUES ($1, $2)", input.AccountID, initialBalance)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	// Empty response if successful
	w.WriteHeader(http.StatusNoContent)
}
