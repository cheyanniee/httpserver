package handlers

import (
	"encoding/json"
	"errors"
	"httpserver/models"
	"httpserver/utils"
	"net/http"
	"strconv"
)

func TransferCurrency(sourceAcc models.Account, destAcc models.Account, amount float64) error {
	tempSourceBalance := sourceAcc.CurrentBalance * 100000
	tempDestinationBalance := destAcc.CurrentBalance * 100000
	convertedAmount := amount * 100000

	newSourceBalance := (tempSourceBalance - convertedAmount) / 100000
	newDestinationBalance := (tempDestinationBalance + convertedAmount) / 100000

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

	// Source update
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

	// Destination update
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

	// All good: commit
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func TransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	var input struct {
		SourceAcc      int    `json:"source_account_id"`
		DestinationAcc int    `json:"destination_account_id"`
		Amount         string `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	amount, err := strconv.ParseFloat(input.Amount, 64)
	if err != nil || amount <= 0 {
		utils.WriteError(w, http.StatusBadRequest, "amount must be a positive number")
		return
	}

	source, err := GetAccountByID(input.SourceAcc)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "source account not found")
		return
	}

	dest, err := GetAccountByID(input.DestinationAcc)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "destination account not found")
		return
	}

	if source.CurrentBalance < amount {
		utils.WriteError(w, http.StatusBadRequest, "insufficient balance in source account")
		return
	}

	err = TransferCurrency(*source, *dest, amount)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
