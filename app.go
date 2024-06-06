package main

import (
	"birthday/types"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"net/http"
	"regexp"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userKey int

const (
	_ userKey = iota
	claimsKey
)

const (
	emailRegex string = `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
)

var jwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type ErrorResponse struct {
	Err string `json:"error"`
}

func authorizationRequired(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			respondWithError(w, http.StatusUnauthorized, errors.New("missing auth token"))
			return
		}
		tokenString = tokenString[len("Bearer "):]
		claims, err := verifyToken(tokenString)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err)
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, claims)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
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
	if user.Password == "" {
		validationErrors = append(validationErrors, "password field is required")
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
	users, err := na.dbConnection.GetUsers(r)
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

	birthdayUser := types.BirthdayUser{BirthdayUserRequest: user}
	createdUser, err := na.dbConnection.CreateUser(birthdayUser)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, createdUser)
}

func (na *NotifyApp) putPatchUserHandle(w http.ResponseWriter, r *http.Request, user types.BirthdayUserResponse, id int) {
	var updatedUser types.BirthdayUserRequest
	err := json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	if r.Method == http.MethodPut {
		err = validateBirthdayUser(updatedUser)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err)
			return
		}
	}

	claims, ok := r.Context().Value(claimsKey).(jwt.MapClaims)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, errors.New("missing claims in r.Context. The token may be outdated"))
		return
	}
	subject, ok := claims["sub"]
	if !ok {
		respondWithError(w, http.StatusInternalServerError, errors.New("missing subject in JWT map claims. The token may be outdated"))
		return
	}
	idFromSubject, ok := subject.(float64)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, errors.New("subject is not a number"))
		return
	}
	userId := int(idFromSubject)

	if userId != user.ID {
		respondWithError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	if r.Method == http.MethodPut {
		updatedUser, err := na.dbConnection.UpdateUser(id, updatedUser)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err)
			return
		}
		respondWithJSON(w, http.StatusCreated, updatedUser)
		return
	} else if r.Method == http.MethodPatch {
		patchedUser, err := na.dbConnection.PatchUser(id, updatedUser)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err)
			return
		}
		respondWithJSON(w, http.StatusOK, patchedUser)
		return
	}
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
	if r.Method == http.MethodPut || r.Method == http.MethodPatch {
		na.putPatchUserHandle(w, r, user, id)
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}

func subscribeUnsubscribeBase(w http.ResponseWriter, r *http.Request, vars map[string]string) (int, int, error) {
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, errors.New("invalid user id"))
	}

	claims, ok := r.Context().Value(claimsKey).(jwt.MapClaims)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, errors.New("missing claims in r.Context"))
		return 0, 0, errors.New("")
	}
	subject, ok := claims["sub"]
	if !ok {
		respondWithError(w, http.StatusInternalServerError, errors.New("missing subject in JWT map claims"))
		return 0, 0, errors.New("")
	}
	idFromSubject, ok := subject.(float64)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, errors.New("subject is not a number"))
		return 0, 0, errors.New("")
	}
	userId := int(idFromSubject)

	if userId == id {
		respondWithError(w, http.StatusBadRequest, errors.New("cannot subscribe to oneself"))
		return 0, 0, errors.New("")
	}

	return userId, id, nil
}

func (na *NotifyApp) subscribeToUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, id, err := subscribeUnsubscribeBase(w, r, vars)
	if err != nil {
		return
	}

	err = na.dbConnection.SubscribeToUser(userId, id)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			respondWithError(w, http.StatusNotFound, err)
		case err.Error() == "already subscribed":
			respondWithJSON(w, http.StatusOK, "already subscribed to user's birthday with id "+vars["id"])
		default:
			respondWithError(w, http.StatusInternalServerError, err)
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, "subscribed to user's birthday with id "+vars["id"])
}

func (na *NotifyApp) unsubscribeFromUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, id, err := subscribeUnsubscribeBase(w, r, vars)
	if err != nil {
		return
	}

	err = na.dbConnection.UnSubscribeFromUser(userId, id)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			respondWithError(w, http.StatusNotFound, err)
		default:
			respondWithError(w, http.StatusInternalServerError, err)
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, "unsubscribed from user's birthday with id "+vars["id"])
}

func birthdaysSubscriptionsBase(w http.ResponseWriter, r *http.Request) (int, error) {
	claims, ok := r.Context().Value(claimsKey).(jwt.MapClaims)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, errors.New("missing claims in r.Context"))
		return 0, errors.New("")
	}
	subject, ok := claims["sub"]
	if !ok {
		respondWithError(w, http.StatusInternalServerError, errors.New("missing subject in JWT map claims"))
		return 0, errors.New("")
	}
	idFromSubject, ok := subject.(float64)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, errors.New("subject is not a number"))
		return 0, errors.New("")
	}
	userId := int(idFromSubject)
	return userId, nil
}

func (na *NotifyApp) getBirthdaysHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := birthdaysSubscriptionsBase(w, r)
	if err != nil {
		return
	}
	users, err := na.dbConnection.GetBirthdays(userId, r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}

func (na *NotifyApp) getSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := birthdaysSubscriptionsBase(w, r)
	if err != nil {
		return
	}
	users, err := na.dbConnection.GetSubscriptions(userId, r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}

func (na *NotifyApp) getTokenhandler(w http.ResponseWriter, r *http.Request) {
	var loginData types.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	user, err := na.dbConnection.GetUserByEmail(loginData.Email)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, errors.New("incorrect password"))
		return
	}

	payload := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	t, err := token.SignedString(jwtSecretKey)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.New("JWT token signing"))
		return
	}

	respondWithJSON(w, http.StatusCreated, types.Token{Token: t})
}

func verifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
