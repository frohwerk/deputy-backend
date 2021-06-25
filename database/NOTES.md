Neues "Projekt" bzw. neuen Namespace anlegen:
1. Login as admin (or any user allowed to create new namespaces)
> oc new-project deputy
> oc policy add-role-to-user <role> other-user -n deputy

2. Image Registry Credentials-Secret anlegen für registry.redhat.io
> Open Openshift Console at path /console/project/deputy/browse/secrets
> Create new Secret (Type: Image Secret)
- Name: registry.redhat.io
- Image Registry Server Address: registry.redhat.io
- Username: YOUR USERNAME
- E-Mail: YOUR EMAIL
- Password: YOUR PASSWORD
--- Note: You need to create a Red Hat account first

3. Deploy postgresql database
> oc apply -f database\deployments\minishift\postgres.volume.yml
--- Note the necessity to find an available volume using the "oc get pv" command
> oc apply -f database\deployments\minishift\postgres.deployment.yml
--- Warn about the hard-coded password and the need to change it for a production environment

4. Expose database on localhost
> oc port-forward deployment/deputy-database LOCAL_PORT:5432

5. Connect as postgres (admin) user and install extension
SQL> CREATE EXTENSION IF NOT EXISTS 'uuid-ossp';

6. Connect as regular user and create deputy database objects



---------------------------------------------------------------------------------------------------------------------------------------------------

POSTGRESQL_USER
User name for PostgreSQL account to be created

POSTGRESQL_PASSWORD
Password for the user account

POSTGRESQL_DATABASE
Database name

---------------------------------------------------------------------------------------------------------------------------------------------------

Uploading custom images to openshift container registry:
https://docs.openshift.com/container-platform/3.11/install_config/registry/accessing_registry.html#access-pushing-and-pulling-images

---------------------------------------------------------------------------------------------------------------------------------------------------

environment variables:
PGPASSWORD
PGPASSFILE

https://www.postgresql.org/docs/current/libpq-envars.html
https://www.postgresql.org/docs/current/libpq-pgpass.html

---------------------------------------------------------------------------------------------------------------------------------------------------

PGPASSFILE=/whatever/password psql -d deputy -U deputy -f my-sql-script.sql

---------------------------------------------------------------------------------------------------------------------------------------------------

sh-4.4$ psql --help
psql is the PostgreSQL interactive terminal.

Usage:
  psql [OPTION]... [DBNAME [USERNAME]]

General options:
  -c, --command=COMMAND    run only single command (SQL or internal) and exit
  -d, --dbname=DBNAME      database name to connect to (default: "postgres")
  -f, --file=FILENAME      execute commands from file, then exit
  -l, --list               list available databases, then exit
  -v, --set=, --variable=NAME=VALUE
                           set psql variable NAME to VALUE
                           (e.g., -v ON_ERROR_STOP=1)
  -V, --version            output version information, then exit
  -X, --no-psqlrc          do not read startup file (~/.psqlrc)
  -1 ("one"), --single-transaction
                           execute as a single transaction (if non-interactive)
  -?, --help[=options]     show this help, then exit
      --help=commands      list backslash commands, then exit
      --help=variables     list special variables, then exit

Input and output options:
  -a, --echo-all           echo all input from script
  -b, --echo-errors        echo failed commands
  -e, --echo-queries       echo commands sent to server
  -E, --echo-hidden        display queries that internal commands generate
  -L, --log-file=FILENAME  send session log to file
  -n, --no-readline        disable enhanced command line editing (readline)
  -o, --output=FILENAME    send query results to file (or |pipe)
  -q, --quiet              run quietly (no messages, only query output)
  -s, --single-step        single-step mode (confirm each query)
  -S, --single-line        single-line mode (end of line terminates SQL command)

Output format options:
  -A, --no-align           unaligned table output mode
      --csv                CSV (Comma-Separated Values) table output mode
  -F, --field-separator=STRING
                           field separator for unaligned output (default: "|")
  -H, --html               HTML table output mode
  -P, --pset=VAR[=ARG]     set printing option VAR to ARG (see \pset command)
  -R, --record-separator=STRING
                           record separator for unaligned output (default: newline)
  -t, --tuples-only        print rows only
  -T, --table-attr=TEXT    set HTML table tag attributes (e.g., width, border)
  -x, --expanded           turn on expanded table output
  -z, --field-separator-zero
                           set field separator for unaligned output to zero byte
  -0, --record-separator-zero
                           set record separator for unaligned output to zero byte

Connection options:
  -h, --host=HOSTNAME      database server host or socket directory (default: "local socket")
  -p, --port=PORT          database server port (default: "5432")
  -U, --username=USERNAME  database user name (default: "postgres")
  -w, --no-password        never prompt for password
  -W, --password           force password prompt (should happen automatically)

For more information, type "\?" (for internal commands) or "\help" (for SQL
commands) from within psql, or consult the psql section in the PostgreSQL
documentation.

Report bugs to <pgsql-bugs@lists.postgresql.org>.
