package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func respondWithError(w http.ResponseWriter, responseCode int, message string) {
	respondWithJSON(w, responseCode, message)
}

func respondWithJSON(w http.ResponseWriter, reponseCode int, payload any) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(reponseCode)
	w.Write(response)
}

func (na *NotifyApp) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := na.DB.GetUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}

func (na *NotifyApp) getUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user id")
	}

	user, err := na.DB.GetUser(id)
	// check whether the user was found or 404
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (na *NotifyApp) subscribeToUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user id")
	}

	err = na.DB.SubscribeToUser(id)
	// check whether the user was found or 404
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, "Subscribed to user's birthday with id "+vars["id"])
}

func (na *NotifyApp) getBirthdaysHandler(w http.ResponseWriter, r *http.Request) {
	users, err := na.DB.GetBirthdays()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}
