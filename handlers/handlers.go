package handlers

import (
	"net/http"

	"github.com/CloudyKit/jet/v6"
	"github.com/jimmitjoo/gemquick"
)

type Handlers struct {
	App *gemquick.Gemquick
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	err := h.App.Render.Page(w, r, "home", nil, nil)

	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) GoPage(w http.ResponseWriter, r *http.Request) {
	err := h.App.Render.GoPage(w, r, "home", nil)

	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) JetPage(w http.ResponseWriter, r *http.Request) {
	err := h.App.Render.JetPage(w, r, "jet-template", nil, nil)

	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) SessionTest(w http.ResponseWriter, r *http.Request) {
	myData := "daota"

	h.App.Session.Put(r.Context(), "data", myData)

	myValue := h.App.Session.GetString(r.Context(), "data")

	vars := make(jet.VarMap)
	vars.Set("data", myValue)

	err := h.App.Render.JetPage(w, r, "sessions", vars, nil)

	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}
