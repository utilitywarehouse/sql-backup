# sql-backup

[![Build Status](https://drone.prod.merit.uw.systems/api/badges/utilitywarehouse/sql-backup/status.svg)](https://drone.prod.merit.uw.systems/utilitywarehouse/sql-backup)

Backs up pgsql based db's, currently supports postgres. Default values are for pg_dump.

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
