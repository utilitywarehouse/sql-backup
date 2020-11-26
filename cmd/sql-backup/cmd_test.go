package main

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"github.com/utilitywarehouse/sql-backup/internal/db"
	"github.com/utilitywarehouse/sql-backup/internal/dbcli"
	"github.com/utilitywarehouse/sql-backup/internal/pool"
	"github.com/utilitywarehouse/sql-backup/internal/store"
)

func TestRetrieverFromFlags_InvalidDsn(t *testing.T) {
	set := &flag.FlagSet{}
	set.String("dbcli-dsn", "[fe80::1%en0]", "")

	c := cli.NewContext(&cli.App{}, set, nil)

	_, err := retrieverFromFlags(c)
	assert.Error(t, err)
}

func TestRetrieverFromFlags_ValidDsn(t *testing.T) {
	set := &flag.FlagSet{}
	set.String("dbcli-dsn", "localhost/dbname", "")

	c := cli.NewContext(&cli.App{}, set, nil)

	r, err := retrieverFromFlags(c)
	assert.Nil(t, err)
	assert.IsType(t, db.SystemRetriever{}, r)

	systemR, ok := r.(db.SystemRetriever)
	assert.True(t, ok)
	assert.Equal(t, "postgresql://localhost/dbname", systemR.Dsn)
}

func TestRetrieverFromFlags_Only(t *testing.T) {
	expected := []string{"foo", "bar", "egg"}

	set := &flag.FlagSet{}
	strSlice := &cli.StringSlice{}
	for _, v := range expected {
		assert.NoError(t, strSlice.Set(v))
	}
	set.Var(strSlice, "only", "")

	c := cli.NewContext(&cli.App{}, set, nil)

	r, err := retrieverFromFlags(c)
	assert.Nil(t, err)
	assert.IsType(t, db.FilteredRetriever{}, r)

	fixedR, ok := r.(db.FilteredRetriever)
	assert.True(t, ok)
	assert.Equal(t, db.OnlyFilterType, fixedR.Filter)
	assert.Equal(t, expected, fixedR.DBs)
}

func TestRetrieverFromFlags_Exclude(t *testing.T) {
	expected := []string{"foo", "bar", "egg"}

	set := &flag.FlagSet{}
	strSlice := &cli.StringSlice{}
	for _, v := range expected {
		assert.NoError(t, strSlice.Set(v))
	}
	set.Var(strSlice, "exclude", "")
	set.String("dbcli-dsn", "user:pw@localhost:3242", "")

	c := cli.NewContext(&cli.App{}, set, nil)

	r, err := retrieverFromFlags(c)
	assert.Nil(t, err)
	assert.IsType(t, db.FilteredRetriever{}, r)

	excludeRetriever, ok := r.(db.FilteredRetriever)
	assert.True(t, ok)
	assert.Equal(t, db.ExcludeFilterType, excludeRetriever.Filter)
	assert.Equal(t, expected, excludeRetriever.DBs)
}

func TestDumperFromFlags_InvalidBinaryPath(t *testing.T) {
	set := &flag.FlagSet{}
	set.String("dbcli-binary", "invalid-dbcli", "")

	c := cli.NewContext(&cli.App{}, set, nil)

	_, err := dumperFromFlags(c)
	assert.NotNil(t, err)
}

func TestDumperFromFlags_ValidBinaryPath(t *testing.T) {
	f, err := ioutil.TempFile("", "TestDumperFromFlags_ValidBinaryPath")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	set := &flag.FlagSet{}
	set.String("dbcli-binary", f.Name(), "")

	c := cli.NewContext(&cli.App{}, set, nil)

	d, err := dumperFromFlags(c)
	assert.Nil(t, err)
	assert.IsType(t, dbcli.CliDumper{}, d)

	cliDumper, ok := d.(dbcli.CliDumper)
	assert.True(t, ok)
	assert.Equal(t, f.Name(), cliDumper.Cmd)
	assert.Equal(t, cliDumper.Timeout, time.Duration(0))
}

func TestDumperFromFlags_Timeout(t *testing.T) {
	f, err := ioutil.TempFile("", "TestDumperFromFlags_ValidBinaryPath")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	expected := 30 * time.Minute

	set := &flag.FlagSet{}
	set.String("dbcli-binary", f.Name(), "")
	set.Duration("dbcli-timeout", expected, "")

	c := cli.NewContext(&cli.App{}, set, nil)

	d, err := dumperFromFlags(c)
	assert.Nil(t, err)
	assert.IsType(t, dbcli.CliDumper{}, d)

	cliDumper, ok := d.(dbcli.CliDumper)
	assert.True(t, ok)
	assert.Equal(t, expected, cliDumper.Timeout)
}

func TestPoolFromFlags_PoolSize(t *testing.T) {
	expected := 10

	set := &flag.FlagSet{}
	set.Int("pool", expected, "")

	c := cli.NewContext(&cli.App{}, set, nil)

	p := poolFromFlags(c)
	assert.IsType(t, pool.SizablePool{}, p)

	sizablePool, ok := p.(pool.SizablePool)
	assert.True(t, ok)
	assert.Equal(t, expected, sizablePool.Size)
}

func TestStorerFromFlags_DefaultFile(t *testing.T) {
	expected := "/some/dir"

	set := &flag.FlagSet{}
	set.String("dir", expected, "")

	c := cli.NewContext(&cli.App{}, set, nil)

	s := storerFromFlags(c)
	assert.IsType(t, store.File{}, s)

	fileStorer, ok := s.(store.File)
	assert.True(t, ok)
	assert.Equal(t, expected, fileStorer.Dir)
}

func TestStorerFromFlags_AwsBucket(t *testing.T) {
	expected := "some-bucket-goes-here"

	set := &flag.FlagSet{}
	set.String("driver", "aws", "")
	set.String("bucket", expected, "")

	c := cli.NewContext(&cli.App{}, set, nil)

	s := storerFromFlags(c)
	assert.IsType(t, store.S3{}, s)

	s3Storer, ok := s.(store.S3)
	assert.True(t, ok)
	assert.Equal(t, expected, s3Storer.Bucket)
}
