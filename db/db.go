package db

import (
	"fmt"
	"os"

	"birthday/types"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func (db DataBase) GetUsers() ([]types.BirthdayUser, error) {
	var users []types.BirthdayUser
	results := db.DB.Find(&users)
	return users, results.Error
}

func (db DataBase) CreateUser(user types.BirthdayUser) error {
	result := db.DB.Create(&user)
	return result.Error
}

func (db DataBase) GetUser(id int) (types.BirthdayUser, error) {
	var user types.BirthdayUser
	result := db.DB.First(&user, id)
	return user, result.Error
}

func (db DataBase) SubscribeToUser(id int) error {
	var user types.BirthdayUser
	result := db.DB.First(&user, id)
	if result.Error != nil {
		return result.Error
	}
	user.IsSubscribed = true
	db.DB.Save(&user)
	return nil
}

func (db DataBase) GetBirthdays() ([]types.BirthdayUser, error) {
	var users []types.BirthdayUser
	results := db.DB.Where("is_subscribed = ?", "1").Find(&users)
	return users, results.Error
}
