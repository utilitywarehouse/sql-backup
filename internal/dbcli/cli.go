package dbcli

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Dumper is an interface for a Cli Dumper
type Dumper interface {
	Validate() error
	Dump(ctx context.Context, db string, w io.Writer) error
}

// CliDumper contains the required information to use a DB Cli tool to dump a DB.
type CliDumper struct {
	Cmd     string
	Flags   string
	DSN     string
	Timeout time.Duration
}

// NewDumper returns a populated CliDumper
func NewDumper(cmd, flags, dsn string) (CliDumper, error) {
	if _, err := os.Stat(cmd); os.IsNotExist(err) {
		_, lookErr := exec.LookPath(cmd)
		if lookErr != nil {
			return CliDumper{}, errors.Wrapf(lookErr, "failed to find db binary")
		}
	}
	return CliDumper{Cmd: cmd, Flags: flags, DSN: dsn}, nil
}

// Validate checks the Cli connection to DB
func (d CliDumper) Validate() error {
	switch d.Cmd {
	case "cockroach":
		// #nosec G204
		nodeCmd := exec.Command(d.Cmd, "node", "ls", d.Flags)
		if err := nodeCmd.Run(); err != nil {
			return errors.Wrapf(err, "failed to validate db connection")
		}
	case "pg_dump":
		u, err := dsnToURL(d.DSN)
		if err != nil {
			return err
		}
		// Lazy but assuming if pg_dump was found that pg_isready will also work
		// #nosec G204
		pgCmd := exec.Command("pg_isready", "-h", u.Hostname(), "-U", u.User.Username())
		if err := pgCmd.Run(); err != nil {
			return errors.Wrapf(err, "failed to validate db connection")
		}
	default:
		return errors.New("unknown dbcli command")
	}
	return nil
}

// Dump takes a dump of the databases from a DB host
func (d CliDumper) Dump(ctx context.Context, db string, w io.Writer) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Not checking error as was checked in NewDumper
	cmdPath, _ := exec.LookPath(d.Cmd) // nolint:errcheck

	// #nosec G204
	dumpCmd := exec.Command(cmdPath, "dump", db, d.Flags)
	switch d.Cmd {
	case "cockroach":
		// no action
	case "pg_dump":
		u, err := dsnToURL(d.DSN)
		if err != nil {
			return err
		}
		// #nosec G204
		dumpCmd = exec.Command(d.Cmd, "-d", db, "-h", u.Hostname(), "-U", u.User.Username())
		if pass, ok := u.User.Password(); ok {
			dumpCmd.Env = []string{
				fmt.Sprintf("PGPASSWORD=%s", pass),
			}
		}
	default:
		return errors.New("unknown dbcli command")
	}

	buf := bufio.NewWriter(w)
	errBuff := &bytes.Buffer{}
	dumpCmd.Stdout = buf
	dumpCmd.Stderr = errBuff

	if err := dumpCmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start dumper")
	}

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- dumpCmd.Wait()
	}()

	timeoutCh := make(chan struct{})
	if d.Timeout != 0 {
		timeoutTimer := time.AfterFunc(d.Timeout, func() {
			close(timeoutCh)
		})
		defer timeoutTimer.Stop()
	}

	select {
	case <-ctx.Done():
		if err := dumpCmd.Process.Kill(); err != nil {
			<-doneCh // wait for the command to terminate
		}
		return errors.Wrap(ctx.Err(), "context was cancelled")
	case <-timeoutCh:
		if err := dumpCmd.Process.Kill(); err != nil {
			<-doneCh // wait for the command to terminate
		}
		return fmt.Errorf("timed out dumping database: %s", db)
	case err := <-doneCh:
		if err != nil {
			log.WithField("db", db).WithError(err).Error(string(errBuff.Bytes()))
			return errors.Wrap(err, "dumper failed")
		}
		return buf.Flush()
	}
}

func dsnToURL(dsn string) (*url.URL, error) {
	if !strings.HasPrefix(dsn, "postgresql://") {
		dsn = "postgresql://" + dsn
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.New("invalid DSN")
	}
	return u, nil
}
