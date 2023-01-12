package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/joho/godotenv"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var build = "develop"

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
	}{
		Version: conf.Version{
			SVN:  build,
			Desc: "copywright of Hadeda, LLC",
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

	// Parse environment variables for variables starting with the prefix
	help, err := conf.ParseOSArgs(prefix, &cfg)
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

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

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