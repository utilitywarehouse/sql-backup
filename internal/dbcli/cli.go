package dbcli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
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
	cmdPath, _ := exec.LookPath(d.Cmd)

	dumpCmd := exec.Command(cmdPath, "dump", db, d.Flags)
	switch d.Cmd {
	case "cockroach":
		// no action
	case "pg_dump":
		u, err := dsnToURL(d.DSN)
		if err != nil {
			return err
		}
		dumpCmd = exec.Command(d.Cmd, "-d", db, "-h", u.Hostname(), "-U", u.User.Username())
	default:
		return errors.New("unknown dbcli command")
	}

	buf := bufio.NewWriter(w)
	dumpCmd.Stdout = buf

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
		dumpCmd.Process.Kill()
		return errors.Wrap(ctx.Err(), "context was cancelled")
	case <-timeoutCh:
		dumpCmd.Process.Kill()
		return fmt.Errorf("timed out dumping database: %s", db)
	case err := <-doneCh:
		if err != nil {
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
