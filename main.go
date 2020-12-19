package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
)

var (
	// ErrEmptyUsername is returned when a given username
	// is empty during NormalizeUsername and CheckUsername.
	ErrEmptyUsername = errors.New("Username cannot be empty")

	// ErrUserNameTaken is returned when CheckUsername finds
	// a duplicate of the username passed.
	ErrUserNameExists = errors.New("This username already exists")

	// ErrDBError is returned when there's a DB error that is not
	// rowNotFound.
	ErrDBError = errors.New("There was an error querying the databse")

	// ErrPwdIncorrect is returned when a user tries to log in and
	// the password provided does not match.
	ErrPwdIncorrect = errors.New("Password's do not match")
)

const (
	DB_USER = "postgres"
	DB_PASS = "secret"
	DB_NAME = "user_behaviour_api"
)

// DB is a interface that implements methods
// of our database object to satisfy the interface
// condition.
type DB interface {
	Insert(user User) error
	CheckUsername(user User) bool
	PasswordCheck(user User) bool
}

// PostgresDBObject represents a PSQL databse object.
type PostgresDBObject struct {
	db *sql.DB
}

// User creates a user instance
type User struct {
	Username string `json:"name"`
	Password string `json:"password"`
}

// OpenDB opens the database with the postgres credentials.
func OpenDB() *sql.DB {

	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DB_USER, DB_PASS, DB_NAME)

	db, err := sql.Open("postgres", dbinfo)

	if err != nil {
		panic(ErrDBError)
	}

	return db
}

// ======================= DB INTERFACE METHODS =======================

// InsertIntoDB uses a given DB instance to insert
// a user into a given databse.
func InsertIntoDB(db DB, user User) {
	err := db.Insert(user)
	if err != nil {
		fmt.Println(err)
	}
}

// CheckUserName calls the CheckUsername method of
// a DB object, this can be Mongo, SQL etc.
// In this case we are calling the method from
// PSQL.
func CheckUserName(db DB, user User) (bool, error) {
	exists := db.CheckUsername(user)
	if exists {
		return true, ErrUserNameExists
	}
	return false, nil
}

// CheckPassword calls the PasswordCheck method of
// a DB object, this can be Mongo, SQLetc.
// In this case we are calling the method from
// PSQL.
func CheckPassword(db DB, user User) (bool, error) {
	match := db.PasswordCheck(user)
	if match {
		return true, nil
	}
	return false, ErrPwdIncorrect
}

// ======================= END =======================

// ======================= PSQL METHODS =======================

// Insert is a method attached to PostgresDBObject
// that inserts a user into the database.
func (p PostgresDBObject) Insert(user User) error {
	sqlStatement := `
		INSERT INTO users (username, password)
		VALUES ($1, $2)`

	_, err := p.db.Exec(sqlStatement, user.Username, user.Password)

	if err != nil {
		return err
	}
	return nil
}

// NormaliseUsername will clean the input data for
// the user's username before passing it is executed
// by the DB query.
func (p PostgresDBObject) NormaliseUsername(user User) error {
	if user.Username == "" {
		return ErrEmptyUsername
	}
	user.Username = strings.ToLower(user.Username)
	user.Username = strings.TrimSpace(user.Username)
	user.Username = strings.Title(user.Username)
	return nil
}

// CheckUsername checks whether the username is already
// in use.
func (p PostgresDBObject) CheckUsername(user User) bool {
	sqlStmt := `SELECT username FROM users WHERE username=$1;`
	row := p.db.QueryRow(sqlStmt)
	err := row.Scan(user.Username)
	if err != nil {
		return false
	}
	return true
}

// PasswordCheck checks whether the password matches
// the given password.
func (p PostgresDBObject) PasswordCheck(user User) bool {
	sqlStmt := `SELECT password FROM users WHERE password=$1;`
	row := p.db.QueryRow(sqlStmt)
	err := row.Scan(user.Password)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(ErrDBError)
		}
		return false
	}
	return true
}

// ======================= END =======================

// ======================= HANDLERS =======================

// SignUp processes incoming POST request for users who
// wish to log in.
func SignUp(w http.ResponseWriter, r *http.Request) {
	var user User
	var p PostgresDBObject
	var db DB

	db = PostgresDBObject{
		db: OpenDB(),
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		fmt.Print(err)
	}

	if err := p.NormaliseUsername(user); err != nil {
		log.Print(err)
	}

	_, err := CheckUserName(db, user)
	if err != nil {
		if err == ErrUserNameExists {
			http.Redirect(w, r, "/signup", http.StatusBadRequest)
		}
		log.Print(ErrDBError)
	}

	InsertIntoDB(db, user)
}

// LogIn processes incoming GET requests for users who
// wish to sign up.
func LogIn(w http.ResponseWriter, r *http.Request) {
	var user User
	var db DB

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		fmt.Print(err)
	}

	exists, err := CheckUserName(db, user)
	if err != ErrUserNameExists {
		log.Print(ErrDBError)
	}

	pwdMatch, err := CheckPassword(db, user)
	if err != nil {
		log.Print(ErrPwdIncorrect)
	}

	if pwdMatch && exists {
		w.WriteHeader(200)
		fmt.Println("You successfully logged in")
	}

}

// ======================= END =======================

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
