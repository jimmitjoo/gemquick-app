package main

import (
	"log"
	"myapp/data"
	"myapp/handlers"
	"myapp/middleware"
	"os"

	"github.com/jimmitjoo/gemquick"
)

func initApplication() *application {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// init Gemquick
	gem := &gemquick.Gemquick{}
	err = gem.New(path)

	if err != nil {
		log.Fatal(err)
	}

	gem.AppName = "myapp"

	myMiddleware := &middleware.Middleware{
		App: gem,
	}

	myHandlers := &handlers.Handlers{
		App: gem,
	}

	app := &application{
		App:        gem,
		Handlers:   myHandlers,
		Middleware: myMiddleware,
	}

	app.App.Routes = app.routes()

	app.Models = data.New(app.App.DB.Pool)
	myHandlers.Models = app.Models
	app.Middleware.Models = app.Models

	return app
}
