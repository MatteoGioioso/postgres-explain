FROM postgres:13

COPY bench.sh /bench.sh
RUN chmod +x /bench.sh

ENTRYPOINT ["/bench.sh"]