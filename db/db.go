package db

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

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
	err := db.DB.Model(&user).Scopes(Paginate(r)).Order("id ASC").Find(&usersResponse).Error
	if err != nil {
		return nil, err
	}
	return usersResponse, nil
}

func (db DataBase) CreateUser(user types.BirthdayUser) (types.BirthdayUserResponse, error) {
	var userCheck types.BirthdayUser
	emailCheck := db.DB.Where("email = ?", user.Email).First(&userCheck)
	if emailCheck.RowsAffected > 0 {
		return types.BirthdayUserResponse{}, errors.New("user with this email already exists")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	if err != nil {
		return types.BirthdayUserResponse{}, err
	}
	user.Password = string(hashedPassword)
	err = db.DB.Create(&user).Error
	if err != nil {
		return types.BirthdayUserResponse{}, err
	}
	return types.BirthdayUserResponse{
		ID: user.ID,
		BirthdayUserBase: types.BirthdayUserBase{
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Birthday:  user.Birthday,
		},
	}, nil
}

func (db DataBase) GetUserByEmail(email string) (types.BirthdayUser, error) {
	var userCheck types.BirthdayUser
	err := db.DB.Where("email = ?", email).First(&userCheck).Error
	if err != nil {
		return types.BirthdayUser{}, err
	}
	return userCheck, nil
}

func (db DataBase) GetUser(id int) (types.BirthdayUserResponse, error) {
	var user types.BirthdayUser
	var userResponse types.BirthdayUserResponse
	err := db.DB.Model(&user).First(&userResponse, id).Error
	if err != nil {
		return types.BirthdayUserResponse{}, err
	}
	return userResponse, nil
}

func (db DataBase) SubscribeToUser(userThatSubscibesId, userToSubscribeid int) error {
	var userThatSubscribes types.BirthdayUser
	var subscriptions []types.BirthdayUser

	err := db.DB.First(&userThatSubscribes, userThatSubscibesId).Error
	if err != nil {
		return err
	}

	err = db.DB.Model(&userThatSubscribes).Where(THROUGH_MANY_TO_MANY_TABLE_SECOND_COLUMN+" = ?", userToSubscribeid).Association(MANY_TO_MANY_FIELD).Find(&subscriptions)
	if err != nil {
		return err
	}
	if len(subscriptions) > 0 {
		return errors.New("already subscribed")
	}

	var userToSubscribe types.BirthdayUser
	err = db.DB.First(&userToSubscribe, userToSubscribeid).Error
	if err != nil {
		return err
	}

	err = db.DB.Model(&userThatSubscribes).Association(MANY_TO_MANY_FIELD).Append(&userToSubscribe)
	if err != nil {
		return err
	}

	return nil
}

func (db DataBase) UnSubscribeFromUser(userThatSubscibesId, userToSubscribeid int) error {
	var userThatSubscribes types.BirthdayUser

	err := db.DB.First(&userThatSubscribes, userThatSubscibesId).Error
	if err != nil {
		return err
	}

	var userToSubscribe types.BirthdayUser
	err = db.DB.First(&userToSubscribe, userToSubscribeid).Error
	if err != nil {
		return err
	}

	err = db.DB.Model(&userThatSubscribes).Association(MANY_TO_MANY_FIELD).Delete(userToSubscribe, userThatSubscribes)
	if err != nil {
		return err
	}

	return nil
}

func (db DataBase) GetBirthdays(userThatSubscibesId int, r *http.Request) ([]types.BirthdayUserResponse, error) {
	var userThatSubscribes types.BirthdayUser
	var subscriptions []types.BirthdayUserResponse

	err := db.DB.First(&userThatSubscribes, userThatSubscibesId).Error
	if err != nil {
		return nil, err
	}

	currentTime := time.Now()
	curentMonth := currentTime.Month()
	currentDay := currentTime.Day()

	err = db.DB.Model(&userThatSubscribes).Scopes(Paginate(r)).Where("EXTRACT(MONTH FROM birthday) = ? AND EXTRACT(DAY FROM birthday) = ?", curentMonth, currentDay).Association(MANY_TO_MANY_FIELD).Find(&subscriptions)
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (db DataBase) GetSubscriptions(userThatSubscibesId int, r *http.Request) ([]types.BirthdayUserResponse, error) {
	var userThatSubscribes types.BirthdayUser
	var subscriptions []types.BirthdayUserResponse

	err := db.DB.First(&userThatSubscribes, userThatSubscibesId).Error
	if err != nil {
		return nil, err
	}

	err = db.DB.Model(&userThatSubscribes).Scopes(Paginate(r)).Association(MANY_TO_MANY_FIELD).Find(&subscriptions)
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (db DataBase) UpdateUser(id int, newUser types.BirthdayUserRequest) (types.BirthdayUserResponse, error) {
	var oldUser types.BirthdayUser
	err := db.DB.First(&oldUser, id).Error
	if err != nil {
		return types.BirthdayUserResponse{}, err
	}
	oldUser.FirstName = newUser.FirstName
	oldUser.LastName = newUser.LastName
	oldUser.Email = newUser.Email
	oldUser.Birthday = newUser.Birthday
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 14)
	if err != nil {
		return types.BirthdayUserResponse{}, err
	}
	oldUser.Password = string(hashedPassword)
	err = db.DB.Save(&oldUser).Error
	if err != nil {
		return types.BirthdayUserResponse{}, err
	}
	return types.BirthdayUserResponse{
		ID: oldUser.ID,
		BirthdayUserBase: types.BirthdayUserBase{
			FirstName: oldUser.FirstName,
			LastName:  oldUser.LastName,
			Email:     oldUser.Email,
			Birthday:  oldUser.Birthday,
		},
	}, nil
}

func (db DataBase) PatchUser(id int, newUser types.BirthdayUserRequest) (types.BirthdayUserResponse, error) {
	var oldUser types.BirthdayUser
	err := db.DB.First(&oldUser, id).Error
	if err != nil {
		return types.BirthdayUserResponse{}, err
	}
	pass := newUser.Password
	var hashedPassword []byte
	if pass != "" {
		hashedPassword, err = bcrypt.GenerateFromPassword([]byte(newUser.Password), 14)
		if err != nil {
			return types.BirthdayUserResponse{}, err
		}
	}
	newUser.Password = string(hashedPassword)
	err = db.DB.Model(&oldUser).Updates(&newUser).Error
	if err != nil {
		return types.BirthdayUserResponse{}, err
	}
	return types.BirthdayUserResponse{
		ID: oldUser.ID,
		BirthdayUserBase: types.BirthdayUserBase{
			FirstName: oldUser.FirstName,
			LastName:  oldUser.LastName,
			Email:     oldUser.Email,
			Birthday:  oldUser.Birthday,
		},
	}, nil
}
