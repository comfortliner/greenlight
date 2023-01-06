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
	router.HandlerFunc(http.MethodGet, "/", app.homeHandler)
	router.HandlerFunc(http.MethodGet, "/user/login", app.loginHandler)
	router.HandlerFunc(http.MethodGet, "/user/signup", app.signupHandler)
	router.HandlerFunc(http.MethodGet, "/user/tokenverification", app.tokenVerificationHandler)

	// ==========================================================================================================
	// BACKEND
	// ==========================================================================================================

	// healthcheck
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// user
	router.HandlerFunc(http.MethodPost, "/v1/user/signup", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/user/activate", app.activateUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/user/activate", app.activateUserHandler)

	// tokens
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

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
