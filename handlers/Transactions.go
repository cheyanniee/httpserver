package handlers

import (
	"encoding/json"
	"errors"
	"httpserver/models"
	"httpserver/utils"
	"net/http"
	"strconv"
)

// Helper function to transfer currency
func TransferCurrency(sourceAcc models.Account, destAcc models.Account, amount float64) error {

	// Multiply by 100000 for higher accuracy when subtracting
	tempSourceBalance := sourceAcc.CurrentBalance * 100000
	tempDestinationBalance := destAcc.CurrentBalance * 100000
	convertedAmount := amount * 100000

	// Divide by 100000 for storage
	newSourceBalance := (tempSourceBalance - convertedAmount) / 100000
	newDestinationBalance := (tempDestinationBalance + convertedAmount) / 100000

	// DB begin
	tx, err := models.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		// rollback if the transaction is still active
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	// Update source
	res1, err := tx.Exec(
		"UPDATE accounts SET balance = $1 WHERE account_id = $2",
		newSourceBalance, sourceAcc.AccountID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	rows1, err := res1.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if rows1 == 0 {
		tx.Rollback()
		return errors.New("source account not found")
	}

	// Update destination
	res2, err := tx.Exec(
		"UPDATE accounts SET balance = $1 WHERE account_id = $2",
		newDestinationBalance, destAcc.AccountID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	rows2, err := res2.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if rows2 == 0 {
		tx.Rollback()
		return errors.New("destination account not found")
	}

	// Commit if all successful
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Handler for transactions
func TransactionHandler(w http.ResponseWriter, r *http.Request) {

	// Ensure usage of POST method
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Input structure
	var input struct {
		SourceAcc      int    `json:"source_account_id"`
		DestinationAcc int    `json:"destination_account_id"`
		Amount         string `json:"amount"`
	}

	// Verify JSON is valid
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Verify amount is a number
	amount, err := strconv.ParseFloat(input.Amount, 64)
	if err != nil || amount <= 0 {
		utils.WriteError(w, http.StatusBadRequest, "amount must be a positive number")
		return
	}
	// Verify source account exists
	source, err := GetAccountByID(input.SourceAcc)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "source account not found")
		return
	}

	// Verify destination account exists
	dest, err := GetAccountByID(input.DestinationAcc)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "destination account not found")
		return
	}

	// Verify source account has sufficient balance
	if source.CurrentBalance < amount {
		utils.WriteError(w, http.StatusBadRequest, "insufficient balance in source account")
		return
	}

	// Attempt to transfer the currency
	err = TransferCurrency(*source, *dest, amount)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch updated balances from DB
	updatedSource, _ := GetAccountByID(input.SourceAcc)
	updatedDest, _ := GetAccountByID(input.DestinationAcc)

	// If successful, provide current balances
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"source_account_id": updatedSource.AccountID,
		"source_balance":    updatedSource.CurrentBalance,
		"dest_account_id":   updatedDest.AccountID,
		"dest_balance":      updatedDest.CurrentBalance,
	})
}
