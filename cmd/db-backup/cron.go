package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/utilitywarehouse/go-operational/op"
)

var (
	lastBackupSuccessful = true

	errorsSeen = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "db_backup",
		Name:      "errors_seen",
		Help:      "Count of errors seen",
	})
	retryAttempted = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "db_backup",
		Name:      "retries_attempted",
		Help:      "Count of retries after a failed backup",
	})
	backupTimer = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "db_backup",
		Name:      "backup_timer",
		Help:      "Time taken to run backup",
	})
	databaseBackupFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "db_backup",
		Name:      "database_backup_failed",
		Help:      "Count of failed database backups to storage",
	})
	databaseBackupSuccessful = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "db_backup",
		Name:      "database_backup_successful",
		Help:      "Count of successful database backups to storage",
	})
)

// CronCmd contains the relevant information to schedule a backup
type CronCmd struct {
	schedule        string
	backoffStrategy backoff.BackOff
	once            *once
}

func (cmd *CronCmd) setup(c *cli.Context) error {
	if c.String("schedule") == "" {
		return errors.New("Missing cron schedule")
	}
	cmd.schedule = c.String("schedule")

	if c.Int("retries") > 0 {
		cmd.backoffStrategy = backoff.WithMaxRetries(backoff.NewConstantBackOff(c.Duration("retry-backoff")), uint64(c.Int("retries")))
		log.WithFields(log.Fields{
			"retries": c.Int("retries"),
			"backoff": c.Duration("retry-backoff"),
		}).Debug("Backup attempts will be retried")
	} else {
		cmd.backoffStrategy = &backoff.StopBackOff{}
		log.Debug("Backup retries disabled")
	}

	o, err := onceFromFlags(c)
	if err != nil {
		return err
	}
	cmd.once = o

	return nil
}

// Run executes a Cron job
func (cmd *CronCmd) Run(c *cli.Context) error {
	if err := cmd.setup(c); err != nil {
		return err
	}

	log.WithField("schedule", cmd.schedule).Info("Starting backup with schedule")
	cmd.startOpListener(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sCh := make(chan os.Signal)
		signal.Notify(sCh, os.Interrupt, syscall.SIGTERM)
		<-sCh

		log.Info("Shutdown requested")
		cancel()
	}()

	errCh := make(chan error)
	successCh := make(chan struct{})
	backoffStrategy := backoff.WithContext(cmd.backoffStrategy, ctx)

	cr := cron.New()
	err := cr.AddFunc(cmd.schedule, func() {
		log.Debug("Starting backup attempt")

		backupCb := func() error {
			timer := prometheus.NewTimer(prometheus.ObserverFunc(backupTimer.Set))
			defer timer.ObserveDuration()
			return cmd.once.Backup(ctx)
		}
		errCb := func(err error, duration time.Duration) {
			errCh <- err
			retryAttempted.Inc()
		}

		if err := backoff.RetryNotify(backupCb, backoffStrategy, errCb); err != nil {
			errCh <- errors.Wrapf(err, "backup attempts exhausted")
			log.Warn("Failed to run backup")
		} else {
			successCh <- struct{}{}
		}

		log.WithField("next", cr.Entries()[0].Next).Info("Next scheduled run")
	})
	if err != nil {
		return err
	}
	defer cr.Stop()
	cr.Start()
	log.WithField("next", cr.Entries()[0].Next).Info("Next scheduled run")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			lastBackupSuccessful = false
			errorsSeen.Inc()
			databaseBackupFailed.Inc()
			log.Error(err)
		case <-successCh:
			lastBackupSuccessful = true
			databaseBackupSuccessful.Inc()
			log.Info("Backup successful")
		}
	}
}

func (cmd *CronCmd) startOpListener(c *cli.Context) {
	http.Handle("/__/", op.NewHandler(
		op.NewStatus(c.App.Name, c.App.Usage).
			AddOwner("partner@uw", "#partner-platform").
			AddOwner("telecom", "#telecom-support").
			SetRevision(c.App.Version).
			ReadyAlways().
			WithInstrumentedChecks().
			AddMetrics(
				errorsSeen,
				retryAttempted,
				backupTimer,
				databaseBackupFailed,
				databaseBackupSuccessful,
			).
			AddChecker("db-connection", cmd.dbHealthCheck(c)).
			AddChecker("last-backup-successful", func(cr *op.CheckResponse) {
				if !lastBackupSuccessful {
					cr.Degraded("Last backup attempt failed", "Verify db is running & wait until next schedule backup")
					return
				}
				cr.Healthy("Last backup successful")
			}),
	))

	go func() {
		log.Infof("Operational server started on port %v", c.Int("operational-port"))
		http.ListenAndServe(fmt.Sprintf(":%v", c.Int("operational-port")), nil)
	}()
}

func (cmd *CronCmd) dbHealthCheck(c *cli.Context) func(cr *op.CheckResponse) {
	return func(cr *op.CheckResponse) {
		if err := cmd.once.Dumper.Validate(); err != nil {
			cr.Unhealthy(err.Error(), "Check db is running", "Database backups are not running")
			return
		}

		cr.Healthy("Connected to db")
	}
}
