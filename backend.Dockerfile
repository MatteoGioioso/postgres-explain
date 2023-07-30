FROM ubuntu:20.04 AS base
LABEL maintainer="Matteo Gioioso <gioioso.matteo@gmail.com>"

ARG USER
ARG GROUP
ARG UID=1001
ARG GID=1001
ARG WALG_VERSION=v2.0.1

ENV GIN_MODE=release
ENV USER=$USER
ENV GROUP=$GROUP
ENV UID=$UID
ENV GID=$GID

RUN apt-get update && apt-get upgrade -y

RUN addgroup $GROUP
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "$(pwd)" \
    --ingroup "$GROUP" \
    --no-create-home \
    --uid "$UID" \
    "$USER"

FROM base

ADD backend/bin/ /

# Clean up
RUN apt-get autoremove --purge && apt-get clean

USER $USER

ENTRYPOINT ["/backend"]