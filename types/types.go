package types

import "time"

type BirthdayUserRequest struct {
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Email        string    `json:"email"`
	Birthday     time.Time `json:"birthday"`
	IsSubscribed bool      `json:"isSubscribed"`
}

type BirthdayUser struct {
	ID uint `json:"id" gorm:"primaryKey"`
	BirthdayUserRequest
}
