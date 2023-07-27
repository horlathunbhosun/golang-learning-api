package main

import (
	"net/http"
	"time"
	"vue-api/internal/data"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	//mux.Get("/users/login", app.Login)
	mux.Post("/users/login", app.Login)
	mux.Post("/users/logout", app.Logout)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.AuthTokenMiddleware)

		mux.Post("/foo", func(w http.ResponseWriter, r *http.Request) {
			payload := jsonResponse{
				Error:   false,
				Message: "Bar",
			}
			err := app.writeJSON(w, http.StatusOK, payload)
			if err != nil {
				return
			}
		})
	})

	mux.Get("/users/all", func(w http.ResponseWriter, r *http.Request) {
		var users data.User
		all, err := users.GetAll()
		if err != nil {
			app.errorLog.Println(err)
			return
		}
		payload := jsonResponse{
			Error:   false,
			Message: "success",
			Data:    envelope{"users": all},
		}
		app.writeJSON(w, http.StatusOK, payload)
	})

	mux.Get("/users/add", func(w http.ResponseWriter, r *http.Request) {
		var u = data.User{
			Email:     "test@example.com",
			FirstName: "ME",
			LastName:  "There",
			Password:  "password",
		}

		app.infoLog.Println("Adding user")

		id, err := app.models.User.Insert(u)
		if err != nil {
			app.errorLog.Println(err)
			app.errorJSON(w, err, http.StatusForbidden)
			return
		}

		app.infoLog.Println("Got back Id of", id)
		newUser, _ := app.models.User.GetOne(id)
		err = app.writeJSON(w, http.StatusOK, newUser)
		if err != nil {
			return
		}
	})

	mux.Get("/test-generate-token", func(writer http.ResponseWriter, request *http.Request) {
		token, err := app.models.User.Token.GenerateToken(1, 60*time.Minute)

		if err != nil {
			app.errorLog.Println(err)
			return
		}

		token.Email = "admin@example.com"
		token.CreatedAt = time.Now()
		token.UpdatedAt = time.Now()

		payload := jsonResponse{
			Error:   false,
			Message: "success",
			Data:    token,
		}

		err = app.writeJSON(writer, http.StatusOK, payload)
		if err != nil {
			return
		}
	})

	mux.Get("/test-save-token", func(writer http.ResponseWriter, request *http.Request) {
		token, err := app.models.User.Token.GenerateToken(2, 60*time.Minute)

		if err != nil {
			app.errorLog.Println(err)
			return
		}

		user, err := app.models.User.GetOne(2)
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		token.UserID = user.ID
		token.Email = user.Email
		token.CreatedAt = time.Now()
		token.UpdatedAt = time.Now()

		err = token.Insert(*token, *user)
		if err != nil {
			app.errorLog.Println(err)
			return
		}
		payload := jsonResponse{
			Error:   false,
			Message: "success",
			Data:    token,
		}

		err = app.writeJSON(writer, http.StatusOK, payload)
		if err != nil {
			return
		}
	})

	mux.Get("/test-validate-token", func(writer http.ResponseWriter, request *http.Request) {
		tokenToValidate := request.URL.Query().Get("token")
		valid, err := app.models.Token.ValidToken(tokenToValidate)
		if err != nil {
			app.errorJSON(writer, err)
			return
		}
		var payload jsonResponse
		payload.Error = false
		payload.Data = valid

		app.writeJSON(writer, http.StatusOK, payload)

	})

	return mux
}
