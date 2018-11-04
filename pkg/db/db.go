package db

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // init sql driver
	"golang.org/x/oauth2"
)

// User DB model
type User struct {
	UserID   string
	UserName string
}

// Weight DB model
type Weight struct {
	Weight    float64
	Timestamp int64 // epoch time (secs since 1970)
}

// WithingsToken DB model
type WithingsToken struct {
	UserID string
	Token  oauth2.Token
}

// Init a SQLite db
func Init() *sql.DB {

	dbPath := filepath.Join(os.Getenv("HOME"), ".config", "wfsync", "wfsync.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Failed to open DB: ", err)
	}

	// create user table:
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS users 
					(userId TEXT NOT NULL PRIMARY KEY,
					 userName TEXT NOT NULL)`)

	if err != nil {
		log.Fatal(err)
	}

	// create withings token table:
	_, err = db.Exec(
		// token is a json-encoded oauth2.Token
		`CREATE TABLE IF NOT EXISTS withingsTokens
					(id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
					 userId TEXT NOT NULL,
					 token TEXT NOT NULL, 
					 FOREIGN KEY(userId) REFERENCES users(userId))`)

	if err != nil {
		log.Fatal(err)
	}

	// create weight measurements table:
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS weights
					(id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
					 userId TEXT NOT NULL,
					 weight FLOAT NOT NULL,
					 timestamp INTEGER NOT NULL,
					 FOREIGN KEY(userId) REFERENCES users(userId))`)

	if err != nil {
		log.Fatal(err)
	}

	return db
}

// UserGet looks up a user by user id
func UserGet(db *sql.DB, userID string) (User, bool) {
	rows, err := db.Query("SELECT * FROM users where userId=?", userID)
	if err != nil {
		log.Fatal("Failed to query for user: ", err)
	}
	defer rows.Close()

	exists := rows.Next()

	user := User{}

	if exists {
		err = rows.Scan(&user.UserID, &user.UserName)
		if err != nil {
			log.Fatal("Failed to read row: ", err)
		}
	}

	return user, exists
}

// UserSave saves a user by id
func UserSave(db *sql.DB, userID string, userName string) {

	_, exists := UserGet(db, userID)

	if exists {
		log.Printf("User %s already exists.", userID)
		return
	}

	log.Printf("Saving user %s", userID)
	_, err := db.Exec("INSERT INTO users (userid, username) VALUES (?, ?)", userID, userName)
	if err != nil {
		log.Fatal("Failed to insert user: ", err)
	}
}

// WeightExists checks if a weight value is already saved
func WeightExists(db *sql.DB, userID string, weight Weight) (exists bool) {
	rows, err := db.Query("SELECT * FROM weights where userId=? AND weight=? AND timestamp=?",
		userID, weight.Weight, weight.Timestamp)
	if err != nil {
		log.Fatal("Failed to query for weight: ", err)
	}
	defer rows.Close()

	exists = rows.Next()
	return exists
}

// WeightSave saves a single weight measurement
func WeightSave(db *sql.DB, userID string, weight Weight) {
	exists := WeightExists(db, userID, weight)

	if !exists {
		_, err := db.Exec("INSERT INTO weights (userId, weight, timestamp) VALUES (?, ?, ?)",
			userID, weight.Weight, weight.Timestamp)

		if err != nil {
			log.Fatal("Failed to insert weight: ", err)
		}

	}
}

// WeightsSync saves a number of weight measurements for the user
func WeightsSync(db *sql.DB, userID string, weights []Weight) {
	for _, weight := range weights {
		WeightSave(db, userID, weight)
	}
}

// WithingsTokenGet retrieves a withings token, if one was previously saved
func WithingsTokenGet(db *sql.DB, user User) (*oauth2.Token, bool) {

	rows, err := db.Query("SELECT token FROM withingsTokens where userId=?", user.UserID)

	if err != nil {
		log.Fatal("An error occurred fetching a token: ", err)
	}

	defer rows.Close()

	token := oauth2.Token{}
	exists := rows.Next()

	if exists {
		var buf string
		err = rows.Scan(&buf)
		if err != nil {
			log.Fatal("Failed to scan row: ", err)
		}
		err = json.Unmarshal([]byte(buf), &token)
		if err != nil {
			log.Fatal("Failed to unmarshal token: ", err)
		}
	}

	return &token, exists
}

// WithingsTokenSave save the withings token in the user record
func WithingsTokenSave(db *sql.DB, user User, token *oauth2.Token) {

	buf, err := json.Marshal(token)
	if err != nil {
		log.Fatal("Failed to marshal token: ", err)
	}
	tokenStr := string(buf)

	_, exists := WithingsTokenGet(db, user)
	if exists {
		// replace the existing token
		log.Print("Updating withings token for user: ", user.UserID)
		_, err := db.Exec("UPDATE withingsTokens SET token=? WHERE userId=?", tokenStr, user.UserID)

		if err != nil {
			log.Fatal("Failed to update token value: ", err)
		}
	} else {
		// insert a new token record
		log.Print("Saving new withings token for user: ", user.UserID)
		_, err = db.Exec("INSERT INTO withingsTokens (userId, token) VALUES (?, ?)", user.UserID, tokenStr)

		if err != nil {
			log.Fatal("Failed to insert token: ", err)
		}
	}
}

// WithingsTokensGetAll retrieves saved withings API tokens
func WithingsTokensGetAll(db *sql.DB) *[]WithingsToken {

	rows, err := db.Query("SELECT userId, token FROM withingsTokens")
	if err != nil {
		log.Fatal("Failed to read all tokens: ", err)
	}
	defer rows.Close()

	tokens := make([]WithingsToken, 0)

	for rows.Next() {

		var userID string
		var tokenStr string

		err = rows.Scan(&userID, &tokenStr)
		if err != nil {
			log.Fatal("Failed to scan row: ", err)
		}

		tokenBuf := []byte(tokenStr)
		var token oauth2.Token

		err = json.Unmarshal(tokenBuf, &token)
		if err != nil {
			log.Fatal("Failed to unmarshal token:", err)
		}

		t := WithingsToken{
			UserID: userID,
			Token:  token,
		}

		tokens = append(tokens, t)
	}

	return &tokens
}
