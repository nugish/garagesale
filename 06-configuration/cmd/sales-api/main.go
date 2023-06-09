package main

import (
	"context"
	"fmt"
	"github.com/nugish/garagesale/internal/platform/conf"
	"time"

	"github.com/nugish/garagesale/cmd/sales-api/internal/handlers"
	"github.com/nugish/garagesale/internal/platform/database"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	var cfg struct {
		Web struct {
			Address         string        `conf:"default:localhost:8000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		DB struct {
			User       string `conf:"default:postgres"`
			Password   string `conf:"default:postgres,noprint"`
			Host       string `conf:"default:localhost"`
			Name       string `conf:"default:postgres"`
			DisableTLS bool   `conf:"default:false"`
		}
	}

	// =============================================================================
	// App starting

	log.Printf("main: Started")
	defer log.Println("main: Completed")

	// =============================================================================
	// Get Configuration

	if err := conf.Parse(os.Args[1:], "sales", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("SALES", &cfg)
			if err != nil {
				log.Fatalf("error: generating config usage: %v", err)
			}
			fmt.Println(usage)
			return
		}
		log.Fatalf("error: parsing config: %s", err)
	}

	out, err := conf.String(&cfg)
	if err != nil {
		log.Fatalf("error: generating config for output: %v", err)
	}
	log.Printf("main: Config: \n%v\n", out)

	// =============================================================================
	// Setup Dependencies

	db, err := database.Open(database.Config{
		Host:       cfg.DB.Host,
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	})
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// =============================================================================
	// Start API Service

	ps := handlers.Product{DB: db}

	api := http.Server{
		Addr:         cfg.Web.Address,
		Handler:      http.HandlerFunc(ps.List),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("main: API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// =============================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("error: listening and serving: %s", err)

	case <-shutdown:
		log.Println("main: Start shutdown")

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shut down and load shed.
		err := api.Shutdown(ctx)
		if err != nil {
			log.Printf("main: Graceful shutdown did not complete in %v: %v", cfg.Web.ShutdownTimeout, err)
			err = api.Close()
		}

		if err != nil {
			log.Fatalf("main: could not stop server gracefully: %v", err)
		}
	}
}
