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
// TODO confirm equivilent for pg_dump
func (d CliDumper) Validate() error {
	nodeCmd := exec.Command(d.Cmd, "node", "ls", d.Flags)
	if err := nodeCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to validate db connection")
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
		if !strings.HasPrefix(d.DSN, "postgresql://") {
			d.DSN = "postgresql://" + d.DSN
		}
		u, err := url.Parse(d.DSN)
		if err != nil {
			return errors.New("invalid DSN")
		}

		dumpCmd = exec.Command(d.Cmd, "-d", db, "-h", u.Hostname(), "-U", u.User.Username())
	default:
		return errors.New("unknown command")
	}

	buf := bufio.NewWriter(w)
	dumpCmd.Stdout = buf

	if err := dumpCmd.Start(); err != nil {
		return err
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
		return ctx.Err()
	case <-timeoutCh:
		dumpCmd.Process.Kill()
		return fmt.Errorf("timed out dumping database: %s", db)
	case err := <-doneCh:
		if err != nil {
			return err
		}
		return buf.Flush()
	}
}
