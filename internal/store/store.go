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

// Storer interface abstracts a Writer func
type Storer interface {
	Writer(ctx context.Context, filename string) (io.WriteCloser, error)
}

// Filename embelishes a output file.
func Filename(database, format string) string {
	return fmt.Sprintf(time.Now().Format(format), database)
}

// File type is used for file based operations
type File struct {
	Dir string
}

// Writer writes a File type.
func (s File) Writer(ctx context.Context, filename string) (io.WriteCloser, error) {
	if s.Dir == "" {
		s.Dir = "./"
	}
	if !strings.HasSuffix(s.Dir, "/") {
		s.Dir = s.Dir + "/"
	}
	return os.Create(s.Dir + filename)
}

// S3 type is used for S3 based opertaions
type S3 struct {
	Bucket string
}

// Writer writes an S3 type.
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
