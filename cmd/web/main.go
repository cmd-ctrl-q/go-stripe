package main

import (
	"html/template"
	"log"
)

const version = "1.0.0"
const cssVersion = "1"

type config struct {
	port uint
	env  string
	api  string
	db   struct {
		dsn string // data source name
	}
	stripe struct {
		secret string
		key    string
	}
}

type application struct {
	config        config
	infoLog       *log.Logger
	errorLog      *log.Logger
	templateCache map[string]*template.Template
	version       string
}

func main() {

}
