package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"vue-api/internal/data"
	"vue-api/internal/driver"
)

// config is the type for all application configuration
type config struct {
	port int
}

/*
  - application is the type for all data we want to share with the various parts of our application.
    we will share this information in most cases by using this type as the reciever for functions

*
*/
type application struct {
	config      config
	infoLog     *log.Logger
	errorLog    *log.Logger
	models      data.Models
	environment string
}

//main is the main entry point for our application

func main() {
	var cfg config
	cfg.port = 8081

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	dsn := os.Getenv("DSN")
	environment := os.Getenv("ENV")
	//"host=localhost port=5432 user=postgres password=password dbname=vueapi sslmode=disable timezone=UTC connect_timeout=5"

	db, err := driver.ConnectPostgres(dsn)
	if err != nil {
		log.Fatal("Can't Connect to Database")
	}

	fmt.Println(variadicSum(1, 2, 3, 4, 45))

	defer db.SQL.Close()

	app := &application{
		config:      cfg,
		infoLog:     infoLog,
		errorLog:    errorLog,
		models:      data.New(db.SQL),
		environment: environment,
	}
	err = app.serve()

	if err != nil {
		log.Fatal(err)
	}

}

func variadicSum(numbers ...int) (result int) {
	for _, v := range numbers {
		result += v
	}
	return
}

// serve starts the web server
func (app *application) serve() error {
	app.infoLog.Println("API listening on port", app.config.port)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(),
	}
	return srv.ListenAndServe()
}
