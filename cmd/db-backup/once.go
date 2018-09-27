package main

import (
	"compress/gzip"
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/utilitywarehouse/db-backup/internal/db"
	"github.com/utilitywarehouse/db-backup/internal/dbcli"
	"github.com/utilitywarehouse/db-backup/internal/pool"
	"github.com/utilitywarehouse/db-backup/internal/store"
)

// OnceCmd is used to have a one time run of a backup
type OnceCmd struct{}

// Run executes a one time backup
func (cmd *OnceCmd) Run(c *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sCh := make(chan os.Signal)
		signal.Notify(sCh, os.Interrupt, syscall.SIGTERM)
		<-sCh

		log.Println("Shutdown requested")
		cancel()
	}()

	o, err := onceFromFlags(c)
	if err != nil {
		return err
	}

	return o.Backup(ctx)
}

type once struct {
	Retriever          db.Retriever
	Dumper             dbcli.Dumper
	Pool               pool.Pooler
	Store              store.Storer
	BackupFormat       string
	DisableCompression bool
}

func onceFromFlags(c *cli.Context) (*once, error) {
	o := &once{}

	var err error
	o.Retriever, err = retrieverFromFlags(c)
	if err != nil {
		return nil, err
	}
	o.Dumper, err = dumperFromFlags(c)
	if err != nil {
		return nil, err
	}
	o.Pool = poolFromFlags(c)
	o.Store = storerFromFlags(c)
	o.BackupFormat = c.GlobalString("backup-format")
	o.DisableCompression = c.GlobalBool("disable-compression")

	return o, nil
}

func (o *once) filename(database string) string {
	name := store.Filename(database, o.BackupFormat)
	if !o.DisableCompression && !strings.HasSuffix(name, ".gz") {
		name = name + ".gz"
	}
	return name
}

func (o *once) Backup(ctx context.Context) error {
	dbs, err := o.Retriever.Retrieve(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve databases")
	}
	if len(dbs) == 0 {
		log.Warn("No databases to backup")
		return nil
	}
	log.WithField("dbs", strings.Join(dbs, ",")).Debug("Backing up databases")

	return o.Pool.Start(ctx, dbs, func(cbCtx context.Context, db string) error {
		filename := o.filename(db)
		log.WithFields(log.Fields{
			"db":       db,
			"filename": filename,
		}).Debug("Starting database backup")

		storeW, err := o.Store.Writer(cbCtx, filename)
		if err != nil {
			return errors.Wrap(err, "failed to get main writer")
		}

		var wErr error
		if o.DisableCompression {
			wErr = o.Dumper.Dump(cbCtx, db, storeW)
		} else {
			gzW := gzip.NewWriter(storeW)
			wErr = o.Dumper.Dump(cbCtx, db, gzW)
			if err := gzW.Close(); err != nil {
				return errors.Wrap(err, "failed to close gzip writer")
			}
		}

		if err := storeW.Close(); err != nil {
			return errors.Wrap(err, "failed to close main writer")
		}
		if wErr == nil {
			log.WithField("db", db).Debug("Database backup complete")
		} else {
			wErr = errors.Wrap(wErr, "dumping failed")
		}

		return wErr
	})
}
