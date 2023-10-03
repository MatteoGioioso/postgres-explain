FROM postgres:13

COPY bench.sh /bench.sh
COPY bench.sql /bench.sql
RUN chmod +x /bench.sh

ENTRYPOINT ["/bench.sh"]