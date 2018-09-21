package main

import (
	"context"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	appName        = "db-backup"
	appDescription = "Backup up db databases. By default all databases are backed up."
)

var (
	gitSummary = "replaced by `make build`"
	gitBranch  = "replaced by `make build`"
	buildStamp = "replaced by `make build`"
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = appDescription
	app.Description = ""
	app.Version = fmt.Sprintf("%v-%v (%v)", gitSummary, gitBranch, buildStamp)

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "log",
			Usage:  "Set the log level",
			EnvVar: "LOG",
			Value:  "info",
			Hidden: true,
		},
		cli.StringFlag{
			Name:   "backup-format",
			Usage:  "Filename format. Passed through time.Format & fmt.Sprintf",
			EnvVar: "BACKUP_FORMAT",
			Value:  "%s_2006-01-02_150405.sql",
		},
		cli.StringFlag{
			Name:   "dbcli-binary",
			Usage:  "Path to the db cli binary, eg. cockroach or pg_dump",
			EnvVar: "DBCLI_PATH",
			Value:  "cockroach",
		},
		cli.StringFlag{
			Name:   "dbcli-flags",
			Usage:  "Flags to pass to db",
			EnvVar: "DBCLI_FLAGS",
			Value:  "--insecure",
		},
		cli.StringFlag{
			Name:   "dbcli-dsn",
			Usage:  "db connection DSN",
			EnvVar: "DBCLI_DSN",
			Value:  "root@localhost:26257/system?sslmode=disable",
		},
		cli.DurationFlag{
			Name:   "dbcli-timeout",
			Usage:  "Timeout when calling `db dump`",
			EnvVar: "DBCLI_TIMEOUT",
		},
		cli.IntFlag{
			Name:   "pool",
			Usage:  "Number of databases to concurrently dump",
			EnvVar: "POOL",
			Value:  5,
			Hidden: true,
		},
		cli.BoolFlag{
			Name:   "disable-compression",
			Usage:  "Disable compressing the backed up SQL file",
			EnvVar: "DISABLE_COMPRESSION",
		},
		cli.StringSliceFlag{
			Name:   "only",
			Usage:  "Comma-separated list of databses to backup. If not provided, all are backed up",
			EnvVar: "DBS",
		},
		cli.StringSliceFlag{
			Name:   "exclude",
			Usage:  "Comma-separated list of databses to filter when backing up all",
			EnvVar: "FILTER_DBS",
		},
		cli.StringFlag{
			Name:   "driver",
			Usage:  "Storage driver. One of 'file' or 'aws'",
			EnvVar: "DRIVER",
			Value:  "file",
		},
		cli.StringFlag{
			Name:   "dir",
			Usage:  "Directory path to store backups in. For driver 'file'",
			EnvVar: "BACKUP_DIR",
			Value:  "./",
		},
		cli.StringFlag{
			Name:   "bucket",
			Usage:  "Name of the AWS bucket to upload files into. For driver 'aws'",
			EnvVar: "BACKUP_BUCKET",
		},
	}
	app.Before = func(c *cli.Context) error {
		lvl, err := log.ParseLevel(c.GlobalString("log"))
		if err != nil {
			return err
		}
		log.SetLevel(lvl)
		return nil
	}

	app.Commands = cli.Commands{
		cli.Command{
			Name:  "once",
			Usage: "Backup databases once and then stop.",
			Action: func(c *cli.Context) error {
				log.Info("Performing backup once...")
				cmd := &OnceCmd{}
				if err := cmd.Run(c); err != nil {
					if err != context.Canceled {
						return err
					}
				}
				log.Info("Backup complete")
				return nil
			},
		},
		cli.Command{
			Name:  "cron",
			Usage: "Backup databases on a cron loop.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "schedule",
					Usage:  "Cron schedule to perform backups on",
					EnvVar: "SCHEDULE",
				},
				cli.IntFlag{
					Name:   "retries",
					Usage:  "Number of times to retry on failure. After this the process will exit",
					EnvVar: "RETRIES",
					Value:  1,
				},
				cli.DurationFlag{
					Name:   "retry-backoff",
					Usage:  "Delay in-between retries",
					EnvVar: "RETRY_BACKOFF",
					Value:  1 * time.Minute,
				},
				cli.StringFlag{
					Name:   "opsgenie-heartbeat-key",
					Usage:  "Opsgene API key",
					EnvVar: "OPSGENIE_HEARTBEAT_KEY",
				},
				cli.StringFlag{
					Name:   "opsgenie-heartbeat-name",
					Usage:  "Opsgenie heartbeat name",
					EnvVar: "OPSGENIE_HEARTBEAT_NAME",
				},
				cli.IntFlag{
					Name:   "operational-port",
					Usage:  "Port to serve HTTP operational endpoints on",
					EnvVar: "OPERATIONAL_PORT",
					Value:  8080,
				},
			},
			Action: func(c *cli.Context) error {
				cmd := &CronCmd{}
				if err := cmd.Run(c); err != nil {
					if err != context.Canceled {
						return err
					}
				}
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
