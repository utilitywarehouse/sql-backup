# db-backup

Backs up pgsql based db's, currently supports postgres and cockroach. Default values are for cockroach.


### cockroach
  db-backup once


### postgres
If using postgres need to set shell var PGPASSWORD

  db-backup --dbcli-binary "pg_dump" --dbcli-dsn "postgres@localhost:5432/postgres?sslmode=disable" once

  
