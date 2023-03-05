package middleware

import (
	"myapp/data"

	"github.com/jimmitjoo/gemquick"
)

type Middleware struct {
	App    *gemquick.Gemquick
	Models data.Models
}
