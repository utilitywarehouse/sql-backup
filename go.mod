module github.com/utilitywarehouse/sql-backup

go 1.15

require (
	github.com/aws/aws-sdk-go v1.44.68
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/lib/pq v1.10.6
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.2
	github.com/robfig/cron v1.2.0
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.8.0
	github.com/urfave/cli v1.22.5
	github.com/utilitywarehouse/go-operational v0.0.0-20190722153447-b0f3f6284543
	gocloud.dev v0.27.0
)
