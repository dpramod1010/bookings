package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/dpramod/bookings/internal/config"
	"github.com/dpramod/bookings/internal/driver"
	"github.com/dpramod/bookings/internal/handlers"
	"github.com/dpramod/bookings/internal/helpers"
	"github.com/dpramod/bookings/internal/models"
	"github.com/dpramod/bookings/internal/render"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

// main is the main function
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	defer close(app.MailChan)

	fmt.Println("Starting Mail Listener..")

	listenForMail()

	fmt.Println(fmt.Sprintf("Staring application on port %s", portNumber))

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func run() (*driver.DB, error) {
	// what am I going to put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	//read flags
	// inProduction := flag.Bool("production", true, "Application is in production")
	// useCache := flag.Bool("cache", true, "Use template cache")
	// dbName := flag.String("dbname", "", "Database Name")
	// dbHost := flag.String("dbhost", "localhost", "Database hostr")
	// dbUser := flag.String("dbuser", "", "Database User")
	// dbPass := flag.String("dbpass", "Password123", "Database Password")
	// dbPort := flag.String("dbport", "5232", "Database Port")
	// dbSSL := flag.String("dbssl", "disable", "Database ssl settings (dsable,prefer,require)")
	//host=localhost port=5432 dbname=BookingDatabase user=postgres password=Password123
	flag.Parse()

	// if *dbName == "" || *dbUser == "" {
	// 	fmt.Println("missing required flags")
	// 	os.Exit(1)
	// }

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// set up the session
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// connect to database
	log.Println("Connecting to database...")
	//connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s , sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPass, *dbSSL)
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=BookingDatabase user=postgres password=Password123")
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}

	log.Println("Connected to database!")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
		return nil, err
	}

	app.TemplateCache = tc
	//app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
