package handlers

import (
	"net/http"

	"github.com/justinas/nosurf"
)

func (h *Handlers) ShowCachePage(w http.ResponseWriter, r *http.Request) {
	err := h.render(w, r, "cache", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) SaveInCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		CSRF  string `json:"csrf_token"`
	}

	err := h.App.ReadJson(w, r, &userInput)
	if err != nil {
		h.App.Error500(w, r)
		return
	}

	// Verify CSRF token
	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.Error500(w, r)
		return
	}

	err = h.App.Cache.Set(userInput.Name, userInput.Value)
	if err != nil {
		h.App.Error500(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "Successfully saved in cache"

	_ = h.App.WriteJson(w, http.StatusCreated, resp)

}

func (h *Handlers) GetFromCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Name string `json:"name"`
		CSRF string `json:"csrf_token"`
	}

	var msg string
	var inCache = true

	err := h.App.ReadJson(w, r, &userInput)
	if err != nil {
		h.App.Error500(w, r)
		return
	}

	// Verify CSRF token
	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.Error500(w, r)
		return
	}

	value, err := h.App.Cache.Get(userInput.Name)
	if err != nil {
		msg = "Failed to retrieve from cache"
		inCache = false
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Value   string `json:"value"`
	}

	if inCache {
		msg = "Successfully retrieved from cache"
		resp.Error = false
		resp.Message = msg
		resp.Value = value.(string)
	} else {
		resp.Error = true
		resp.Message = msg
	}

	_ = h.App.WriteJson(w, http.StatusOK, resp)
}

func (h *Handlers) DeleteFromCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Name string `json:"name"`
		CSRF string `json:"csrf_token"`
	}

	err := h.App.ReadJson(w, r, &userInput)

	if err != nil {
		h.App.Error500(w, r)
		return
	}

	// Verify CSRF token
	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.Error500(w, r)
		return
	}

	err = h.App.Cache.Forget(userInput.Name)

	if err != nil {
		h.App.Error500(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "Successfully deleted from cache"

	_ = h.App.WriteJson(w, http.StatusOK, resp)
}

func (h *Handlers) EmptyCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		CSRF string `json:"csrf_token"`
	}

	err := h.App.ReadJson(w, r, &userInput)
	if err != nil {
		h.App.Error500(w, r)
		return
	}

	// Verify CSRF token
	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.Error500(w, r)
		return
	}

	err = h.App.Cache.Flush()
	if err != nil {
		h.App.Error500(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "Successfully emptied cache"

	_ = h.App.WriteJson(w, http.StatusOK, resp)
}
