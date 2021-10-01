package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cmd-ctrl-q/go-stripe/internal/driver"
	"github.com/cmd-ctrl-q/go-stripe/internal/models"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string // data source name
	}
	stripe struct {
		secret string
		key    string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
	}
	secretkey string // key to sign urls
	frontend  string // address for frontend
}

type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
	DB       models.DBModel
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.infoLog.Printf("Starting Backend server in %s mode on port %d\n", app.config.env, app.config.port)

	return srv.ListenAndServe()
}

func main() {
	var cfg config

	// read flags into config variable
	flag.IntVar(&cfg.port, "port", 4001, "Server port to listen on")
	flag.StringVar(&cfg.env, "env", "development", "Application environment {development|production|maintenance}")
	flag.StringVar(&cfg.db.dsn, "dsn", os.Getenv("MARIADB_USER")+":"+os.Getenv("MARIADB_PASS")+"@tcp(localhost:3306)/widgets?parseTime=true&tls=false", "DSN")
	flag.StringVar(&cfg.smtp.host, "smtphost", "smtp.mailtrap.io", "smtp host")
	flag.StringVar(&cfg.smtp.username, "smtpuser", os.Getenv("SMTP_USER"), "smtp user")
	flag.StringVar(&cfg.smtp.password, "smtppass", os.Getenv("SMTP_PASS"), "smtp password")
	flag.IntVar(&cfg.smtp.port, "smtpport", 587, "smtp port")
	flag.StringVar(&cfg.secretkey, "secret", os.Getenv("SECRET_KEY"), "secret key")
	flag.StringVar(&cfg.frontend, "frontend", "http://localhost:4000", "url to frontend")

	flag.Parse()

	key := os.Getenv("STRIPE_KEY")
	secret := os.Getenv("STRIPE_SECRET")
	if key == "" || secret == "" {
		log.Panic("Stripe key or secret is empty. (Possible fixes: .air.toml, Makefile, local env variables)")
	}

	cfg.stripe.key = key
	cfg.stripe.secret = secret

	// set up loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// create db connection pool
	conn, err := driver.OpenDB(cfg.db.dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer conn.Close()

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
		DB:       models.DBModel{DB: conn},
	}

	err = app.serve()
	if err != nil {
		log.Fatal(err)
	}
}
