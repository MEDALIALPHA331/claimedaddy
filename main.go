package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/julienschmidt/httprouter"
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

func (hanlder *UserHandler) HandleGetUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	users, err := hanlder.store.GetAllUsers()

	if err != nil {
		hanlder.logger.Fatalf(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}


func (handler *UserHandler) HandleHome(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte("Hello my friend\n"))
}


func main() {
	logger := log.Default()
	db, err := sql.Open("sqlite3", file)
	// ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)

	if err != nil {
		logger.Fatalf("Error opening the database", err)
	}

	userHandler := NewUserHandler(db, logger)

	router := httprouter.New()

	router.GET("/", userHandler.HandleHome)
    router.GET("/users", userHandler.HandleGetUsers)

    logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", PORT), router))
}

