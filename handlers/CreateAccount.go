package handlers

import (
	"encoding/json"
	"httpserver/models"
	"httpserver/utils"
	"net/http"
	"strconv"
)

func CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	var input struct {
		AccountID      int    `json:"account_id"`
		InitialBalance string `json:"initial_balance"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	initialBalance, err := strconv.ParseFloat(input.InitialBalance, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "initial_balance must be a number")
		return
	}

	_, err = models.DB.Exec("INSERT INTO accounts (account_id, balance) VALUES ($1, $2)", input.AccountID, initialBalance)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
