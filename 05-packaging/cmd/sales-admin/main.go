package main

import (
	"flag"
	"log"

	"github.com/nugish/garagesale/internal/platform/database"
	"github.com/nugish/garagesale/internal/schema"
)

func main() {

	// =============================================================================
	// Setup Dependencies

	db, err := database.Open()
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
}
