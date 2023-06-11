package database

import (
	"net/url"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Register the postgres database/sql driver
)

// Open opens a database connection
func Open() (*sqlx.DB, error) {
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
