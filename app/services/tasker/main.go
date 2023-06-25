package main

/*
* how to read an env file from a mounted volume? with this I can remove line 18 from dockerfile.tasker
 */

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	conf "github.com/ardanlabs/conf/v3"
	"github.com/jnkroeker/khyme/app/services/tasker/handlers"
	"github.com/jnkroeker/khyme/business/sys/database"
	"github.com/jnkroeker/khyme/business/web/auth"
	"github.com/jnkroeker/khyme/foundation/vault"
	"github.com/joho/godotenv"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var build = "develop"

// Here we are in the app layer (folder) of the tasker project
// App layer is the "presentation" layer
// the app layer is responsible for start-up and shutdown of tasker
// and accepting user input and providing output
// the app layer will call into the business layer (folder) with the input it collected

func main() {

	// construct the application logger
	log, err := initLogger("KHYME-TASKER")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer log.Sync()

	// perform start-up and shutdown sequence
	if err := run(log); err != nil {
		log.Errorw("start-up", "ERROR", err)
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {
	// ========================================================================================
	// GOMAXPROCS

	// Set the correct number of threads for the service
	// either by what is available on the machine or some quota
	if _, err := maxprocs.Set(); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}

	log.Infow("start-up", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// ========================================================================================
	// Configuration

	// There is only one place for all configuration to be loaded from; here in main.go of the main package
	// no other package can access the configuration other than here
	// every configuration has a default that works atleast in Development
	// the more defaults can work across your environments, the better
	// the service should work out-of-the-box with defaults

	// A literal struct cannot be passed around the program
	cfg := struct {
		conf.Version
		Task struct {
			ServiceHost     string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s,mask"`
			Set             string        `conf:"default:TASKS"`
			Queue           string        `conf:"default:Q"`
			Dlq             string        `conf:"default:DLQ"`
			BatchSize       int           `conf:"default:0"`
		}
		Vault struct {
			Address   string `conf:"default:vault.khyme-system"`
			MountPath string `conf:"default:secret"`
			Token     string `conf:"default:mytoken,mask"`
		}
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:postgres, mask"`
			Host         string `conf:"default:database-service.database-system"` // pod-to-pod comms with service name
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:0"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "copywright of Hadeda, LLC",
		},
	}

	// Read the .env file into khymeEnv map
	var khymeEnv map[string]string
	khymeEnv, err := godotenv.Read()
	if err != nil {
		return fmt.Errorf("error fetching environment variables: %w", err)
	}

	// Set environment variables using khymeEnv map
	os.Clearenv()
	for k, v := range khymeEnv {
		os.Setenv(k, v)
	}

	const prefix = "TASK"

	// Parse environment variables from the commandline for variables starting with the prefix
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// ========================================================================================
	// App Starting

	log.Infow("starting Tasker service", "version", build)
	defer log.Infow("shutdown of Tasker service complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	// ========================================================================================
	// Database Support

	// Create connectivity to the database.
	log.Infow("startup", "status", "initializing database support", "host", cfg.DB.Host)

	db, err := database.Open(database.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		log.Infow("startup", "status", "database initialization failed", "host", cfg.DB.Host)
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer func() {
		log.Infow("shutdown", "status", "stopping database support", "host", cfg.DB.Host)
		db.Close()
	}()

	// ========================================================================================
	// Authentication Support

	log.Infow("startup", "status", "initializing authentication support")

	vault, err := vault.New(vault.Config{
		Address:   cfg.Vault.Address,
		Token:     cfg.Vault.Token,
		MountPath: cfg.Vault.MountPath,
	})
	if err != nil {
		return fmt.Errorf("constructing vault: %w", err)
	}

	authCfg := auth.Config{
		Log:       log,
		DB:        db,
		KeyLookup: vault,
	}

	auth, err := auth.New(authCfg)
	if err != nil {
		return fmt.Errorf("constructing auth: %w", err)
	}

	// ========================================================================================
	// Start Debug Service

	log.Infow("startup", "status", "debug router started", "host", cfg.Task.DebugHost)

	// The Debug function returns a mux to listen and serve on for all the debug
	// related endpoints. this includes the standard library endpoints and our own.

	// Construct the mux for debugging
	debugMux := handlers.DebugMux(build, log, db)

	// Start the service listening for debug requests
	// Not concerned about shutting this down with load shedding
	go func() {
		if err := http.ListenAndServe(cfg.Task.DebugHost, debugMux); err != nil {
			log.Errorw("shutdown", "status", "debug router closed", "host", cfg.Task.DebugHost, "ERROR", err)
		}
	}()

	// ========================================================================================
	// Start API Service

	log.Infow("startup", "status", "initializing Tasker API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Construct the MUX for the API calls
	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Shutdown: shutdown,
		Log:      log,
		Auth:     auth,
		DB:       db,
	})

	// In order to implement load-shedding, (aka on shutdown the goroutines currently handling requests can complete)
	// we need an http server. Load-shedding wont work on http.ListenAndServe
	// Construct a server to service the requests against the mux.
	api := http.Server{
		Addr:         cfg.Task.ServiceHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.Task.ReadTimeout,
		WriteTimeout: cfg.Task.WriteTimeout,
		IdleTimeout:  cfg.Task.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	/*
	 * RULE: If a goroutine creates another goroutine, it is responsible for its child.
	 *       Child goroutines should terminate BEFORE their parents
	 *       and the parent should be aware of the child's termination.
	 */

	// Make a channel to listen for errors coming from the listener.
	// Use a buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for Tasker api requests
	/*
	 * this goroutine is the parent of all goroutines created to handle Tasker requests
	 * all incoming requests are initially serviced by the api server's Handler method (the onion's outer-most layer)
	 */
	go func() {
		log.Infow("startup", "status", "Tasker api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// ========================================================================================
	// Shutdown

	// This 'blocking select' holds the Tasker service running until it is time for shutdown.
	// It listens for signals coming from the two previously created channels, serverErrors and shutdown
	// We can't load-shed on serverErrors, because something low level errored
	// On a ctrl-c or a kubernetes shutdown message we can load-shed
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "Tasker shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", " Tasker shutdown complete", "signal", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Task.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and shed load
		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

func initLogger(service string) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]interface{}{
		"service": "KHYME_TASKER",
	}

	log, err := config.Build()
	if err != nil {
		fmt.Println("Error constructing logger: ", err)
		os.Exit(1)
	}

	return log.Sugar(), nil
}
