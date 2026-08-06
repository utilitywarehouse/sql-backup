package store

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/s3blob"
	"gocloud.dev/gcp"
)

// Storer interface abstracts a Writer func
type Storer interface {
	Writer(ctx context.Context, filename string) (io.WriteCloser, error)
}

// Filename embellishes a output file.
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
	Bucket     string
	Dir        string
	BufferSize int
}

// Writer writes an S3 type.
func (s S3) Writer(ctx context.Context, filename string) (io.WriteCloser, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}
	client := s3.NewFromConfig(cfg)
	bucket, err := s3blob.OpenBucketV2(ctx, client, s.Bucket, nil)
	if err != nil {
		return nil, err
	}

	if s.Dir != "" {
		filename = filepath.Join(s.Dir, filename)
	}

	w, err := bucket.NewWriter(ctx, filename, &blob.WriterOptions{BufferSize: s.BufferSize})
	if err != nil {
		return nil, err
	}
	return w, nil
}

// GCS type is used for GCS storage on GCP
type GCS struct {
	Bucket string
	Dir    string
}

// Writer writes to googe cloud storage
func (g GCS) Writer(ctx context.Context, filename string) (io.WriteCloser, error) {
	creds, err := gcp.DefaultCredentials(ctx)
	if err != nil {
		return nil, err
	}
	c, err := gcp.NewHTTPClient(gcp.DefaultTransport(), gcp.CredentialsTokenSource(creds))
	if err != nil {
		return nil, err
	}

	bucket, err := gcsblob.OpenBucket(ctx, c, g.Bucket, nil)
	if err != nil {
		return nil, err
	}

	if g.Dir != "" {
		filename = filepath.Join(g.Dir, filename)
	}

	w, err := bucket.NewWriter(ctx, filename, nil)
	if err != nil {
		return nil, err
	}
	return w, nil
}
