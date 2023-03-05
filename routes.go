package main

import (
	"fmt"
	"myapp/data"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (route *application) routes() *chi.Mux {
	// middleware must come before any routes

	// add routes here
	route.get("/", route.Handlers.Home)
	route.get("/go-page", route.Handlers.GoPage)
	route.get("/jet-page", route.Handlers.JetPage)
	route.get("/sessions", route.Handlers.SessionTest)
	route.get("/login", route.Handlers.UserLogin)
	route.post("/login", route.Handlers.PostUserLogin)
	route.get("/logout", route.Handlers.UserLogout)
	route.get("/form", route.Handlers.Form)
	route.post("/form", route.Handlers.PostForm)

	route.get("/json", route.Handlers.Json)
	route.get("/xml", route.Handlers.XML)
	route.get("/download-file", route.Handlers.DownloadFile)
	route.get("/crypto", route.Handlers.TestCrypto)

	route.get("/cache-test", route.Handlers.ShowCachePage)
	route.post("/api/save-in-cache", route.Handlers.SaveInCache)
	route.post("/api/get-from-cache", route.Handlers.GetFromCache)
	route.post("/api/delete-from-cache", route.Handlers.DeleteFromCache)
	route.post("/api/empty-cache", route.Handlers.EmptyCache)

	route.get("/create-user", func(w http.ResponseWriter, r *http.Request) {
		// Create a new user
		user := data.User{
			FirstName: "Daota",
			LastName:  "Nguyen",
			Email:     "asd@asd.se",
			Active:    1,
			Password:  "password",
		}

		// Insert the user into the database
		id, err := route.Models.Users.Create(user)
		if err != nil {
			route.App.ErrorLog.Println("error inserting user:", err)
		}

		route.App.InfoLog.Println("user inserted with id:", id)
	})
	route.get("/all-users", func(w http.ResponseWriter, r *http.Request) {
		users, err := route.Models.Users.All()
		if err != nil {
			route.App.ErrorLog.Println("error getting all users:", err)
		}

		for _, x := range users {
			route.App.InfoLog.Println("user:", x.FirstName, x.LastName, x.Email)
		}
	})
	route.get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		user_id, err := strconv.Atoi(id)
		if err != nil {
			route.App.ErrorLog.Println("error converting id:", err)
		}
		user, err := route.Models.Users.Find(user_id)
		if err != nil {
			route.App.ErrorLog.Println("error getting user:", err)
		}

		route.App.InfoLog.Println("user:", user.FirstName, user.LastName, user.Email)
	})
	route.get("/update-user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		user_id, err := strconv.Atoi(id)
		if err != nil {
			route.App.ErrorLog.Println("error converting id:", err)
		}
		oldUser, err := route.Models.Users.Find(user_id)
		if err != nil {
			route.App.ErrorLog.Println("error getting user:", err)
		}

		oldUser.FirstName = "Daota"
		oldUser.LastName = ""
		oldUser.Email = ""

		validator := route.App.Validator(nil)
		oldUser.Validate(validator)

		if !validator.Valid() {
			fmt.Fprint(w, validator.Errors)
			return
		}

		user, err := route.Models.Users.Update(*oldUser)
		if err != nil {
			route.App.ErrorLog.Println("error updating user:", err)
		}

		route.App.InfoLog.Println("user updated:", user.FirstName, user.LastName, user.Email)

	})

	route.get("/jet", func(w http.ResponseWriter, r *http.Request) {
		route.App.Render.JetPage(w, r, "testjet", nil, nil)
	})

	// static routes
	fileServer := http.FileServer(http.Dir("./public"))
	route.App.Routes.Handle("/public/*", http.StripPrefix("/public", fileServer))

	return route.App.Routes
}
