# sql-backup

[![Build Status](https://drone.prod.merit.uw.systems/api/badges/utilitywarehouse/sql-backup/status.svg)](https://drone.prod.merit.uw.systems/utilitywarehouse/sql-backup)

Backs up pgsql based db's, currently supports postgres and cockroach. Default values are for cockroach.
This started as a fork of [cockroach-backup](https://github.com/utilitywarehouse/cockroach-backup)

### cockroach [deprecated]
`sql-backup --dbcli-binary "cockroach" --dbcli-dsn "root@localhost:26257/system?sslmode=disable" once`

**Deprecated:** As of CockroachDB v20.2 the
[built-in backup feature was made available to free users](https://www.cockroachlabs.com/blog/distributed-backup-restore/).
As such, it is now recommended users make use of this, instead of `sql-backup`.
The [BACKUP](https://www.cockroachlabs.com/docs/stable/backup.html) and
[CREATE SCHEDULE FOR BACKUP](https://www.cockroachlabs.com/docs/v20.2/create-schedule-for-backup.html)
statements can be used to create ad-hoc and scheduled backups respectively.

### postgres
If using postgres need to set shell var PGPASSWORD

`sql-backup --dbcli-binary "pg_dump" --dbcli-dsn "postgres@localhost:5432/postgres?sslmode=disable" once`

## Scheduling

The --schedule flag or SCHEDULE var can be set in the following formats

#### Every hour on the half hour"
`"0 30 * * * *"`
#### Every hour
`"@hourly"`
#### Every hour thrity
`"@every 1h30m"`
