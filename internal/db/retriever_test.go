package db_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/db-backup/internal/db"
)

func TestFilteredRetriever_Exclude(t *testing.T) {
	expected := []string{"two"}

	r := db.FilteredRetriever{
		R:      StubbedRetriever{DBs: []string{"one", "two", "three"}},
		Filter: db.ExcludeFilterType,
		DBs:    []string{"one", "three"},
	}

	dbs, err := r.Retrieve(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expected, dbs)
}

func TestFilteredRetriever_ExcludeNotFound(t *testing.T) {
	r := db.FilteredRetriever{
		R:      StubbedRetriever{DBs: []string{"one", "two", "three"}},
		Filter: db.ExcludeFilterType,
		DBs:    []string{"one", "four"},
	}

	_, err := r.Retrieve(context.Background())
	assert.NotNil(t, err)
}

func TestFilteredRetriever_Only(t *testing.T) {
	expected := []string{"one", "three"}

	r := db.FilteredRetriever{
		R:      StubbedRetriever{DBs: []string{"one", "two", "three"}},
		Filter: db.OnlyFilterType,
		DBs:    expected,
	}

	dbs, err := r.Retrieve(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expected, dbs)
}

func TestFilteredRetriever_OnlyNotFound(t *testing.T) {
	r := db.FilteredRetriever{
		R:      StubbedRetriever{DBs: []string{"one", "two", "three"}},
		Filter: db.OnlyFilterType,
		DBs:    []string{"one", "three", "four"},
	}

	_, err := r.Retrieve(context.Background())
	assert.NotNil(t, err)
}

type StubbedRetriever struct {
	DBs []string
}

func (r StubbedRetriever) Retrieve(ctx context.Context) ([]string, error) {
	return r.DBs, nil
}
