package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/comfortliner/greenlight/internal/validator"
	"github.com/go-playground/form/v4"
	"github.com/julienschmidt/httprouter"
)

// Retrieve the "id" URL Parameter from the current request context, then convert ist to an integer and return it.
// If the operation isn't successful, return 0 and an error.
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

type envelope map[string]any

// Define a writeJSON() helper for sending responses. This takes the destination http.ResponseWriter,
// the HTTP status code to send, the data to encode to JSON, and a header map containing any additional HTTP headers
// we want to include in the response.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {

	// Encode the data to JSON, returning an error if there was one.
	// TODO: MarshalIndent() method should only be used in development environment, else use Marshal() method.
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')

	// Loop through the header map and add each header to the http.ResponseWriter header map
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add the correct Content-Type, then write the status code and JSON response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// Define a readJSON() helper for returning clearer, easy-to-action, error messages.
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	// Use http.MaxBytesReader() to limit size of the request body to 1MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Check with DesallowUnknonwFields() method if the JSON from the client includes any field
	// which cannot be mapped to the target destination.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		// TODO: Type maxBytesError added in go1.19
		// var maxBytesError *http.maxBytesError

		switch {
		// Check whether the error has the type *json.SyntaxError.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		// Check if Decode() return an io.ErrUnexpectedEOF error for syntax errors in the JSON.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		// This occur when the JSON value is the wrong type for the target destiation.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d", unmarshalTypeError.Offset)

		// An io.EOF error will be returned by Decode() if the request body is empty.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// Check if the JSON contains a field which cannot be mapped to the target destination
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// TODO: Type maxBytesError added in go1.19
		// Check if the JSON request body exceeded our size limit.
		// case errors.As(err, &maxBytesError):
		// 	return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		// A json.invalidUnmarshalError error will be returned if we pass something unexpected, that is
		// not a non-nil pointer to Decode().
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// For anything else, return the error message as-is.
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// The readString() helper returns a string value from the query string.
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}
	return s
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}
	return i
}

// The background() helper accepts an function as a parameter which will be run as a background go routine.
// It is used to recover any panic.
func (app *application) background(fn func()) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()

		// Recover any panic.
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()

		// Execute the function that we passed as a parameter.
		fn()
	}()
}

// The render() helper method renders the templates from the Template Cache.
func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverErrorResponse(w, r, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

// The newTemplateData() helper is used to define the 'default' data for our templates.
func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		AppName:         app.config.name,
		AppVersion:      app.config.version,
		UserName:        "Gast",
		CurrentYear:     time.Now().Year(),
		IsAuthenticated: app.isAuthenticated(r),
	}
}

// The decodePostForm() helper defines a mapping between HTML form and the destination data fields.
func (app *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		return err
	}

	return nil
}

func (app *application) isAuthenticated(r *http.Request) bool {
	return app.sessionManager.Exists(r.Context(), "authenticatedUserID")
}
