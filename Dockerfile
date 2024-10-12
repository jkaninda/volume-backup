FROM golang:1.21.0 AS build
WORKDIR /app

# Copy the source code.
COPY . .
# Installs Go dependencies
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/volume-backup

FROM alpine:3.20.3
ENV STORAGE=local
ENV AWS_S3_ENDPOINT=""
ENV AWS_S3_BUCKET_NAME=""
ENV AWS_ACCESS_KEY=""
ENV AWS_SECRET_KEY=""
ENV AWS_S3_PATH=""
ENV AWS_REGION="us-west-2"
ENV AWS_DISABLE_SSL="false"
ENV GPG_PASSPHRASE=""
ENV SSH_USER=""
ENV SSH_PASSWORD=""
ENV SSH_HOST=""
ENV SSH_IDENTIFY_FILE=""
ENV SSH_PORT=22
ENV REMOTE_PATH=""
ENV FTP_HOST=""
ENV FTP_PORT=21
ENV FTP_USER=""
ENV FTP_PASSWORD=""
ENV BACKUP_CRON_EXPRESSION=""
ENV TG_TOKEN=""
ENV TG_CHAT_ID=""
ENV TZ=UTC
ENV GNUPGHOME="/config/gnupg"
ARG WORKDIR="/config"
ARG BACKUPDIR="/backup"
ARG TEMPLATES_DIR="/config/templates"
ARG DATADIR="/data"
ARG BACKUP_TMP_DIR="/tmp/backup"
ENV VERSION=${appVersion}
LABEL author="Jonas Kaninda"
LABEL version=${appVersion}

RUN apk --update add --no-cache ca-certificates tzdata
RUN mkdir $WORKDIR
RUN mkdir $BACKUPDIR
RUN mkdir $DATADIR
RUN mkdir -p $BACKUP_TMP_DIR
RUN mkdir -p $TEMPLATES_DIR
RUN chmod 777 $WORKDIR
RUN chmod 777 $BACKUPDIR
RUN chmod 777 $BACKUP_TMP_DIR
RUN chmod 777 $DATADIR
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

