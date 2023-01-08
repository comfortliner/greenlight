package main

import (
	"net/http"
)

type homeForm struct {
	FieldErrors map[string]string
}

type loginForm struct {
	Email       string
	Password    string
	FieldErrors map[string]string
	RedirectURL string
}

type signupForm struct {
	Name        string
	Email       string
	Password    string
	Password2   string
	FieldErrors map[string]string
	RedirectURL string
}

type tokenVerificationHandlerForm struct {
	Token       string
	FieldErrors map[string]string
	RedirectURL string
}

func (app *application) homeTmplHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = homeForm{}

	app.render(w, r, http.StatusOK, "home.tmpl.html", data)
}

func (app *application) loginTmplHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = loginForm{}

	app.render(w, r, http.StatusOK, "login.tmpl.html", data)
}

func (app *application) signupTmplHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = signupForm{}

	app.render(w, r, http.StatusOK, "signup.tmpl.html", data)
}

func (app *application) tokenVerificationHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = tokenVerificationHandlerForm{}

	app.render(w, r, http.StatusOK, "tokenverification.tmpl.html", data)
}
