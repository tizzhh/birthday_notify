package main

import (
	"fmt"
	"log"
	"net/http"

	"birthday/db"
	"birthday/types"

	"github.com/gorilla/mux"
)

const (
	APP_PORT                   string = ":8000"
	DB_HOST_ENV                string = "DB_HOST"
	DB_USER_ENV                string = "POSTGRES_USER"
	DB_PASS_ENV                string = "POSTGRES_PASSWORD"
	DB_NAME_ENV                string = "POSTGRES_DB"
	DB_PORT_ENV                string = "DB_PORT"
	DB_CONNECTION_URL_TEMPLATE string = "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable"
	ADMIN_USER_FIRST_NAME             = "ADMIN_USER_FIRST_NAME"
	ADMIN_USER_LAST_NAME              = "ADMIN_USER_LAST_NAME"
	ADMIN_USER_EMAIL                  = "ADMIN_USER_EMAIL"
	ADMIN_USER_BIRTHDAY               = "ADMIN_USER_BIRTHDAY"
	ADMIN_USER_PASSWORD               = "ADMIN_USER_PASSWORD"
)

type NotifyApp struct {
	Router       *mux.Router
	dbConnection db.DataBase
}

func (na *NotifyApp) Run(port string) {
	log.Fatal(http.ListenAndServe(port, na.Router))
}

func Initialize() (NotifyApp, error) {
	var na NotifyApp
	var err error
	na.dbConnection, err = db.ConnectToDb(DB_HOST_ENV, DB_USER_ENV, DB_PASS_ENV, DB_NAME_ENV, DB_PORT_ENV, DB_CONNECTION_URL_TEMPLATE)
	if err != nil {
		return NotifyApp{}, fmt.Errorf("failed to connect to a database: %w", err)
	}
	err = na.dbConnection.DB.AutoMigrate(&types.BirthdayUser{})
	if err != nil {
		return NotifyApp{}, err
	}
	na.Router = mux.NewRouter()
	return na, nil
}

func (na *NotifyApp) setupRoutes() {
	na.Router.HandleFunc("/api/users", na.getUsersHandler).Methods("GET")
	na.Router.HandleFunc("/api/users", na.createUsersHandler).Methods("POST")
	na.Router.Handle("/api/users/{id:[0-9]+}", authorizationRequired(http.HandlerFunc(na.getUserHandler))).Methods("PUT", "PATCH")
	na.Router.HandleFunc("/api/users/{id:[0-9]+}", na.getUserHandler).Methods("GET")
	na.Router.Handle("/api/users/{id:[0-9]+}/subscribe", authorizationRequired(http.HandlerFunc(na.subscribeToUserHandler))).Methods("POST")
	na.Router.Handle("/api/users/{id:[0-9]+}/unsubscribe", authorizationRequired(http.HandlerFunc(na.unsubscribeFromUserHandler))).Methods("POST")
	na.Router.Handle("/api/birthdays", authorizationRequired(http.HandlerFunc(na.getBirthdaysHandler))).Methods("GET")
	na.Router.Handle("/api/subscriptions", authorizationRequired(http.HandlerFunc(na.getSubscriptionsHandler))).Methods("GET")
	na.Router.HandleFunc("/api/auth/token", na.getTokenhandler).Methods("POST")
	na.Router.HandleFunc("/api/liveness", livenessCheckHandler).Methods("GET")
}

func main() {
	notifyApp, err := Initialize()
	if err != nil {
		log.Fatalf("Error during app initialization: %v\n", err)
	}
	notifyApp.setupRoutes()
	notifyApp.Run(APP_PORT)
}
