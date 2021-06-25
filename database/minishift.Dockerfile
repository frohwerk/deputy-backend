FROM registry.redhat.io/rhel8/postgresql-13
COPY docker-entrypoint-initdb.d/*.sql /docker-entrypoint-initdb.d/