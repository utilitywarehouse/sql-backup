package main

import (
	"fmt"
	"path/filepath"
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
		dumpPrefix string
	}{
		{
			"%s_2006-01-02.sql",
			false,
			fmt.Sprintf("%s_%s.sql.gz", db, date),
			"test",
		},
		{
			"%s_2006-01-02.sql",
			true,
			fmt.Sprintf("%s_%s.sql", db, date),
			"test",
		},
		{
			"%s_2006-01-02.sql.gz",
			false,
			fmt.Sprintf("%s_%s.sql.gz", db, date),
			"test",
		},
	}

	for _, input := range inputs {
		o := &once{
			BackupFormat:       input.backupFormat,
			DisableCompression: input.disableCompression,
			DumpPrefix: input.dumpPrefix,
		}

		filename := filepath.Join(o.DumpPrefix, o.filename(db))
		assert.Equal(t,  filepath.Join(o.DumpPrefix, input.expected), filename)
	}
}
