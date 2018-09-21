# db-backup

Backs up pgsql based db's, currently supports postgres and cockroach. Default values are for cockroach.


### cockroach
  db-backup once


### postgres
If using postgres need to set shell var PGPASSWORD

  db-backup --dbcli-binary "pg_dump" --dbcli-dsn "postgres@localhost:5432/postgres?sslmode=disable" once

## Scheduling

The --schedule flag or SCHEDULE var can be set in the following formats

`"0 30 * * * *"` - Every hour on the half hour"
`"@hourly"` - "Every hour"
`"@every 1h30m"` - Every hour thirty"
