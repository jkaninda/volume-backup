FROM golang:1.24.4 AS build
WORKDIR /app

# Copy the source code.
COPY . .
# Installs Go dependencies
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/volume-backup

FROM alpine:3.21.3
ENV TZ=UTC
ARG WORKDIR="/config"
ARG BACKUPDIR="/backup"
ARG BACKUP_TMP_DIR="/tmp/backup"
ARG TEMPLATES_DIR="/config/templates"
ARG appVersion=""
ENV VERSION=${appVersion}
LABEL author="Jonas Kaninda"
LABEL version=${appVersion}
LABEL github="github.com/jkaninda/volume-backup"

RUN apk --update add --no-cache tzdata ca-certificates
RUN mkdir -p $WORKDIR $BACKUPDIR $TEMPLATES_DIR $BACKUP_TMP_DIR && \
     chmod a+rw $WORKDIR $BACKUPDIR $BACKUP_TMP_DIR
COPY --from=build /app/volume-backup /usr/local/bin/volume-backup
COPY ./templates/* $TEMPLATES_DIR/
RUN chmod +x /usr/local/bin/volume-backup

RUN ln -s /usr/local/bin/volume-backup /usr/local/bin/bkup

# Create the data script and make it executable
RUN printf '#!/bin/sh\n/usr/local/bin/volume-backup backup "$@"' > /usr/local/bin/backup && \
    chmod +x /usr/local/bin/backup

# Create the restore script and make it executable
RUN printf '#!/bin/sh\n/usr/local/bin/volume-backup restore "$@"' > /usr/local/bin/restore && \
    chmod +x /usr/local/bin/restore
WORKDIR $WORKDIR
ENTRYPOINT ["/usr/local/bin/volume-backup"]

