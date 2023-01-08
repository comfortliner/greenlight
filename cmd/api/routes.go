package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// Initialize a new httprouter router instance.
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// ==========================================================================================================
	// FRONTEND
	// ==========================================================================================================

	// serve static files
	router.ServeFiles("/static/*filepath", http.Dir("./ui/static/"))

	// template routes
	router.Handler(http.MethodGet, "/", app.sessionManager.LoadAndSave(http.HandlerFunc(app.homeTmplHandler)))

	// mod_sfta
	// StateFul Token Authentication
	router.Handler(http.MethodGet, "/user/signup", app.sessionManager.LoadAndSave(http.HandlerFunc(app.signupTmplHandler)))
	router.Handler(http.MethodPost, "/user/signup", app.sessionManager.LoadAndSave(http.HandlerFunc(app.signupUserHandler)))

	router.Handler(http.MethodGet, "/user/tokenverification", app.sessionManager.LoadAndSave(http.HandlerFunc(app.tokenVerificationHandler)))
	router.Handler(http.MethodPost, "/user/activate", app.sessionManager.LoadAndSave(http.HandlerFunc(app.activateUserHandler)))

	router.Handler(http.MethodGet, "/user/login", app.sessionManager.LoadAndSave(http.HandlerFunc(app.loginTmplHandler)))
	router.Handler(http.MethodPost, "/user/login", app.sessionManager.LoadAndSave(http.HandlerFunc(app.loginUserHandler)))

	router.Handler(http.MethodPost, "/user/logout", app.sessionManager.LoadAndSave(http.HandlerFunc(app.logoutUserHandler)))

	// ==========================================================================================================
	// BACKEND
	// ==========================================================================================================

	// healthcheck
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// tokens
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// ToDo: Implement Session Manager to the following routes.
	// movies
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission("movies:read", app.listMoviesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("movies:write", app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission("movies:read", app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("movies:write", app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("movies:write", app.deleteMovieHandler))

	// Return the httprouter instance.
	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))

	// TODO Use Composable middleware chains as described in Chapter 6.5 from Alex Edwards book 'Let´s Go'
}
