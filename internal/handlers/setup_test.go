package handlers

import (
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sarkartanmay393/RoomReservation-WebApp/internal/config"
	"github.com/sarkartanmay393/RoomReservation-WebApp/internal/driver"
	"github.com/sarkartanmay393/RoomReservation-WebApp/internal/helpers"
	"github.com/sarkartanmay393/RoomReservation-WebApp/internal/models"
	"github.com/sarkartanmay393/RoomReservation-WebApp/internal/render"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

var app config.AppConfig
var session *scs.SessionManager
var pathToTemplates = "web/templates"
var functions template.FuncMap

func getRoutes() http.Handler {

	gob.Register(&models.Reservation{})
	gob.Register(&models.ChosenDates{})
	gob.Register(&models.Restriction{})
	gob.Register(&models.Room{})
	gob.Register(&models.User{})
	gob.Register(&models.RoomRestriction{})

	log.Println("Connecting to database...")
	dsn := "host=localhost port=5432 dbname=roomreservation user=postgres password=Tanmay3597!"
	dbconn, err := driver.ConnectSQL(dsn)
	if err != nil {
		log.Fatalln("unable to connect database: ", err)
	}
	log.Println("Connected to database!")

	// Starting of Logging information.
	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Creating template cache for the whole app to get started.
	app.TemplateCache, err = render.CreateTemplateCache()
	app.UseCache = true      // Manual value setup in app config.
	app.InProduction = false // Manual value setup in app config.
	if err != nil {
		log.Print("Error creating template cache\n") // No newline because of testing.
	}
	render.AttachConfig(&app)                    // appConfig is transferred to render.go file.
	temporaryRepo := CreateNewRepo(&app, dbconn) // Creates Repo with appConfig and db connection pool to be transferred.
	helpers.ConnectToHelpers(&app)               // App config now to helpers package.
	AttachRepo(temporaryRepo)
	// Now application config will be available in render.go and handlers.go file.

	// Session Management Implementation
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.SessionManager = session // Transfers this session object to app config.

	mux := chi.NewRouter() // Instance of router using chi package.

	mux.Use(middleware.Recoverer) // Tackle panic attack as a middleware.
	// For POST requests we have to turn off CSRF Check middleware.
	// mux.Use(CSRFCheckTest)                  // Checks for Cross-site request forgery attacks.
	mux.Use(app.SessionManager.LoadAndSave) // Loads and saves session data.

	mux.Get("/", http.HandlerFunc(Repo.HomeHandler))               // Serve root page request.
	mux.Get("/form", http.HandlerFunc(Repo.CoedHandler))           // Serve /form page request.
	mux.Get("/singlebed", http.HandlerFunc(Repo.SinglebedHandler)) // Serve /form page request.
	mux.Get("/coed", http.HandlerFunc(Repo.CoedHandler))           // Serve /form page request.
	mux.Get("/highland", http.HandlerFunc(Repo.HighlandHandler))   // Serve /form page request.

	mux.Get("/reservation", http.HandlerFunc(Repo.ReservationHandler))        // Serve /form page request.
	mux.Post("/reservation", http.HandlerFunc(Repo.PostReservationHandler))   // Serve /form page request.
	mux.Post("/reservation-json", http.HandlerFunc(Repo.AvailabilityHandler)) // Serve /form page request.

	mux.Get("/contact", http.HandlerFunc(Repo.ContactHandler)) // Serve /form page request.

	// Serves get and post request on 'make-reservation' path.
	mux.Get("/make-reservation", http.HandlerFunc(Repo.MakeReservationHandler))
	mux.Post("/make-reservation", http.HandlerFunc(Repo.PostMakeReservationHandler))
	mux.Get("/reservation-summary", http.HandlerFunc(Repo.ReservationSummaryHandler))

	fileServer := http.FileServer(http.Dir("/static/"))               // fileServer handles file system contents.
	mux.Handle("/static/*", http.StripPrefix("/static/", fileServer)) // Handles all files.

	return mux // Returns the http.handler for further use in main.go
}

// CreateTestTemplateCache creates a map of web.
func CreateTestTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}
	for _, page := range pages {

		name := filepath.Base(page) // "/template/home.page.tmpl" -> "home/page.tmpl"
		templateSet, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		layouts, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}
		if len(layouts) > 0 {
			templateSet, err = templateSet.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = templateSet
	}
	return myCache, nil
}
