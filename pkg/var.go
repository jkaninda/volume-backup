// Package pkg /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package pkg

const tmpPath = "/tmp/backup"
const gpgHome = "/config/gnupg"
const algorithm = "aes256"
const gpgExtension = "gpg"
const dataPath = "/data"
const backupDestination = "/backup"

var (
	storage    = "local"
	file       = ""
	encryption = false
)

var tdbRVars = []string{
	"TARGET_DB_HOST",
	"TARGET_DB_PORT",
	"TARGET_DB_NAME",
	"TARGET_DB_USERNAME",
	"TARGET_DB_PASSWORD",
}

// sshVars Required environment variables for SSH remote server storage
var sshVars = []string{
	"SSH_USER",
	"SSH_HOST",
	"SSH_PORT",
	"REMOTE_PATH",
}
var ftpVars = []string{
	"FTP_HOST",
	"FTP_USER",
	"FTP_PASSWORD",
	"FTP_PORT",
}

// AwsVars Required environment variables for AWS S3 storage
var awsVars = []string{
	"AWS_S3_ENDPOINT",
	"AWS_S3_BUCKET_NAME",
	"AWS_ACCESS_KEY",
	"AWS_SECRET_KEY",
	"AWS_REGION",
}
