package userInfo

import (
	"database/sql"
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
	var name string
	for rows.Next() {
		rows.Scan(&id, &rollno, &name)
		users = append(users, User{
			ID:     id,
			Rollno: rollno,
			Name:   name,
		})
	}
	return users
}

//To check if the user already exists
func (s *SQLite) UserExists(user User) bool {
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
func (s *SQLite) Add(user User) {
	stmt, err := s.DB.Prepare(
		"INSERT INTO user_info(Rollno, Name) VALUES (?, ?)")
	CheckError(err)
	if !(s.UserExists(user)) {
		stmt.Exec(user.Rollno, user.Name)
	}
}


//Creates a table user_info using sqlite
func FromSQLite(db *sql.DB) *SQLite {
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS "user_info" (
		"ID"	  INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"Rollno"  INT,
		"Name"    TEXT
	  );
	`)

	CheckError(err)
	stmt.Exec()

	return &SQLite{
		DB: db,
	}
}
