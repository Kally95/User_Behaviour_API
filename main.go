package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
)

const (
	DB_USER = "postgres"
	DB_PASS = "secret"
	DB_NAME = "user_behaviour_api"
)

var (
	ErrEmptyUsername = errors.New("Username cannot be empty")
)

// User creates a user instance
type User struct {
	Username string `json:"name"`
	Password string `json:"password"`
}

type DB interface {
	Insert(user User) error
}

type PostgresDBObject struct {
	db *sql.DB
}

// Opendb opens the database with the postgres credentials.
func OpenDB() *sql.DB {

	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DB_USER, DB_PASS, DB_NAME)

	db, err := sql.Open("postgres", dbinfo)

	if err != nil {
		panic(err)
	}

	return db
}

func (p PostgresDBObject) Insert(user User) error {
	sqlStatement := `
		INSERT INTO users (username, password)
		VALUES ($1, $2)`

	_, err := p.db.Exec(sqlStatement, user.Username, user.Password)

	if err != nil {
		return err
		//errors.New("db_error: couldn't execute sql statement")
	}
	return nil
}

func InsertIntoDB(db DB, user User) {
	err := db.Insert(user)
	if err != nil {
		fmt.Println(err)
	}
}

// NormaliseUsername will clean the input data for
// the user's username before passing it is executed
// by the DB query.
func NormaliseUsername(user *User) error {
	if user.Username == "" {
		return ErrEmptyUsername
	}
	user.Username = strings.ToLower(user.Username)
	user.Username = strings.TrimSpace(user.Username)
	user.Username = strings.Title(user.Username)
	return nil
}

// SignUp processes incoming POST request for users who
// wish to log in.
func SignUp(w http.ResponseWriter, r *http.Request) {
	var user User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		fmt.Print(err)
	}

	var db DB

	db = PostgresDBObject{
		db: OpenDB(),
	}

	if err := NormaliseUsername(&user); err != nil {
		fmt.Println(err)
	}

	// InsertIntoDB(db, u)
	InsertIntoDB(db, user)
}

// LogIn processes incoming GET requests for users who
// wish to sign up.
func LogIn(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func main() {
	// Wipes existing entries from DB.
	//internal.DropDB()
	//internal.BuildDB()
	fmt.Println("Successfully connected")
	http.HandleFunc("/signup", SignUp)
	http.HandleFunc("/login", LogIn)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
