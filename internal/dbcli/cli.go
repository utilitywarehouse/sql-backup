package dbcli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
)

type Dumper interface {
	Validate() error
	Dump(ctx context.Context, db string, w io.Writer) error
}

type CliDumper struct {
	Cmd     string
	Flags   string
	DSN     string
	Timeout time.Duration
}

func NewDumper(cmd, flags, dsn string) (CliDumper, error) {
	if _, err := os.Stat(cmd); os.IsNotExist(err) {
		_, lookErr := exec.LookPath(cmd)
		if lookErr != nil {
			return CliDumper{}, errors.Wrapf(lookErr, "failed to find db binary")
		}
	}
	return CliDumper{Cmd: cmd, Flags: flags, DSN: dsn}, nil
}

func (d CliDumper) Validate() error {
	nodeCmd := exec.Command(d.Cmd, "node", "ls", d.Flags)
	if err := nodeCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to validate db connection")
	}
	return nil
}

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
		dumpCmd = exec.Command(cmdPath, "-d", db, d.Flags)
		//dumpCmd = exec.Command(d.Cmd, "-d", db, "-h", "localhost", "-U", "postgres")
	default:
		return errors.New("unknown command")
	}

	fmt.Println(dumpCmd)

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
