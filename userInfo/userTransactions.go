package userInfo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
)

type AwardAmount struct {
	Rollno       int `json:"rollno"`
	AmountToSend int `json:"amount_to_send"`
}

type TransferAmount struct {
	SenderRollno   int `json:"sender_rollno"`
	RecieverRollno int `json:"reciever_rollno"`
	AmountToSend   int `json:"amount_to_send"`
}

type CurrentBalance struct {
	Rollno int `json:"rollno"`
}

type Response struct {
	Message string `json:"message"`
}

func AwardCoins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var awardAmount AwardAmount

	err := json.NewDecoder(r.Body).Decode(&awardAmount)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = UpdateAwardCoins(awardAmount)

	if err != nil {
		res := Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	res := Response{
		Message: "Transaction Successful",
	}
	json.NewEncoder(w).Encode(res)
}

func TransferCoins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var transferAmount TransferAmount

	err := json.NewDecoder(r.Body).Decode(&transferAmount)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = UpdateTransferCoins(transferAmount)

	if err != nil {
		res := Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	res := Response{
		Message: "Transaction Successful",
	}
	json.NewEncoder(w).Encode(res)

}

func CoinBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var currentBalance CurrentBalance

	err := json.NewDecoder(r.Body).Decode(&currentBalance)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "./user_info.db")
	CheckError(err)
	sqldb := FromSQLite(db)
	if !(sqldb.UserExists(currentBalance.Rollno)) {
		res := Response{
			Message: "User with this Rollno does not exists",
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	coins := sqldb.GetBalance(currentBalance.Rollno)

	res := Response{
		Message: "Current Balance: " + strconv.Itoa(coins),
	}
	json.NewEncoder(w).Encode(res)

}

//Called when user recieves certain amount of Coins through admin
func UpdateAwardCoins(awardAmount AwardAmount) error {
	db, err := sql.Open("sqlite3", "./user_info.db")
	CheckError(err)
	sqldb := FromSQLite(db)

	if !(sqldb.UserExists(awardAmount.Rollno)) {
		return errors.New("User with this Rollno does not exists")
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	CheckError(err)

	_, err = tx.ExecContext(ctx, "UPDATE user_info SET Coins = Coins + ? WHERE Rollno =?", awardAmount.AmountToSend, awardAmount.Rollno)
	if err != nil {
		// Incase we find any error in the query execution, rollback the transaction
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	CheckError(err)

	return nil
}

//Called when coins are transfered between users
func UpdateTransferCoins(transferAmount TransferAmount) error {
	db, err := sql.Open("sqlite3", "./user_info.db")
	CheckError(err)
	sqldb := FromSQLite(db)

	if !(sqldb.UserExists(transferAmount.SenderRollno)) {
		return errors.New("sender's rollno does not exists")
	}

	if !(sqldb.UserExists(transferAmount.RecieverRollno)) {
		return errors.New("reciever's rollno does not exists")
	}

	sendersBalance := sqldb.GetBalance(transferAmount.SenderRollno)
	if sendersBalance < transferAmount.AmountToSend {
		return errors.New("insufficient balance")
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	CheckError(err)

	lock := sync.RWMutex{}
	lock.RLock()

	_, err = tx.ExecContext(ctx,
		"UPDATE user_info SET Coins = Coins - ? WHERE Rollno = ? AND Coins - ?>=0", transferAmount.AmountToSend, transferAmount.SenderRollno, transferAmount.AmountToSend)
	if err != nil {
		// Incase we find any error in the query execution, rollback the transaction
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE user_info SET Coins = Coins + ? WHERE Rollno = ?", transferAmount.AmountToSend, transferAmount.RecieverRollno)
	if err != nil {
		// Incase we find any error in the query execution, rollback the transaction
		tx.Rollback()
		return err
	}

	lock.RUnlock()

	err = tx.Commit()
	CheckError(err)

	return nil
}
