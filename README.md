# Docker Volume data backup

Docker volume data backup to local, AWS S3, FTP or SSH remote server storage

[![Build](https://github.com/jkaninda/volume-backup/actions/workflows/release.yml/badge.svg)](https://github.com/jkaninda/volume-backup/actions/workflows/release.yml)
![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/jkaninda/volume-backup?style=flat-square)
![Docker Pulls](https://img.shields.io/docker/pulls/jkaninda/volume-backup?style=flat-square)

Mounting paths:
 - /data: volume data
 - /backup: Backup destination for local storage backup

## Storage:
- Local
- AWS S3 or any S3 Alternatives for Object Storage
- FTP remote server
- SSH remote server

## Quickstart

## Backup
### Simple backup using Docker CLI

```shell
docker run --rm  --name volume-backup \
-v "data:/data" \
-v "./backup:/backup" \
 jkaninda/volume-backup backup --cron-expression "@every 20m"
```
### Recurring backup

```shell
docker run --rm  --name volume-backup \
-v "data:/data" \
-v "./backup:/backup" \
 jkaninda/volume-backup backup --cron-expression "@every 15m"
```
#### Backup a single file

```shell
docker run --rm  --name volume-backup \
-v "data:/data" \
-v "./backup:/backup" \
jkaninda/volume-backup backup --file my-file-inside-container.json
```
### Backup using AWS S3 object storage

```env
AWS_ACCESS_KEY=
AWS_SECRET_KEY=
AWS_S3_BUCKET_NAME=
AWS_S3_ENDPOINT=http://minio:9000
AWS_DISABLE_SSL=false
AWS_REGION=eu
AWS_S3_PATH=/volume-backup
BACKUP_PREFIX=backup
```
```shell
docker run --rm  --name volume-backup \
--env-file env \
jkaninda/volume-backup backup --storage s3 --cron-expression "@midnight"
```

### Backup using SSH remote server storage

```env
SSH_HOST=192.168.1.44
SSH_USER=toto
SSH_PASSWORD=password
SSH_IDENTIFY_FILE=/config/id_ed25519
SSH_PORT=22
REMOTE_PATH=/home/toto/backup
BACKUP_PREFIX=backup
```
```shell
docker run --rm  --name volume-backup \
-v "data:/data" \
--env-file env \
jkaninda/volume-backup backup --storage ssh --cron-expression "@midnight"
```

### Backup using FTP remote server storage

```env
FTP_HOST=192.168.1.44
FTP_USER=toto
FTP_PASSWORD=password
FTP_PORT=21
REMOTE_PATH=/ftp/jonas/toto
TZ=Europe/Paris
```
```shell
docker run --rm  --name volume-backup \
-v "data:/data" \
--env-file env \
jkaninda/volume-backup backup --storage ftp --cron-expression "@midnight"
```
## Restore a backup

### Restore from AWS S3 object storage

```env
AWS_ACCESS_KEY=
AWS_SECRET_KEY=
AWS_S3_BUCKET_NAME=
AWS_S3_ENDPOINT=http://192.168.1.30:9000
AWS_DISABLE_SSL=false
AWS_REGION=eu
AWS_S3_PATH=/volume-backup
BACKUP_PREFIX=backup
```
```shell
docker run --rm  --name volume-backup \
-v "data:/data" \
--env-file env \
jkaninda/volume-backup restore --storage s3 --file backup_20241001_112322.tar.gz
```


```shell
docker run --rm  --name volume-backup \
-v "data:/data" \
--env-file env \
jkaninda/volume-backup restore --storage s3 --file backup_20241001_112322.tar
```
## Encrypt backup
To encrypt and decrypt your backup, you need to set `GPG_PASSPHRASE` environment variable

```shell
docker run --rm  --name volume-backup \
-v "data:/data" \
-v "./backup:/backup" \
-e "GPG_PASSPHRASE=passphrase" \
 jkaninda/volume-backup backup --cron-expression "@every 20m"
```

## Backup notification

### Telegram notification

```shell
TG_TOKEN=Telegram token (`BOT-ID:BOT-TOKEN`)
TG_CHAT_ID=
```
### Email notification

```shell
MAIL_HOST=
MAIL_PORT=587
MAIL_USERNAME=
MAIL_PASSWORD=!
MAIL_FROM=Backup Jobs <backup@example.com>
#Multiple recipients separated by a comma
MAIL_TO=me@example.com,team@example.com,manager@example.com
MAIL_SKIP_TLS=false
# Backup time format for notification 
TIME_FORMAT=2006-01-02 at 15:04:05
#Backup reference, in case you want to identify every backup instance
BACKUP_REFERENCE=docker/Paris cluster
```


---
## Run in Scheduled mode

This image can be run as CronJob in Kubernetes for a regular backup which makes deployment on Kubernetes easy as Kubernetes has CronJob resources.
For Docker, you need to run it in scheduled mode by adding `--cron-expression  "* * * * *"` flag or by defining `BACKUP_CRON_EXPRESSION=0 1 * * *` environment variable.

## Syntax of crontab (field description)

The syntax is:

- 1: Minute (0-59)
- 2: Hours (0-23)
- 3: Day (0-31)
- 4: Month (0-12 [12 == December])
- 5: Day of the week(0-7 [7 or 0 == sunday])

Easy to remember format:

```conf
* * * * * command to be executed
```

```conf
- - - - -
| | | | |
| | | | ----- Day of week (0 - 7) (Sunday=0 or 7)
| | | ------- Month (1 - 12)
| | --------- Day of month (1 - 31)
| ----------- Hour (0 - 23)
------------- Minute (0 - 59)
```

> At every 30th minute

```conf
*/30 * * * *
```
> “At minute 0.” every hour
```conf
0 * * * *
```

> “At 01:00.” every day

```conf
0 1 * * *
```
## Predefined schedules
You may use one of several pre-defined schedules in place of a cron expression.

| Entry                  | Description                                | Equivalent To |
|------------------------|--------------------------------------------|---------------|
| @yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 1 1 *     |
| @monthly               | Run once a month, midnight, first of month | 0 0 1 * *     |
| @weekly                | Run once a week, midnight between Sat/Sun  | 0 0 * * 0     |
| @daily (or @midnight)  | Run once a day, midnight                   | 0 0 * * *     |
| @hourly                | Run once an hour, beginning of hour        | 0 * * * *     |

### Intervals
You may also schedule backup task at fixed intervals, starting at the time it's added or cron is run. This is supported by formatting the cron spec like this:

@every <duration>
where "duration" is a string accepted by time.

For example, "@every 1h30m10s" would indicate a schedule that activates after 1 hour, 30 minutes, 10 seconds, and then every interval after that.