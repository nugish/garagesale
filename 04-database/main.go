package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/nugish/garagesale/schema"
)

func main() {

	// =============================================================================
	// App starting

	log.Printf("main: Started")
	defer log.Println("main: Completed")

	// =============================================================================
	// Setup Dependencies

	db, err := openDB()
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	flag.Parse()
	switch flag.Arg(0) {
	case "migrate":
		if err := schema.Migrate(db); err != nil {
			log.Fatal("applying migrations", err)
		}
		log.Println("migrations complete")
	case "seed":
		if err := schema.Seed(db); err != nil {
			log.Fatal("applying seed data ", err)
		}
		log.Println("seed data inserted")
	}

	// =============================================================================
	// Start API Service

	ps := ProductService{db: db}

	api := http.Server{
		Addr:         "localhost:8000",
		Handler:      http.HandlerFunc(ps.List),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
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
		const timeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Asking listener to shut down and load shed.
		err := api.Shutdown(ctx)
		if err != nil {
			log.Printf("main: Graceful shutdown did not complete in %v: %v", timeout, err)
			err = api.Close()
		}

		if err != nil {
			log.Fatalf("main: could not stop server gracefully: %v", err)
		}
	}
}

func openDB() (*sqlx.DB, error) {
	q := url.Values{}
	q.Set("sslmode", "disable")
	q.Set("timezone", "Asia/Jakarta")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword("garagesale", "mypassword1223"),
		Host:     "localhost:5432",
		Path:     "garagesale",
		RawQuery: q.Encode(),
	}

	return sqlx.Open("postgres", u.String())
}

type Product struct {
	ID          string    `db:"product_id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Cost        int       `db:"cost" json:"cost"`
	Quantity    int       `db:"quantity" json:"quantity"`
	DateCreated time.Time `db:"date_created" json:"date_created"`
	DateUpdated time.Time `db:"date_updated" json:"date_updated"`
}

// ProductService has handler methods for dealing with products.
type ProductService struct {
	db *sqlx.DB
}

// List gives all products as a list.
func (p *ProductService) List(w http.ResponseWriter, r *http.Request) {

	list := []Product{}

	const q = `SELECT product_id, name, cost, quantity, date_updated, date_created FROM products`
	if err := p.db.Select(&list, q); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error query db", err)
		return
	}

	data, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error marshalling", err)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset:utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(data); err != nil {
		log.Println("error writing", err)
	}
}
