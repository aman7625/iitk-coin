package userInfo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/aman7625/iitk-coin/middleware"
)

type Action struct {
	Id     int    `json:"id"`
	Status string `json:"status"`
}

//Called when user recieves certain amount of Coins through admin
func UpdateAwardCoins(awardAmount AwardAmount) error {
	db, err := sql.Open("sqlite3", "./user_info.db")
	CheckError(err)
	sqldb := FromSQLite(db)

	if !(sqldb.UserExists(awardAmount.Rollno)) {
		return errors.New("User with this Rollno does not exists")
	}

	if !(sqldb.isRewardValid(awardAmount)) {
		return errors.New("you do not satisfy the criteria for getting rewarded")
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

	dt := time.Now()
	detail := TransactionDetail{transactionType: "reward", sender: 0, reciever: awardAmount.Rollno, amountSent: 0, amountRecieved: awardAmount.AmountToSend, tax: "nil", dateTime: dt.String()}

	db, err = sql.Open("sqlite3", "./user_info.db")
	CheckError(err)
	sqldb = TransactionHistory(db)
	sqldb.AddTransaction(detail)

	return nil
}

//Called when coins are transfered between users
func UpdateTransferCoins(transferAmount TransferAmountDTO) error {
	db, err := sql.Open("sqlite3", "./user_info.db")
	CheckError(err)
	sqldb := FromSQLite(db)

	if !(sqldb.UserExists(transferAmount.SenderRollno)) {
		return errors.New("sender's rollno does not exists")
	}

	if !(sqldb.UserExists(transferAmount.RecieverRollno)) {
		return errors.New("reciever's rollno does not exists")
	}

	if transferAmount.SenderRollno == transferAmount.RecieverRollno {
		return errors.New("you cannot transfer coins to your own account")
	}

	sendersBalance := sqldb.GetBalance(transferAmount.SenderRollno)
	if sendersBalance < transferAmount.AmountToSend {
		return errors.New("insufficient balance")
	}

	tsqldb := TransactionHistory(db)

	//find the number of events sender has participated
	numberOfEventsParticipated := tsqldb.eventsParticipated(transferAmount.SenderRollno)
	threshold := 2

	//return if user hasn't participated in sufficient number of events
	if numberOfEventsParticipated < threshold {
		return errors.New("you are not eligible to make transaction, since you've not particpiated in enough events")
	}

	//find the applicable tax on transaction
	tax := GetTaxValue(transferAmount.SenderRollno, transferAmount.RecieverRollno)

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
		"UPDATE user_info SET Coins = Coins + ? WHERE Rollno = ?", transferAmount.AmountToSend-transferAmount.AmountToSend*tax, transferAmount.RecieverRollno)
	if err != nil {
		// Incase we find any error in the query execution, rollback the transaction
		tx.Rollback()
		return err
	}

	lock.RUnlock()

	err = tx.Commit()
	CheckError(err)

	dt := time.Now()

	//storing transaction details in database
	detail := TransactionDetail{transactionType: "transfer", sender: transferAmount.SenderRollno, reciever: transferAmount.RecieverRollno, amountSent: transferAmount.AmountToSend, amountRecieved: transferAmount.AmountToSend - transferAmount.AmountToSend*tax, tax: strconv.FormatFloat(tax*100, 'f', 0, 64) + "%",
		dateTime: dt.String()}

	tsqldb.AddTransaction(detail)

	return nil
}

//Function called when Admin takes action on pending Redeem Requests
func TakeAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var action Action

	db, err := sql.Open("sqlite3", "./user_info.db")
	CheckError(err)

	//Authenticating whether user is Admin
	AdminRollno, err := middleware.UserAuthentication(w, r)
	if err != nil {
		res := Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(res)
		return
	}
	userdb := FromSQLite(db)
	isAdmin := userdb.isUserAdmin(AdminRollno)
	if !isAdmin {
		res := Response{
			Message: "Only Admins can perform this action",
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&action)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sqldb := RedeemRequestTable(db)

	if action.Status == "rejected" {
		sqldb.ModifyStatus(action)
		res := Response{
			Message: "Successfully rejected redeem request",
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	//If action is approved need to check redeemer's balance and update it
	//If insufficient balance status would be rejected

	//Get Sender's rollno and Amount to Redeem using Id from action object
	rollno, amountToRedeem := sqldb.GetSender(action)

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		res := Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE user_info SET Coins = Coins - ? WHERE Rollno = ? AND Coins - ?>=0", amountToRedeem, rollno, amountToRedeem)
	if err != nil {
		// Incase we find any error in the query execution, rollback the transaction
		tx.Rollback()
		res := Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	err = tx.Commit()
	if err != nil {
		res := Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	dt := time.Now()
	detail := TransactionDetail{transactionType: "redeem", sender: rollno, reciever: 0, amountSent: amountToRedeem, amountRecieved: 0, tax: "nil", dateTime: dt.String()}

	tsqldb := TransactionHistory(db)
	tsqldb.AddTransaction(detail)

	//If everything goes fine, modify the status to approved
	sqldb.ModifyStatus(action)
	res := Response{
		Message: "Appropriate Action Taken Successfully",
	}
	json.NewEncoder(w).Encode(res)
}
