package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/lib/pq" // Imports Postgres SQL driver
)

type Retriever interface {
	Retrieve(context.Context) ([]string, error)
}

type SystemRetriever struct {
	Dsn string
}

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

type FilterType int

const (
	ExcludeFilterType FilterType = iota
	OnlyFilterType
)

type FilteredRetriever struct {
	R      Retriever
	Filter FilterType
	DBs    []string
}

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
