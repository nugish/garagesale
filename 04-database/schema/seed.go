package schema

import "github.com/jmoiron/sqlx"

// seeds is a string constant containing all the queries needed to get the
// db seeded to a useful state for development.
//
// Using a constant in a .go file is an easy way to ensure the queries are part
// of the compiled executable and avoids pathing issues with the working
// directory. It has the downside that it lacks syntax highlighting and may be
// harder to read for some cases compared to using .sql files. You may also
// consider a combined approach using a tool like packr or go-bindata.
//
// Note that database servers beside PostgreSQL may not support running
// multiple queries as part of the same execution so this single large constant
// may need to be broken up.

const seed = `
	-- Create sample products
	INSERT INTO products (product_id, name, cost, quantity, date_created, date_updated) VALUES
	('a2b0639f-2cc6-44b8-b97b-15d69dbb511e', 'Comic Books', 50, 42, '2019-01-01 00:00:01.000001+00', '2019-01-01 00:00:01.000001+00'),
	('72f8b983-3eb4-48db-9ed0-e45cc6bd716b', 'McDonalds Toys', 75, 120, '2019-01-01 00:00:02.000001+00', '2019-01-01 00:00:02.000001+00')
	ON CONFLICT DO NOTHING;
`

// Seed runs the above query to add some data and bring the database into a usefule state.
func Seed(db *sqlx.DB) error {

	// Using transactions in case a rollback is needed when errors occur
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(seed); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}
