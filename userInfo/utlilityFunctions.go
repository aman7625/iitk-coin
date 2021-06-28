package userInfo

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"sync"
	"time"
)

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

	db, err = sql.Open("sqlite3", "./transaction_history.db")
	CheckError(err)
	sqldb = TransactionHistory(db)
	sqldb.AddTransaction(detail)

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

	if transferAmount.SenderRollno == transferAmount.RecieverRollno {
		return errors.New("you cannot transfer coins to your own account")
	}

	sendersBalance := sqldb.GetBalance(transferAmount.SenderRollno)
	if sendersBalance < transferAmount.AmountToSend {
		return errors.New("insufficient balance")
	}

	tdb, err := sql.Open("sqlite3", "./transaction_history.db")
	CheckError(err)
	tsqldb := TransactionHistory(tdb)

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

//Called when User Redeems Coin
func UpdateRedeemCoins(redeemAmount RedeemAmount) error {
	db, err := sql.Open("sqlite3", "./user_info.db")
	CheckError(err)

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	CheckError(err)

	_, err = tx.ExecContext(ctx, "UPDATE user_info SET Coins = Coins - ? WHERE Rollno =?", redeemAmount.AmountToRedeem, redeemAmount.Rollno)
	if err != nil {
		// Incase we find any error in the query execution, rollback the transaction
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	CheckError(err)

	dt := time.Now()
	detail := TransactionDetail{transactionType: "redeem", sender: redeemAmount.Rollno, reciever: 0, amountSent: redeemAmount.AmountToRedeem, amountRecieved: 0, tax: "nil", dateTime: dt.String()}

	db, err = sql.Open("sqlite3", "./transaction_history.db")
	CheckError(err)
	sqldb := TransactionHistory(db)
	sqldb.AddTransaction(detail)

	return nil
}
