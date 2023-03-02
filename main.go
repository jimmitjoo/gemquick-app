package main

import (
	"myapp/data"
	"myapp/handlers"

	"github.com/jimmitjoo/gemquick"
)

type application struct {
	App      *gemquick.Gemquick
	Handlers *handlers.Handlers
	Models   data.Models
}

func main() {
	g := initApplication()
	g.App.ListenAndServe()
}
