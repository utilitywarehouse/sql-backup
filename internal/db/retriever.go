package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/lib/pq" // Imports Postgres SQL driver
)

const (
	// ExcludeFilterType ...
	ExcludeFilterType FilterType = iota
	// OnlyFilterType ...
	OnlyFilterType
)

// Retriever is an interface to a Retriever function
type Retriever interface {
	Retrieve(context.Context) ([]string, error)
}

// SystemRetriever is an instance of a Retreiver
type SystemRetriever struct {
	Dsn string
}

// FilterType is used as an enum of filter types
type FilterType int

// FilteredRetriever is a retreiever with filters applied
type FilteredRetriever struct {
	R      Retriever
	Filter FilterType
	DBs    []string
}

// NewSystemRetriever returns a popualted SystemRetriever
func NewSystemRetriever(dsn string) (SystemRetriever, error) {
	if !strings.HasPrefix(dsn, "postgresql://") {
		dsn = "postgresql://" + dsn
	}

	url, err := url.Parse(dsn)
	if err != nil {
		return SystemRetriever{}, err
	}
	return SystemRetriever{Dsn: url.String()}, nil
}

// Retrieve retrieves the list of databases from a DB host.
func (r SystemRetriever) Retrieve(ctx context.Context) ([]string, error) {
	db, err := sql.Open("postgres", r.Dsn)
	if err != nil {
		return []string{}, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT datname FROM pg_database WHERE datistemplate = false AND datname != 'system'")
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()

	var dbs []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return []string{}, err
		}
		dbs = append(dbs, name)
	}

	return dbs, nil
}

// Retrieve is an isntance of a Retreiver with appllied filters
func (r FilteredRetriever) Retrieve(ctx context.Context) ([]string, error) {
	found, err := r.R.Retrieve(ctx)
	if err != nil {
		return found, err
	}

	var filteredDBs []string
	if r.Filter == ExcludeFilterType {
		filteredDBs = found
		for i, exclude := range r.DBs {
			var matched bool
			for _, db := range found {
				if db == exclude {
					matched = true
				}
			}
			if !matched {
				return []string{}, fmt.Errorf("unable to find database: %s", exclude)
			}

			filteredDBs = append(filteredDBs[:i], filteredDBs[i+1:]...)
		}
	}
	if r.Filter == OnlyFilterType {
		for _, only := range r.DBs {
			var matched bool
			for _, db := range found {
				if db == only {
					filteredDBs = append(filteredDBs, db)
					matched = true
				}
			}
			if !matched {
				return filteredDBs, fmt.Errorf("unable to find database: %s", only)
			}
		}
	}

	return filteredDBs, nil
}
