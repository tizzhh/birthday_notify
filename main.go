package main

import (
	"fmt"
	"log"
	"net/http"

	"birthday/db"
	"github.com/gorilla/mux"
)

const (
	APP_PORT                   string = "8080"
	DB_HOST_ENV                string = "DB_HOST"
	DB_USER_ENV                string = "DB_USER"
	DB_PASS_ENV                string = "DB_PASS"
	DB_NAME_ENV                string = "DB_NAME"
	DB_PORT_ENV                string = "DB_PORT"
	DB_CONNECTION_URL_TEMPLATE string = "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable"
)

type NotifyApp struct {
	Router *mux.Router
	DB     db.DataBase
}

func (na *NotifyApp) Run(port string) {
	log.Fatal(http.ListenAndServe(port, na.Router))
}

func Initialize() (NotifyApp, error) {
	var na NotifyApp
	var err error
	na.DB, err = db.ConnectToDb(DB_HOST_ENV, DB_USER_ENV, DB_PASS_ENV, DB_NAME_ENV, DB_PORT_ENV, DB_CONNECTION_URL_TEMPLATE)
	if err != nil {
		return NotifyApp{}, fmt.Errorf("failed to connect to a database: %w", err)
	}
	na.Router = mux.NewRouter()
	return na, nil
}

func (na *NotifyApp) setupRoutes() {
	na.Router.HandleFunc("/users", na.getUsersHandler).Methods("GET")
	na.Router.HandleFunc("/users/{id:[0-9]+}", na.getUserHandler).Methods("GET")
	na.Router.HandleFunc("/users/{id:[0-9]+}", na.subscribeToUserHandler).Methods("POST")
	na.Router.HandleFunc("/birthdays", na.getBirthdaysHandler).Methods("GET")
}

func main() {
	notifyApp, err := Initialize()
	if err != nil {
		log.Fatalf("Error during app initialization: %v\n", err)
	}
	notifyApp.setupRoutes()
	notifyApp.Run(APP_PORT)
}
