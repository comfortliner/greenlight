package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/comfortliner/greenlight/internal/data"
	"github.com/comfortliner/greenlight/internal/mailer"
	"github.com/comfortliner/greenlight/internal/vcs"

	"github.com/alexedwards/scs/mssqlstore"
	"github.com/alexedwards/scs/v2"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/go-playground/form/v4"
)

// Define a config stuct to hold all the configuration settings for our application.
// We will read in these configuration settings from command-line flags when the application starts.
type config struct {
	cors struct {
		trustedOrigins []string
	}
	db struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	env  string
	name string
	port int
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	version string
}

// Define an application struct to hold the dependencies for our HTTP handlers, helpers and middleware.
type application struct {
	config         config
	logger         *slog.Logger
	mailer         mailer.Mailer
	models         data.Models
	wg             sync.WaitGroup
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	// Declare an instance of the config struct.
	var cfg config

	// Application
	cfg.name = "Greenlight"
	cfg.version = vcs.Version()

	// Read the value of the given flags.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "Microsoft SQL Server DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "SQLServer max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "SQLServer max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "SQLServer max connection idle time")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "d272ee6f9eb15f", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "2f762fd7bf31fc", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "API <no-reply@api.net>", "SMTP sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)",
		func(val string) error {
			cfg.cors.trustedOrigins = strings.Fields(val)
			return nil
		})

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", cfg.version)
		os.Exit(0)
	}

	// Initialize a new structured logger which writes log entries to the standard out stream,
	// formatted as a JSON object.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Call the openDB() helper function to create the connection pool.
	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Defer a call to db.Close() method so that the connection pool is closed before
	// the main() function exits.
	defer db.Close()

	logger.Info("database connection pool established")

	// Initialize a new template cache.
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Initialize a decoder instance.
	formDecoder := form.NewDecoder()

	// TODO: sessionManager instance...
	// important for TLS connections: sessionManager.Cookie.Secure = true

	sessionManager := scs.New()
	sessionManager.Store = mssqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	// Declare an instance of the application struct.
	app := &application{
		config:         cfg,
		logger:         logger,
		mailer:         mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		models:         data.NewModels(db),
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// The openDB() function returns a sql.DB connection pool.
func openDB(cfg config) (*sql.DB, error) {

	// Use sql.Open() method to create an empty connection pool.
	db, err := sql.Open("sqlserver", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
