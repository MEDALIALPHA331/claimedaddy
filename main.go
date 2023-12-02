package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

var (
	PORT = os.Getenv("PORT")
)

const file string = "activities.db"

type User struct {
	ID          int    `json:"id,omitempty"`
	Description string `json:"description"`
	Fullname    string `json:"fullname"`
	Email       string `json:"email"`
	Password    string `json:"-"`
}

type UserStore interface {
	GetAllUsers() []User
}

type UsersSqlStore struct {
	// mu sync.Mutex
	db *sql.DB
}

func NewSqlStore(db *sql.DB) *UsersSqlStore {

	pingErr := db.Ping()

	if pingErr != nil {
		log.Fatal(pingErr)
		os.Exit(1)
	}

	fmt.Println("Connected!")

	return &UsersSqlStore{
		// mu: mutex.Lock(),
		db: db,
	}
}

func GetUsersFromStore(logger *log.Logger, db *sql.DB) ([]User, error) {
	var users []User

	rows, err := db.Query("select * from user")

	if err != nil {
		logger.Fatalf("Failed to get users from db, error is: %+v", err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Description, &user.Fullname, &user.Email, &user.Password); err != nil {
			return nil, fmt.Errorf("Failed to scan USER from db error is: %v", err)
		}

		users = append(users, user)
	}


	return users, nil

}

func main() {
	logger := log.Default()

	db, err := sql.Open("sqlite3", file)
	myDatabase := NewSqlStore(db)

	mux := http.NewServeMux()


	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {

		users, err := GetUsersFromStore(logger, myDatabase.db)

		if err != nil {
			logger.Fatalf(err.Error())
		}

		res, err := json.Marshal(users)

		w.WriteHeader(200)
		w.Header().Add("Content-Type", "application/json")
		w.Write(res)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello my friend\n"))
	})

	err = http.ListenAndServe(fmt.Sprintf(":%s", PORT), mux)

	logger.Printf("Server is Listening on port %s", fmt.Sprintf(":%s", PORT))

	if err != nil {
		logger.Fatalf("Failed to run the server and listen on port %+v: error is %+v", PORT, err)
		os.Exit(1)
	}
}

// ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
// defer cancel()

// conn, err := myDatabase.db.Conn(ctx)
// defer conn.Close()

// if err != nil {
// 	logger.Fatalf("Failed to connnect to database, error is: %+v", err)
// 	os.Exit(1)
// }
