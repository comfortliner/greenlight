package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/comfortliner/greenlight/internal/data"
	"github.com/comfortliner/greenlight/internal/validator"
)

// Add a signupUserHandler for the "POST /user/signup" endpoint.
func (app *application) signupUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string `json:"name" form:"name"`
		Email     string `json:"email" form:"email"`
		Password  string `json:"password" form:"password"`
		Password2 string `json:"password2" form:"password2"`
	}

	err := app.decodePostForm(r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	form := signupForm{
		Name:        input.Name,
		Email:       input.Email,
		Password:    input.Password,
		Password2:   input.Password2,
		FieldErrors: map[string]string{},
		RedirectURL: "signup.tmpl.html",
	}

	// Copy the values from the input struct to a new User struct.
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = user.Password2.Set(input.Password2)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Use the Valid() method to see if any of the checks failed.
	if data.ValidateUserRegister(v, user); !v.Valid() {
		form.FieldErrors = v.Errors
		app.failedValidationResponseForm(w, r, &form, form.RedirectURL)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			form.FieldErrors = v.Errors
			app.failedValidationResponseForm(w, r, &form, form.RedirectURL)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Add the standard permissions for the new user.
	err = app.models.Permissions.AddForUser(user.ID, "movies:read")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// After the user record has been created in the database, generate a
	// new activation token for the user.
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Use the background helper to execute an anonymous function that sends the welcome email.
	app.background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	http.Redirect(w, r, "/user/tokenverification", http.StatusSeeOther)
}

// Add a activateUserHandler for the "POST /user/activate" endpoint.
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token" form:"token"`
	}

	err := app.decodePostForm(r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	form := tokenVerificationHandlerForm{
		Token:       input.TokenPlaintext,
		FieldErrors: map[string]string{},
		RedirectURL: "tokenverification.tmpl.html",
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Use the Valid() method to see if any of the checks failed.
	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		form.FieldErrors = v.Errors
		app.failedValidationResponseForm(w, r, &form, form.RedirectURL)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			form.FieldErrors = v.Errors
			app.failedValidationResponseForm(w, r, &form, form.RedirectURL)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			v.AddError("token", "unable to update the record due to an edit conflict, please try again")
			form.FieldErrors = v.Errors
			app.failedValidationResponseForm(w, r, &form, form.RedirectURL)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// Add a loginUserHandler for the "POST /user/login" endpoint.
func (app *application) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
	}

	err := app.decodePostForm(r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	form := loginForm{
		Email:       input.Email,
		Password:    input.Password,
		FieldErrors: map[string]string{},
		RedirectURL: "login.tmpl.html",
	}

	// Copy the values from the input struct to a new User struct.
	user := &data.User{
		Email: input.Email,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Use the Valid() method to see if any of the checks failed.
	if data.ValidateUserLogin(v, user); !v.Valid() {
		form.FieldErrors = v.Errors
		app.failedValidationResponseForm(w, r, &form, form.RedirectURL)
		return
	}

	id, err := app.models.Users.Authenticate(user.Email, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidCredentials):
			v.AddError("email", "email or password is incorrect")
			form.FieldErrors = v.Errors
			app.failedValidationResponseForm(w, r, &form, form.RedirectURL)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Add a logoutUserHandler for the "POST /user/logout" endpoint.
func (app *application) logoutUserHandler(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
