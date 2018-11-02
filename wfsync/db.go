package wfsync

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	"log"
	"os"
	"path/filepath"
)

func DBInit() *sql.DB {

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

	// create nokia token table:
	_, err = db.Exec(
		// token is a json-encoded oauth2.Token
		`CREATE TABLE IF NOT EXISTS nokiaTokens
					(id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
					 userId TEXT NOT NULL,
					 token TEXT NOT NULL, 
					 FOREIGN KEY(userId) REFERENCES users(userId))`)

	if err != nil {
		log.Fatal(err)
	}

	return db
}

type User struct {
	userId string
	userName string

}

func DBUserGet(db *sql.DB, userId string) (User, bool) {
	rows, err := db.Query("SELECT * FROM users where userId=?", userId)
	if err != nil {
		log.Fatal("Failed to query for user: ", err)
	}
	defer rows.Close()

	exists := rows.Next()

	user := User{}

	if exists {
		err = rows.Scan(&user.userId, &user.userName)
		if err != nil {
			log.Fatal("Failed to read row: ", err)
		}
	}

	return user, exists
}

func DBUserSave(db *sql.DB, userId string, userName string) {

	_, exists := DBUserGet(db, userId)

	if exists {
		log.Printf("User %s already exists.", userId)
		return
	}

	log.Printf("Saving user %s", userId)
	_, err := db.Exec("INSERT INTO users (userid, username) VALUES (?, ?)", userId, userName)
	if err != nil {
		log.Fatal("Failed to insert user: ", err)
	}
}

// retrieve a nokia token, if one was previously saved
func DBNokiaTokenGet(db *sql.DB, user User) (*oauth2.Token, bool) {

	rows, err := db.Query("SELECT token FROM nokiaTokens where userId=?", user.userId)

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

// save the nokia token in the user record
func DBNokiaTokenSave(db *sql.DB, user User, token *oauth2.Token) {

	buf, err := json.Marshal(token)
	if err != nil {
		log.Fatal("Failed to marshal token: ", err)
	}
	tokenStr := string(buf)

	_, exists := DBNokiaTokenGet(db, user)
	if exists {
		// replace the existing token
		log.Print("Updating nokia token for user: ", user.userId)
		_, err := db.Exec("UPDATE nokiaTokens SET token=? WHERE userId=?", tokenStr, user.userId)

		if err != nil {
			log.Fatal("Failed to update token value: ", err)
		}
	} else {
		// insert a new token record
		log.Print("Saving new nokia token for user: ", user.userId)
		_, err = db.Exec("INSERT INTO nokiaTokens (userId, token) VALUES (?, ?)", user.userId, tokenStr)

		if err != nil {
			log.Fatal("Failed to insert token: ", err)
		}
	}
}

func DBNokiaTokensGetAll(db *sql.DB, tokenCallback func(*oauth2.Token)) {

	rows, err := db.Query("SELECT userId, token FROM nokiaTokens")
	if err != nil {
		log.Fatal("Failed to read all tokens: ", err)
	}
	defer rows.Close()

	for rows.Next() {

		var userId string
		var tokenStr string

		err = rows.Scan(&userId, &tokenStr)
		if err != nil {
			log.Fatal("Failed to scan row: ", err)
		}

		tokenBuf := []byte(tokenStr)
		var token oauth2.Token

		err = json.Unmarshal(tokenBuf, &token)
		if err != nil {
			log.Fatal("Failed to unmarshal token:", err)
		}

		tokenCallback(&token)
	}
}


