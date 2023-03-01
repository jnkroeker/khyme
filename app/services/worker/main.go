package main

/*
 * how to read an env file from a mounted volume? with this I can remove line 18 from dockerfile.worker
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

	"github.com/ardanlabs/conf/v3"
	"github.com/jnkroeker/khyme/app/services/tasker/handlers"
	"github.com/joho/godotenv"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var build = "develop"

func main() {

	// construct the application logger
	log, err := initLogger("KHYME-WORKER")
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

	cfg := struct {
		conf.Version
		Worker struct {
			ServiceHost     string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s,mask"`
			Workdir         string        `conf:"default:/"`
			DockerUser      string        `conf:"default:admin"`
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

	const prefix = "WORKER"

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

	log.Infow("starting Worker service", "version", build)
	defer log.Infow("shutdown of Worker service complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	// ========================================================================================
	// Start Debug Service

	log.Infow("startup", "status", "debug router started", "host", cfg.Worker.DebugHost)

	// The Debug function returns a mux to listen and serve on for all the debug
	// related endpoints. this includes the standard library endpoints.

	// Construct the mux for debugging
	debugMux := handlers.DebugStandardLibraryMux()

	// Start the service listening for debug requests
	// Not concerned about shutting this down with load shedding
	go func() {
		if err := http.ListenAndServe(cfg.Worker.DebugHost, debugMux); err != nil {
			log.Errorw("shutdown", "status", "debug router closed", "host", cfg.Worker.DebugHost, "ERROR", err)
		}
	}()

	// ========================================================================================
	// Start API Service

	log.Infow("startup", "status", "initializing Worker API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// In order to implement load-shedding, (aka on shutdown the goroutines currently handling requests can complete)
	// we need an http server. Load-shedding wont work on http.ListenAndServe
	// Construct a server to service the requests against the mux.
	api := http.Server{
		Addr:         cfg.Worker.ServiceHost,
		Handler:      nil,
		ReadTimeout:  cfg.Worker.ReadTimeout,
		WriteTimeout: cfg.Worker.WriteTimeout,
		IdleTimeout:  cfg.Worker.IdleTimeout,
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
	 * this goroutine is the parent of all goroutines created to handle Worker requests
	 * all incoming requests are initially serviced by the api server's Handler method (the onion's outer-most layer)
	 */
	go func() {
		log.Infow("startup", "status", "Worker api router started", "host", api.Addr)
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
		log.Infow("shutdown", "status", "Worker shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "Worker shutdown complete", "signal", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Worker.ShutdownTimeout)
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
		"service": "KHYME_WORKER",
	}

	log, err := config.Build()
	if err != nil {
		fmt.Println("Error constructing logger: ", err)
		os.Exit(1)
	}

	return log.Sugar(), nil
}
