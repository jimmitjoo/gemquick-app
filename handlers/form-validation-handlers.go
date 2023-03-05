package handlers

import (
	"fmt"
	"myapp/data"
	"net/http"

	"github.com/CloudyKit/jet/v6"
)

func (h *Handlers) Form(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)
	validator := h.App.Validator(nil)
	vars.Set("validator", validator)
	vars.Set("user", data.User{})

	err := h.App.Render.Page(w, r, "form", vars, nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering form:", err)
	}
}

func (h *Handlers) PostForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		h.App.ErrorLog.Println("error parsing form:", err)
	}

	validator := h.App.Validator(nil)

	validator.Required(r, "first_name", "last_name", "email")
	validator.IsEmail("email", r.Form.Get("email"))

	validator.Check(len(r.Form.Get("first_name")) > 1, "first_name", "First name must be at least 2 characters long")
	validator.Check(len(r.Form.Get("last_name")) > 1, "last_name", "First name must be at least 2 characters long")

	if !validator.Valid() {
		vars := make(jet.VarMap)
		vars.Set("validator", validator)
		var user data.User
		user.FirstName = r.Form.Get("first_name")
		user.LastName = r.Form.Get("last_name")
		user.Email = r.Form.Get("email")
		vars.Set("user", user)

		err := h.App.Render.Page(w, r, "form", vars, nil)
		if err != nil {
			h.App.ErrorLog.Println("error rendering form:", err)
			return
		}
		return
	}

	fmt.Fprint(w, "Form submitted successfully")
}
