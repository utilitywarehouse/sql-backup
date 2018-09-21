package store

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/go-cloud/blob/s3blob"
)

type Storer interface {
	Writer(ctx context.Context, filename string) (io.WriteCloser, error)
}

func Filename(database, format string) string {
	return fmt.Sprintf(time.Now().Format(format), database)
}

type File struct {
	Dir string
}

func (s File) Writer(ctx context.Context, filename string) (io.WriteCloser, error) {
	if s.Dir == "" {
		s.Dir = "./"
	}
	if !strings.HasSuffix(s.Dir, "/") {
		s.Dir = s.Dir + "/"
	}
	return os.Create(s.Dir + filename)
}

type S3 struct {
	Bucket string
}

func (s S3) Writer(ctx context.Context, filename string) (io.WriteCloser, error) {
	sess := session.Must(session.NewSession())
	bucket, err := s3blob.OpenBucket(ctx, sess, s.Bucket)
	if err != nil {
		return nil, err
	}

	w, err := bucket.NewWriter(ctx, filename, nil)
	if err != nil {
		return nil, err
	}
	return w, nil
}
