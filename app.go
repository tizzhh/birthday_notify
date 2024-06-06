package main

import (
	"birthday/types"
	"encoding/json"
	"errors"
	"strings"

	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

const (
	emailRegex string = `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
)

type ErrorResponse struct {
	Err string `json:"error"`
}

func validateBirthdayUser(user types.BirthdayUserRequest) error {
	var validationErrors []string
	if user.FirstName == "" {
		validationErrors = append(validationErrors, "firstName field is required")
	}
	if user.LastName == "" {
		validationErrors = append(validationErrors, "lastName field is required")
	}
	if user.Email == "" {
		validationErrors = append(validationErrors, "email field is required")
	}
	if user.Birthday.IsZero() {
		validationErrors = append(validationErrors, "birthday field is required")
	}
	if user.Email != "" {
		re := regexp.MustCompile(emailRegex)
		if !re.MatchString(user.Email) {
			validationErrors = append(validationErrors, "invalid email format")
		}
	}
	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, ", "))
	}
	return nil
}

func respondWithError(w http.ResponseWriter, responseCode int, err error) {
	respondWithJSON(w, responseCode, ErrorResponse{err.Error()})
}

func respondWithJSON(w http.ResponseWriter, reponseCode int, payload any) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(reponseCode)
	w.Write(response)
}

func (na *NotifyApp) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := na.dbConnection.GetUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}

func (na *NotifyApp) createUsersHandler(w http.ResponseWriter, r *http.Request) {
	var user types.BirthdayUserRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	err = validateBirthdayUser(user)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	err = na.dbConnection.CreateUser(types.BirthdayUser{
		BirthdayUserRequest: user,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (na *NotifyApp) getUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, errors.New("invalid user id"))
	}

	user, err := na.dbConnection.GetUser(id)
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			respondWithError(w, http.StatusNotFound, err)
		default:
			respondWithError(w, http.StatusInternalServerError, err)
		}
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (na *NotifyApp) subscribeToUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, errors.New("invalid user id"))
	}

	err = na.dbConnection.SubscribeToUser(id)
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			respondWithError(w, http.StatusNotFound, err)
		default:
			respondWithError(w, http.StatusInternalServerError, err)
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, "Subscribed to user's birthday with id "+vars["id"])
}

func (na *NotifyApp) getBirthdaysHandler(w http.ResponseWriter, r *http.Request) {
	users, err := na.dbConnection.GetBirthdays()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}
