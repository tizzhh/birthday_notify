package types

import "time"

type User struct {
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Email        string    `json:"email"`
	Birthday     time.Time `json:"birthday"`
	IsSubscribed bool      `json:"isSubscribed"`
}
