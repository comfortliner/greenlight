package main

import (
	"net/http"
)

type homeForm struct {
	FieldErrors map[string]string
}

type loginForm struct {
	FieldErrors map[string]string
}

type signupForm struct {
	Name        string
	Email       string
	Password    string
	Password2   string
	FieldErrors map[string]string
}

type tokenVerificationHandlerForm struct {
	A_FieldErrors map[string]string
	A_Page        string
	Token         string
}

func (app *application) homeHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = homeForm{}

	app.render(w, r, http.StatusOK, "home.tmpl.html", data)
}

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = loginForm{}

	app.render(w, r, http.StatusOK, "login.tmpl.html", data)
}

func (app *application) signupHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = signupForm{}

	app.render(w, r, http.StatusOK, "signup.tmpl.html", data)
}

func (app *application) tokenVerificationHandler(w http.ResponseWriter, r *http.Request) {
	page := "tokenverification.tmpl.html"

	data := app.newTemplateData(r)
	data.Form = tokenVerificationHandlerForm{
		A_Page: page,
	}

	app.render(w, r, http.StatusOK, page, data)
}
