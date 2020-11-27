# sql-backup

Backs up pgsql based db's, currently supports postgres and cockroach. Default values are for cockroach.
This started as a fork of [cockroach-backup](https://github.com/utilitywarehouse/cockroach-backup)


### cockroach
`sql-backup --dbcli-binary "cockroach" --dbcli-dsn "root@localhost:26257/system?sslmode=disable" once`

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
