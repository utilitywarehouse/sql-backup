package main

import (
	"context"

	heartbeat "github.com/rcrowe/opsgenie-heartbeat"
	"github.com/urfave/cli"
	"github.com/utilitywarehouse/db-backup/internal/db"
	"github.com/utilitywarehouse/db-backup/internal/dbcli"
	"github.com/utilitywarehouse/db-backup/internal/pool"
	"github.com/utilitywarehouse/db-backup/internal/store"
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
	if c.GlobalString("driver") == "aws" {
		return store.S3{
			Bucket: c.GlobalString("bucket"),
		}
	}
	return store.File{
		Dir: c.GlobalString("dir"),
	}
}

// heartbeater lets you notify that the service ran correctly
type heartbeater interface {
	ping(ctx context.Context) error
}

// noopHeartbeat swallows any calls
type noopHeartbeat struct{}

// ping that the service ran successfully
func (n *noopHeartbeat) ping(ctx context.Context) error {
	return nil
}

// opsgenieHeartbeat pings a Opsgenie heartbeat
type opsgenieHeartbeat struct {
	client heartbeat.PingRequest
	name   string
}

// ping that the service ran successfully
func (n *opsgenieHeartbeat) ping(ctx context.Context) error {
	return n.client.Ping(ctx, n.name)
}
