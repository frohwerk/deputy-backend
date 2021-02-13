FROM postgres:12
COPY docker-entrypoint.sh /usr/local/bin/
# IDEA: Remove image provided directories, create in docker-entrypoint.sh
RUN chgrp -R 0 /var/lib/postgresql && \
    chmod -R g+rwX /var/lib/postgresql && \
    mkdir -p /var/lib/postgresql/data && \
    chgrp -R 0 /var/lib/postgresql/data && \
    chmod 700 /var/lib/postgresql/data && \
    mkdir -p /var/lib/postgresql/data && \
    chmod 700 /var/lib/postgresql/data && \
    chgrp 0 /usr/local/bin/docker-entrypoint.sh && \
    chmod g+rx /usr/local/bin/docker-entrypoint.sh
# RUN mkdir -p /var/lib/postgresql
# ENTRYPOINT [ "/bin/bash" ]
