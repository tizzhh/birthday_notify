package db

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"birthday/types"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"
)

const (
	MANY_TO_MANY_FIELD                       string = "Subscriptions"
	THROUGH_MANY_TO_MANY_TABLE_SECOND_COLUMN string = "subscription_id"
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

func Paginate(r *http.Request) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		q := r.URL.Query()
		page, _ := strconv.Atoi(q.Get("page"))
		if page <= 0 {
			page = 1
		}

		pageSize, _ := strconv.Atoi(q.Get("page_size"))
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func (db DataBase) GetUsers(r *http.Request) ([]types.BirthdayUserResponse, error) {
	var user types.BirthdayUser
	var usersResponse []types.BirthdayUserResponse
	results := db.DB.Model(&user).Scopes(Paginate(r)).Find(&usersResponse)
	return usersResponse, results.Error
}

func (db DataBase) CreateUser(user types.BirthdayUser) error {
	var userCheck types.BirthdayUser
	emailCheck := db.DB.Where("email = ?", user.Email).First(&userCheck)
	if emailCheck.RowsAffected > 0 {
		return errors.New("user with this email already exists")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	result := db.DB.Create(&user)
	return result.Error
}

func (db DataBase) GetUserByEmail(email string) (types.BirthdayUser, error) {
	var userCheck types.BirthdayUser
	emailCheck := db.DB.Where("email = ?", email).First(&userCheck)
	return userCheck, emailCheck.Error
}

func (db DataBase) GetUser(id int) (types.BirthdayUserResponse, error) {
	var user types.BirthdayUser
	var userResponse types.BirthdayUserResponse
	result := db.DB.Model(&user).First(&userResponse, id)
	return userResponse, result.Error
}

func (db DataBase) SubscribeToUser(userThatSubscibesId, userToSubscribeid int) error {
	var userThatSubscribes types.BirthdayUser
	var subscriptions []types.BirthdayUser

	err := db.DB.First(&userThatSubscribes, userThatSubscibesId).Error
	if err != nil {
		return err
	}

	db.DB.Model(&userThatSubscribes).Where(THROUGH_MANY_TO_MANY_TABLE_SECOND_COLUMN+" = ?", userToSubscribeid).Association(MANY_TO_MANY_FIELD).Find(&subscriptions)
	if len(subscriptions) > 0 {
		return errors.New("already subscribed")
	}

	var userToSubscribe types.BirthdayUser
	err = db.DB.First(&userToSubscribe, userToSubscribeid).Error
	if err != nil {
		return err
	}

	db.DB.Model(&userThatSubscribes).Association(MANY_TO_MANY_FIELD).Append(&userToSubscribe)

	return nil
}

func (db DataBase) GetBirthdays(userThatSubscibesId int, r *http.Request) ([]types.BirthdayUserResponse, error) {
	var userThatSubscribes types.BirthdayUser
	var subscriptions []types.BirthdayUserResponse

	err := db.DB.First(&userThatSubscribes, userThatSubscibesId).Error
	if err != nil {
		return nil, err
	}

	db.DB.Model(&userThatSubscribes).Scopes(Paginate(r)).Association(MANY_TO_MANY_FIELD).Find(&subscriptions)

	return subscriptions, nil
}
