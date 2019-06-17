package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilename(t *testing.T) {
	db := "users"
	date := time.Now().Format("2006-01-02")

	inputs := []struct {
		backupFormat       string
		disableCompression bool
		expected           string
	}{
		{
			"%s_2006-01-02.sql",
			false,
			fmt.Sprintf("%s_%s.sql.gz", db, date),
		},
		{
			"%s_2006-01-02.sql",
			true,
			fmt.Sprintf("%s_%s.sql", db, date),
		},
		{
			"%s_2006-01-02.sql.gz",
			false,
			fmt.Sprintf("%s_%s.sql.gz", db, date),
		},
	}

	for _, input := range inputs {
		o := &once{
			BackupFormat:       input.backupFormat,
			DisableCompression: input.disableCompression,
		}

		filename := o.filename(db)
		assert.Equal(t, input.expected, filename)
	}
}
