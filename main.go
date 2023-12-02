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
	InsertNewUser() (map[string]string, error)
	GetAllUsers() ([]User, error)
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




func (store *UsersSqlStore) GetAllUsers() ([]User, error) {
	var users []User

	rows, err := store.db.Query("select * from user")

	if err != nil {
		log.Fatalf("Failed to get users from db, error is: %+v", err)
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

type UserHandler struct {
	store UsersSqlStore
	logger *log.Logger
}

func NewUserHandler(db  *sql.DB, logger *log.Logger) *UserHandler {
	myDatabase := NewSqlStore(db)

	return &UserHandler{
		store: *myDatabase,
		logger: logger,
	}

}

func (hanlder *UserHandler) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := hanlder.store.GetAllUsers()

	if err != nil {
		hanlder.logger.Fatalf(err.Error())
	}
	res, err := json.Marshal(users)

	if err != nil {
		hanlder.logger.Fatalf("Failed to encode Json: %+v", err.Error())
	}


	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}


func (handler *UserHandler) HandleHome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello my friend\n"))
}


func main() {
	logger := log.Default()
	db, err := sql.Open("sqlite3", file)
	// ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)

	userHandler := NewUserHandler(db, logger)


	mux := http.NewServeMux()

	mux.HandleFunc("/", userHandler.HandleHome)
	mux.HandleFunc("/users", userHandler.HandleGetUsers)

	err = http.ListenAndServe(fmt.Sprintf(":%s", PORT), mux)

	logger.Printf("Server is Listening on port %s", fmt.Sprintf(":%s", PORT))
	if err != nil {
		logger.Fatalf("Failed to run the server and listen on port %+v: error is %+v", PORT, err)
		os.Exit(1)
	}
}

