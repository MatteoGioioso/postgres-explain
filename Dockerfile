FROM ubuntu:20.04 AS base
LABEL maintainer="Matteo Gioioso <info@matteogioioso.com>"

ARG USER
ARG GROUP
ARG UID=1001
ARG GID=1001
ARG NGINX_VERSION=1.23.3-1~focal

ENV USER=$USER
ENV GROUP=$GROUP
ENV UID=$UID
ENV GID=$GID
ENV BOREALIS_DIR=$BOREALIS_DIR
ENV NGINX_VERSION=$NGINX_VERSION
ENV CLICKHOUSE_VERSION=$CLICKHOUSE_VERSION

RUN DEBIAN_FRONTEND=noninteractive \
    && apt-get update && apt-get upgrade -y \
    && apt-get install -y ca-certificates runit sqlite3 software-properties-common wget apt-transport-https dumb-init

RUN addgroup $GROUP
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "$(pwd)" \
    --ingroup "$GROUP" \
    --no-create-home \
    --uid "$UID" \
    "$USER"

FROM base AS dependencies

COPY scripts/launch.sh /borealis/launch.sh

# runit
COPY scripts/runit /borealis/services/
RUN for d in /borealis/services/*; do \
        chmod 755 $d/* \
        && ln -s /borealis/services/$(basename $d) /etc/service/; \
    done

# nginx
RUN DEBIAN_FRONTEND=noninteractive  \
    && wget -q -O - https://nginx.org/keys/nginx_signing.key | gpg --dearmor | tee /usr/share/keyrings/nginx-archive-keyring.gpg >/dev/null \
    && echo "deb [signed-by=/usr/share/keyrings/nginx-archive-keyring.gpg] http://nginx.org/packages/mainline/ubuntu `lsb_release -cs` nginx" | tee /etc/apt/sources.list.d/nginx.list \
    && apt-get update \
    && apt-get install -y nginx=$NGINX_VERSION