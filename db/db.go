package db

import (
	"fmt"
	"os"

	"birthday/types"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	DB_USER_ENV                string = "DB_USER"
	DB_PASS_ENV                string = "DB_PASS"
	DB_HOST_ENV                string = "DB_HOST"
	DB_NAME_ENV                string = "DB_NAME"
	DB_PORT_ENV                string = "DB_PORT"
	DB_CONNECTION_URL_TEMPLATE string = "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable"
)

type DataBase struct {
	DB *gorm.DB
}

func ConnectToDb(dbHost, dbUser, dbPass, dbName, dbPort, connectionUrl string) (DataBase, error) {
	err := godotenv.Load()
	if err != nil {
		return DataBase{}, fmt.Errorf("error loading .env file: %w", err)
	}
	dbConnectionUrl := fmt.Sprintf(connectionUrl, os.Getenv(dbHost), os.Getenv(dbUser), os.Getenv(dbPass), os.Getenv(dbName), os.Getenv(dbPort))
	db, err := gorm.Open(postgres.Open(dbConnectionUrl), &gorm.Config{})
	if err != nil {
		return DataBase{}, fmt.Errorf("error openning a database connection: %w", err)
	}
	return DataBase{DB: db}, nil
}

func (db DataBase) GetUsers() ([]types.User, error) {

	return nil, nil
}

func (db DataBase) GetUser(id int) (types.User, error) {

	return types.User{}, nil
}

func (db DataBase) SubscribeToUser(id int) error {

	return nil
}

func (db DataBase) GetBirthdays() ([]types.User, error) {

	return nil, nil
}
