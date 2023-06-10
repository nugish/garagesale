# 4. Database

- Start a PostgreSQL server running in docker container
- Setup schema
- Seed database
- Connect to database from services

## Notes:

- Executing schema changes requires elevated privileges. The normal api service 
should be running as a DB user with less access
- Using `Select *` has problems.

```
## Start postgres:
docker-compose up -d

## Create the schema and insert some seed data.
go build
./garagesale migrate
./garagesale seed

## Run the app then make requests.
./garagesale
```

## Links:

- [Surprises, Antipatterns and Limitations (of `database/sql`)](http://go-database-sql.org/surprises.html)
