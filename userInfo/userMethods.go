package userInfo

import (
	"database/sql"
	"errors"
	"log"
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

//Get gets users from user_info
func (s *SQLite) Get() []User {
	users := []User{}
	rows, _ := s.DB.Query("SELECT * FROM user_info")
	defer rows.Close()
	var id int
	var rollno int
	var password string
	var name string
	for rows.Next() {
		rows.Scan(&id, &rollno, &password, &name)
		users = append(users, User{
			ID:     id,
			Rollno: rollno,
			Name:   name,
		})
	}
	return users
}

//To check if the user already exists in the database
func (s *SQLite) UserExists(user RegisterRequest) bool {
	rollno := user.Rollno
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
func (s *SQLite) Add(user RegisterRequest) error{
	stmt, err := s.DB.Prepare(
		"INSERT INTO user_info(Rollno, Name, Password) VALUES (?, ?, ?)")
	CheckError(err)

	hashedPassword := HashAndSalt(user.Password)
	
	if !(s.UserExists(user)) {
		stmt.Exec(user.Rollno, user.Name, hashedPassword)
		return nil
	}

	return errors.New("User Already exists")
}

//Creates a table user_info using sqlite
func FromSQLite(db *sql.DB) *SQLite {
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS "user_info" (
		"ID"	  INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"Rollno"  INT,
		"Name"    TEXT,
		"Password"  TEXT
	  );
	`)

	CheckError(err)
	stmt.Exec()

	return &SQLite{
		DB: db,
	}
}
