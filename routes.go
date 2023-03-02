package main

import (
	"myapp/data"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (a *application) routes() *chi.Mux {
	// middleware must come before any routes

	// add routes here
	a.App.Routes.Get("/", a.Handlers.Home)
	a.App.Routes.Get("/go-page", a.Handlers.GoPage)
	a.App.Routes.Get("/jet-page", a.Handlers.JetPage)
	a.App.Routes.Get("/sessions", a.Handlers.SessionTest)
	a.App.Routes.Get("/create-user", func(w http.ResponseWriter, r *http.Request) {
		// Create a new user
		user := data.User{
			FirstName: "Daota",
			LastName:  "Nguyen",
			Email:     "asd@asd.se",
			Active:    1,
			Password:  "password",
		}

		// Insert the user into the database
		id, err := a.Models.Users.Create(user)
		if err != nil {
			a.App.ErrorLog.Println("error inserting user:", err)
		}

		a.App.InfoLog.Println("user inserted with id:", id)
	})
	a.App.Routes.Get("/all-users", func(w http.ResponseWriter, r *http.Request) {
		users, err := a.Models.Users.All()
		if err != nil {
			a.App.ErrorLog.Println("error getting all users:", err)
		}

		for _, x := range users {
			a.App.InfoLog.Println("user:", x.FirstName, x.LastName, x.Email)
		}
	})
	a.App.Routes.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		user_id, err := strconv.Atoi(id)
		if err != nil {
			a.App.ErrorLog.Println("error converting id:", err)
		}
		user, err := a.Models.Users.Find(user_id)
		if err != nil {
			a.App.ErrorLog.Println("error getting user:", err)
		}

		a.App.InfoLog.Println("user:", user.FirstName, user.LastName, user.Email)
	})
	a.App.Routes.Get("/update-user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		user_id, err := strconv.Atoi(id)
		if err != nil {
			a.App.ErrorLog.Println("error converting id:", err)
		}
		oldUser, err := a.Models.Users.Find(user_id)
		if err != nil {
			a.App.ErrorLog.Println("error getting user:", err)
		}

		oldUser.FirstName = "Daota"
		oldUser.LastName = "Nguyen"
		oldUser.Email = "new@email.com"

		user, err := a.Models.Users.Update(*oldUser)
		if err != nil {
			a.App.ErrorLog.Println("error updating user:", err)
		}

		a.App.InfoLog.Println("user updated:", user.FirstName, user.LastName, user.Email)

	})

	a.App.Routes.Get("/jet", func(w http.ResponseWriter, r *http.Request) {
		a.App.Render.JetPage(w, r, "testjet", nil, nil)
	})

	// static routes
	fileServer := http.FileServer(http.Dir("./public"))
	a.App.Routes.Handle("/public/*", http.StripPrefix("/public", fileServer))

	return a.App.Routes
}
