package main

import (
	"github.com/urfave/cli"
	"github.com/utilitywarehouse/sql-backup/internal/db"
	"github.com/utilitywarehouse/sql-backup/internal/dbcli"
	"github.com/utilitywarehouse/sql-backup/internal/pool"
	"github.com/utilitywarehouse/sql-backup/internal/store"
)

func retrieverFromFlags(c *cli.Context) (db.Retriever, error) {
	dsn := c.GlobalString("dbcli-dsn")
	systemRetriever, err := db.NewSystemRetriever(dsn)
	if err != nil {
		return db.SystemRetriever{}, err
	}

	if only := c.GlobalStringSlice("only"); len(only) > 0 {
		return db.FilteredRetriever{
			R:      systemRetriever,
			Filter: db.OnlyFilterType,
			DBs:    only,
		}, nil
	}
	if exclude := c.GlobalStringSlice("exclude"); len(exclude) > 0 {
		return db.FilteredRetriever{
			R:      systemRetriever,
			Filter: db.ExcludeFilterType,
			DBs:    exclude,
		}, nil
	}
	return systemRetriever, nil
}

func dumperFromFlags(c *cli.Context) (dbcli.Dumper, error) {
	dumper, err := dbcli.NewDumper(c.GlobalString("dbcli-binary"), c.GlobalString("dbcli-flags"), c.GlobalString("dbcli-dsn"))
	if err != nil {
		return nil, err
	}
	if duration := c.GlobalDuration("dbcli-timeout"); duration.Seconds() != 0 {
		dumper.Timeout = duration
	}
	return dumper, nil
}

func poolFromFlags(c *cli.Context) pool.Pooler {
	return pool.SizablePool{Size: c.GlobalInt("pool")}
}

func storerFromFlags(c *cli.Context) store.Storer {
	switch c.GlobalString("driver") {
	case "aws":
		return store.S3{Bucket: c.GlobalString("bucket"), Dir: c.GlobalString("dir")}
	case "gcp":
		return store.GCS{Bucket: c.GlobalString("bucket"), Dir: c.GlobalString("dir")}
	default:
		return store.File{Dir: c.GlobalString("dir")}
	}
}
