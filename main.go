package main

import (
	"myapp/data"
	"myapp/handlers"
	"myapp/middleware"

	"github.com/jimmitjoo/gemquick"
)

type application struct {
	App        *gemquick.Gemquick
	Handlers   *handlers.Handlers
	Models     data.Models
	Middleware *middleware.Middleware
}

func main() {
	g := initApplication()
	g.App.ListenAndServe()
}
