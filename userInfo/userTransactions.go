package userInfo

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aman7625/iitk-coin/middleware"
)

type AwardAmount struct {
	Rollno            int64   `json:"rollno"`
	AmountToSend      float64 `json:"amountToSend"`
	FromCouncilMember string  `json:"fromCouncilMember"`
	FromAdmin         string  `json:"fromAdmin,omitempty"`
}

type TransferAmount struct {
	RecieverRollno int64   `json:"reciever_rollno"`
	AmountToSend   float64 `json:"amount_to_send"`
}

type TransferAmountDTO struct {
	SenderRollno   int64
	RecieverRollno int64
	AmountToSend   float64
}

type RedeemRequest struct {
	NumberOfCoins float64 `json:"numberOfCoins"`
	ItemName      string  `json:"itemName"`
}

type RedeemRequestDTO struct {
	NumberOfCoins float64
	ItemName      string
	Sender        int64
	Status        string
}

type DestroyUser struct {
	Prefix               string `json:"prefix"`
	IsActionTakenByAdmin string `json:"isActionTakenByAdmin"`
}

type CurrentBalance struct {
	Rollno int64 `json:"rollno"`
}

type Response struct {
	Message string `json:"message"`
}

type TransactionDetail struct {
	transactionType string
	sender          int64
	reciever        int64
	amountSent      float64
	amountRecieved  float64
	tax             string
	dateTime        string
}

//Called when "/award" endpoint is hit
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

//Called when "/transfer" endpoint is hit
func TransferCoins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var transferAmount TransferAmount

	//Aunthenticate the user sending coins
	rollno, err := middleware.UserAuthentication(w, r)
	if err != nil {
		res := Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&transferAmount)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	transferdto := TransferAmountDTO{SenderRollno: rollno, RecieverRollno: transferAmount.RecieverRollno, AmountToSend: transferAmount.AmountToSend}

	err = UpdateTransferCoins(transferdto)

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

//Called when "/view" endpoint is hit
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
		Message: "Current Balance: " + strconv.FormatFloat(coins, 'f', 2, 64),
	}
	json.NewEncoder(w).Encode(res)

}

func RedeemCoins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var redeemRequest RedeemRequest

	rollno, err := middleware.UserAuthentication(w, r)
	if err != nil {
		res := Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	//fmt.Println(rollno)

	err = json.NewDecoder(r.Body).Decode(&redeemRequest)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Initial Check to see if user has sufficient balance
	db, err := sql.Open("sqlite3", "./user_info.db")
	CheckError(err)
	sqldb := FromSQLite(db)
	coins := sqldb.GetBalance(rollno)

	if coins < redeemRequest.NumberOfCoins {
		res := Response{
			Message: "Insufficient Balance",
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	redeem := RedeemRequestDTO{NumberOfCoins: redeemRequest.NumberOfCoins, ItemName: redeemRequest.ItemName, Sender: rollno, Status: "pending"}

	//Adding redeem request in table
	db, err = sql.Open("sqlite3", "./user_info.db")
	CheckError(err)
	sqldb = RedeemRequestTable(db)
	sqldb.AddRequest(redeem)

	//Else redeem request is sent
	res := Response{
		Message: "Redeem Request has been sent",
	}
	json.NewEncoder(w).Encode(res)
}

//Called when /destroy endpoint is hit
func DestroyGraduatingBatchAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var destroyUser DestroyUser

	err := json.NewDecoder(r.Body).Decode(&destroyUser)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if destroyUser.IsActionTakenByAdmin == "false" {
		res := Response{
			Message: "You don't have authorization to use this method",
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	lowerbound, err := strconv.Atoi(destroyUser.Prefix + "0000")
	CheckError(err)
	upperbound, _ := strconv.Atoi(destroyUser.Prefix + "9999")
	CheckError(err)

	db, err := sql.Open("sqlite3", "./user_info.db")
	CheckError(err)
	sqldb := FromSQLite(db)

	stmt, err := sqldb.DB.Prepare("DELETE FROM user_info WHERE Rollno > ? AND Rollno < ?")
	CheckError(err)
	stmt.Exec(lowerbound, upperbound)

	res := Response{
		Message: "Accounts Destroyed Successfully",
	}
	json.NewEncoder(w).Encode(res)
}
