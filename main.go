package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/sirodoht/heartfort/controllers"
	"github.com/sirodoht/heartfort/email"
	"github.com/sirodoht/heartfort/middleware"
	"github.com/sirodoht/heartfort/models"
	"github.com/sirodoht/heartfort/rand"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

func main() {
	boolPtr := flag.Bool("prod", false, "Provide this flag "+
		"in production. This ensures that a .config file is "+
		"provided before the application starts.")
	flag.Parse()
	cfg := LoadConfig(*boolPtr)
	dbCfg := cfg.Database
	services, err := models.NewServices(
		models.WithGorm(dbCfg.Dialect(), dbCfg.ConnectionInfo()),
		models.WithLogMode(!cfg.IsProd()),
		models.WithUser(cfg.Pepper, cfg.HMACKey),
		models.WithJob(),
	)
	if err != nil {
		panic(err)
	}
	defer services.Close()
	services.AutoMigrate()

	mgCfg := cfg.Mailgun
	emailer := email.NewClient(
		email.WithSender("Heartfort Support", "support@"+mgCfg.Domain),
		email.WithMailgun(mgCfg.Domain, mgCfg.APIKey, mgCfg.PublicAPIKey),
	)

	r := mux.NewRouter()
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(services.User, emailer)
	jobsC := controllers.NewJobs(services.Job, r)

	userMw := middleware.User{
		UserService: services.User,
	}
	requireUserMw := middleware.RequireUser{}

	r.Handle("/", staticC.Home).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.Handle("/logout",
		requireUserMw.ApplyFn(usersC.Logout)).
		Methods("GET")
	r.Handle("/forgot", usersC.ForgotPwView).Methods("GET")
	r.HandleFunc("/forgot", usersC.InitiateReset).Methods("POST")
	r.HandleFunc("/reset", usersC.ResetPw).Methods("GET")
	r.HandleFunc("/reset", usersC.CompleteReset).Methods("POST")
	r.HandleFunc("/cookies", usersC.Cookies).Methods("GET")

	// Job routes
	r.Handle("/jobs",
		requireUserMw.ApplyFn(jobsC.Index)).
		Methods("GET").
		Name(controllers.IndexJobs)
	r.Handle("/jobs/new",
		requireUserMw.Apply(jobsC.New)).
		Methods("GET")
	r.Handle("/jobs",
		requireUserMw.ApplyFn(jobsC.Create)).
		Methods("POST")
	r.HandleFunc("/jobs/{id:[0-9]+}",
		jobsC.Show).
		Methods("GET").
		Name(controllers.ShowJob)
	r.HandleFunc("/jobs/{id:[0-9]+}/edit",
		requireUserMw.ApplyFn(jobsC.Edit)).
		Methods("GET").
		Name(controllers.EditJob)
	r.HandleFunc("/jobs/{id:[0-9]+}/update",
		requireUserMw.ApplyFn(jobsC.Update)).
		Methods("POST")
	r.HandleFunc("/jobs/{id:[0-9]+}/delete",
		requireUserMw.ApplyFn(jobsC.Delete)).
		Methods("POST")

	// Assets
	assetHandler := http.FileServer(http.Dir("./assets/"))
	assetHandler = http.StripPrefix("/assets/", assetHandler)
	r.PathPrefix("/assets/").Handler(assetHandler)

	b, err := rand.Bytes(32)
	if err != nil {
		panic(err)
	}
	// Use the config's IsProd method instead
	csrfMw := csrf.Protect(b, csrf.Secure(cfg.IsProd()))

	// Serve
	fmt.Printf("Starting the server on :%d...\n", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), csrfMw(userMw.Apply(r)))
}