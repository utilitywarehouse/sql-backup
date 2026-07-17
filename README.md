# sql-backup

[![Build Status](https://drone.prod.merit.uw.systems/api/badges/utilitywarehouse/sql-backup/status.svg)](https://drone.prod.merit.uw.systems/utilitywarehouse/sql-backup)

Backs up pgsql based db's, currently supports postgres. Default values are for pg_dump.

### S3 upload part size

When using `--driver aws`, `--buffer-size` (or `BACKUP_BUFFER_SIZE`) sets the multipart upload part size in bytes. Leave at 0 (default) to use the library default; raise it for large backups to reduce the number of parts.

S3 multipart uploads are capped at 10,000 parts. The AWS SDK's default part size is 5MB, so `5MB * 10,000 = ~50GB` is the largest dump that can upload with the default. Since backups are streamed (size unknown upfront), the SDK cannot auto-grow the part size to compensate the way it does for seekable files, so a dump crossing ~50GB fails with "too many parts". Raise `--buffer-size` past `dump_size / 10,000` to avoid this, e.g. `104857600` (100MB) supports dumps up to ~977GB.

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
