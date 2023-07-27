package main

import (
	"errors"
	"net/http"
	"time"
)

type jsonResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type envelope map[string]interface{}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	type credentials struct {
		Username string `json:"email"`
		Password string `json:"password"`
	}

	var creds credentials

	var payload jsonResponse

	err := app.readJSON(w, r, &creds)
	if err != nil {
		app.errorLog.Println(err)
		payload.Error = true
		payload.Message = "Invalid Json supplied, or json missing"
		err = app.writeJSON(w, http.StatusBadRequest, payload)
		return
	}

	//TODO authenticate
	app.infoLog.Println(creds.Username, creds.Password)

	// look up the user by email
	user, err := app.models.User.GetByEmail(creds.Username)
	if err != nil {
		app.errorJSON(w, errors.New("invalid username/password"))
		return
	}

	//validate the user's password
	validPassword, err := user.PasswordMatches(creds.Password)
	if err != nil || !validPassword {
		app.errorJSON(w, errors.New("invalid username/password"))
		return
	}

	//we have a valid user we generate token
	token, err := app.models.Token.GenerateToken(user.ID, 24*time.Hour)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	//save it to token
	err = app.models.Token.Insert(*token, *user)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	//send back response
	payload = jsonResponse{
		Error:   false,
		Message: "Login successful",
		Data:    envelope{"token": token, "user": user},
	}

	err = app.writeJSON(w, http.StatusOK, payload)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

}

func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Token string `json:"token"`
	}
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, errors.New("invalid token"))
		return
	}

	err = app.models.Token.DeleteByToken(requestPayload.Token)
	if err != nil {
		app.errorJSON(w, errors.New("invalid token"))
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Logged Out",
	}
	_ = app.writeJSON(w, http.StatusOK, payload)
}
