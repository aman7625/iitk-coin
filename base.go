package main

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func createTable(db *sql.DB) {
	createUserTable := `CREATE TABLE user (
		"id" integer PRIMARY KEY,
		"rollno" INT,
		"name" STRING
	  );`

	stmt, err := db.Prepare(createUserTable) // Prepare SQL stmt
	checkError(err)
	stmt.Exec() // Execute SQL stmts
}

func addUser(db *sql.DB, rollno int, name string) {
	addUserSQL := `INSERT INTO user(rollno, name) VALUES (?, ?)`

	stmt, err := db.Prepare(addUserSQL)
	checkError(err)

	_, err = stmt.Exec(rollno, name)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	file, err := os.Create("database.db") // Creating database
	checkError(err)
	file.Close()

	sqlitedb, _ := sql.Open("sqlite3", "./database.db") // Connecting to database
	defer sqlitedb.Close()
	createTable(sqlitedb) // Creating Table

	// Adding Users
	addUser(sqlitedb, 191000, "Frodo Baggins")
	addUser(sqlitedb, 191001, "Bilbo Baggins")
	addUser(sqlitedb, 191002, "Samwise Gamgee")
	addUser(sqlitedb, 191003, "Gandalf")
	addUser(sqlitedb, 191004, "Saruman")

}
