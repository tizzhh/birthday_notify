package types

import "time"

type BirthdayUserBase struct {
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Birthday  time.Time `json:"birthday"`
}

type BirthdayUserRequest struct {
	BirthdayUserBase
	Password string `json:"password"`
}

type BirthdayUserResponse struct {
	ID int `json:"id"`
	BirthdayUserBase
}

type BirthdayUser struct {
	ID            int             `json:"id" gorm:"primaryKey"`
	Subscriptions []*BirthdayUser `json:"-" gorm:"many2many:user_subscriptions"`
	BirthdayUserRequest
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"access_token"`
}
