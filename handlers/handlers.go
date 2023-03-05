package handlers

import (
	"fmt"
	"myapp/data"
	"net/http"

	"github.com/CloudyKit/jet/v6"
	"github.com/jimmitjoo/gemquick"
)

type Handlers struct {
	App    *gemquick.Gemquick
	Models data.Models
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	err := h.render(w, r, "home", nil, nil)

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

func (h *Handlers) Json(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ID      int64    `json:"id"`
		Name    string   `json:"name"`
		Hobbies []string `json:"hobbies"`
	}

	payload.ID = 10
	payload.Name = "John Doe"
	payload.Hobbies = []string{"gaming", "coding"}

	err := h.App.WriteJson(w, http.StatusOK, payload)
	if err != nil {
		h.App.ErrorLog.Println("error writing json:", err)
	}
}

func (h *Handlers) XML(w http.ResponseWriter, r *http.Request) {
	type Payload struct {
		ID      int64    `xml:"id"`
		Name    string   `xml:"name"`
		Hobbies []string `xml:"hobbies>hobby"`
	}

	var payload Payload
	payload.ID = 10
	payload.Name = "John Doe"
	payload.Hobbies = []string{"gaming", "coding"}

	err := h.App.WriteXML(w, http.StatusOK, payload)
	if err != nil {
		h.App.ErrorLog.Println("error writing json:", err)
	}
}

func (h *Handlers) DownloadFile(w http.ResponseWriter, r *http.Request) {
	h.App.DownloadFile(w, r, "./public/images/", "celeritas.jpg")
}

func (h *Handlers) TestCrypto(w http.ResponseWriter, r *http.Request) {
	plainText := "Hello World"
	fmt.Fprint(w, "Uncrypted: "+plainText+"\n")

	encrypted, err := h.encrypt(plainText)
	if err != nil {
		h.App.ErrorLog.Println("error encrypting:", err)
		h.App.Error500(w, r)
		return
	}

	fmt.Fprint(w, "Encrypted: "+encrypted+"\n")

	decrypted, err := h.decrypt(encrypted)

	if err != nil {
		h.App.ErrorLog.Println("error decrypting:", err)
		h.App.Error500(w, r)
		return
	}

	fmt.Fprint(w, "Decrypted: "+decrypted+"\n")
}
