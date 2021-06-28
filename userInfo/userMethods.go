package userInfo

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
)

type SQLite struct {
	DB *sql.DB
}

//CheckError
func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

//To check if the user already exists in the database
func (s *SQLite) UserExists(id int64) bool {
	rollno := id
	sqlStmt := `SELECT Rollno FROM user_info WHERE rollno = ?`
	err := s.DB.QueryRow(sqlStmt, rollno).Scan(&rollno)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}

		return false
	}

	return true
}

//Add adds users into user_info
func (s *SQLite) Add(user RegisterRequest) error {
	stmt, err := s.DB.Prepare(
		"INSERT INTO user_info(Rollno, Name, Password, isCouncilMember, isAdmin) VALUES (?, ?, ?, ?, ?)")
	CheckError(err)

	hashedPassword := HashAndSalt(user.Password)

	if !(s.UserExists(user.Rollno)) {
		if user.IsCouncilMember == "" && user.IsAdmin == "" {
			stmt.Exec(user.Rollno, user.Name, hashedPassword, "No", "No")
			return nil
		}
		if user.IsCouncilMember == "" {
			stmt.Exec(user.Rollno, user.Name, hashedPassword, "No", user.IsAdmin)
			return nil
		}
		if user.IsAdmin == "" {
			stmt.Exec(user.Rollno, user.Name, hashedPassword, user.IsCouncilMember, "No")
			return nil
		}
		stmt.Exec(user.Rollno, user.Name, hashedPassword, user.IsCouncilMember, user.IsAdmin)
		return nil
	}

	return errors.New("User Already exists")
}

//Adding Transaction Details in a table
func (s *SQLite) AddTransaction(detail TransactionDetail) {
	stmt, err := s.DB.Prepare(
		"INSERT INTO transaction_history(Type,Sender,Reciever,AmountSent,AmountRecieved,Tax,TransactionDate) VALUES (?,?,?,?,?,?,?)")
	CheckError(err)
	stmt.Exec(detail.transactionType, detail.sender, detail.reciever, detail.amountSent, detail.amountRecieved, detail.tax, detail.dateTime)
}

//Creates a table user_info using sqlite
func FromSQLite(db *sql.DB) *SQLite {
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS "user_info" (
		"ID"	  INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"Rollno"  INT,
		"Name"    TEXT,
		"Password"  TEXT,
		"isCouncilMember" TEXT NOT NULL DEFAULT "No",
		"isAdmin" TEXT  NOT NULL DEFAULT "No", 
		"Coins"  REAL DEFAULT 0
	  );
	`)

	CheckError(err)
	stmt.Exec()

	return &SQLite{
		DB: db,
	}
}

func TransactionHistory(db *sql.DB) *SQLite {
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS "transaction_history" (
		"ID"	  INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"Type" TEXT,
		"Sender"  INT,
		"Reciever"   INT,
		"AmountSent"  REAL,
		"AmountRecieved" REAL,
		"Tax" REAL,
		"TransactionDate" TEXT
	  );
	`)

	CheckError(err)
	stmt.Exec()

	return &SQLite{
		DB: db,
	}
}

//GetBalance gets the current Balance of the user
func (s *SQLite) GetBalance(id int64) float64 {
	rollno := id
	var coins float64
	err := s.DB.QueryRow("SELECT Coins FROM user_info WHERE rollno=?", rollno).Scan(&coins)
	CheckError(err)

	return coins
}

//Check the nature of Tax applicable on the Transaction
func GetTaxValue(senderId int64, recieverId int64) float64 {
	str := strconv.Itoa(int(senderId))
	sender := str[:2]
	str = strconv.Itoa(int(recieverId))
	reciever := str[:2]
	if sender == reciever {
		return 0.02
	}
	return 0.33
}

//Return the number of events that a user has participated in
func (s *SQLite) eventsParticipated(userId int64) int {
	count := 0
	err := s.DB.QueryRow("SELECT COUNT(*) FROM transaction_history WHERE Type='reward' AND Reciever=?", userId).Scan(&count)
	CheckError(err)

	return count
}

//Check if a reward is Valid
//Council Member can gain coin only through admin
//Admins(Gensec and AH) cannot reward their accounts with coins
func (s *SQLite) isRewardValid(awardAmount AwardAmount) bool {
	var isUserAdmin string
	var isUserCouncilMember string

	err := s.DB.QueryRow("SELECT isCouncilMember,isAdmin FROM user_info WHERE rollno=?", awardAmount.Rollno).Scan(&isUserCouncilMember, &isUserAdmin)
	CheckError(err)

	//If coin is rewarded to a normal user
	if isUserCouncilMember == "No" && isUserAdmin == "No" {
		if awardAmount.FromCouncilMember == "Yes" || awardAmount.FromAdmin == "Yes" {
			return true
		}
	}

	//if reward method is used for council member
	if isUserCouncilMember == "Yes" {
		if awardAmount.FromAdmin == "Yes" {
			return true
		}
	}

	return false
}
