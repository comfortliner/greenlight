package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/comfortliner/greenlight/internal/data"
	"github.com/comfortliner/greenlight/internal/validator"
)

// Add a registerUserHandler for the "POST /v1/user/signup" endpoint.
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string `json:"name" form:"name"`
		Email     string `json:"email" form:"email"`
		Password  string `json:"password" form:"password"`
		Password2 string `json:"password2" form:"password2"`
	}

	switch app.isForm(r) {
	case true:
		err := app.decodePostForm(r, &input)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
	default:
		err := app.readJSON(w, r, &input)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
	}

	ct := r.Header.Get("Content-type")

	// Copy the values from the input struct to a new User struct.
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err := user.Password.Set(input.Password)
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
	if data.ValidateUser(v, user); !v.Valid() {
		if ct == "application/x-www-form-urlencoded" {
			form := signupForm{
				Name:        input.Name,
				Email:       input.Email,
				Password:    input.Password,
				Password2:   input.Password2,
				FieldErrors: map[string]string{},
			}

			form.FieldErrors = v.Errors

			data := app.newTemplateData(r)
			data.Form = form

			app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
			return
		}

		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
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

	if ct == "application/x-www-form-urlencoded" {
		http.Redirect(w, r, "/user/tokenverification", http.StatusSeeOther)
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Add a activateUserHandler for the ("POST /v1/user/activate" = HTML Forms)
// and the ("PUT /v1/user/activate" = JSON) endpoint.
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token" form:"token"`
	}

	switch app.isForm(r) {
	case true:
		err := app.decodePostForm(r, &input)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
	default:
		err := app.readJSON(w, r, &input)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Use the Valid() method to see if any of the checks failed.
	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
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
			app.failedValidationResponse(w, r, v.Errors)
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

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
