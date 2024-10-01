// Package pkg /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package pkg

import (
	"fmt"
	"github.com/jkaninda/volume-backup/utils"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

type TgConfig struct {
	Token  string
	ChatId string
}
type BackupConfig struct {
	backupFileName     string
	backupRetention    int
	disableCompression bool
	prune              bool
	encryption         bool
	remotePath         string
	file               string
	passphrase         string
	storage            string
	cronExpression     string
	prefix             string
	fromFolder         bool
}
type FTPConfig struct {
	host       string
	user       string
	password   string
	port       string
	remotePath string
}

// SSHConfig holds the SSH connection details
type SSHConfig struct {
	user         string
	password     string
	hostName     string
	port         string
	identifyFile string
}
type AWSConfig struct {
	endpoint       string
	bucket         string
	accessKey      string
	secretKey      string
	region         string
	disableSsl     bool
	forcePathStyle bool
}

// loadSSHConfig loads the SSH configuration from environment variables
func loadSSHConfig() (*SSHConfig, error) {
	sshVars := []string{"SSH_USER", "SSH_HOST", "SSH_PORT", "REMOTE_PATH"}
	err := utils.CheckEnvVars(sshVars)
	if err != nil {
		return nil, fmt.Errorf("error missing environment variables: %w", err)
	}

	return &SSHConfig{
		user:         os.Getenv("SSH_USER"),
		password:     os.Getenv("SSH_PASSWORD"),
		hostName:     os.Getenv("SSH_HOST"),
		port:         os.Getenv("SSH_PORT"),
		identifyFile: os.Getenv("SSH_IDENTIFY_FILE"),
	}, nil
}
func initFtpConfig() *FTPConfig {
	//Initialize data configs
	fConfig := FTPConfig{}
	fConfig.host = os.Getenv("FTP_HOST")
	fConfig.user = os.Getenv("FTP_USER")
	fConfig.password = os.Getenv("FTP_PASSWORD")
	fConfig.port = os.Getenv("FTP_PORT")
	fConfig.remotePath = os.Getenv("REMOTE_PATH")
	err := utils.CheckEnvVars(ftpVars)
	if err != nil {
		utils.Error("Please make sure all required environment variables for FTP are set")
		utils.Fatal("Error missing environment variables: %s", err)
	}
	return &fConfig
}
func initAWSConfig() *AWSConfig {
	//Initialize data configs
	aConfig := AWSConfig{}
	aConfig.endpoint = os.Getenv("AWS_S3_ENDPOINT")
	aConfig.accessKey = os.Getenv("AWS_ACCESS_KEY")
	aConfig.secretKey = os.Getenv("AWS_SECRET_KEY")
	aConfig.bucket = os.Getenv("AWS_S3_BUCKET_NAME")
	aConfig.region = os.Getenv("AWS_REGION")
	disableSsl, err := strconv.ParseBool(os.Getenv("AWS_DISABLE_SSL"))
	if err != nil {
		utils.Fatal("Unable to parse AWS_DISABLE_SSL env var: %s", err)
	}
	aConfig.disableSsl = disableSsl
	aConfig.forcePathStyle = true
	err = utils.CheckEnvVars(awsVars)
	if err != nil {
		utils.Error("Please make sure all required environment variables for AWS S3 are set")
		utils.Fatal("Error checking environment variables: %s", err)
	}
	return &aConfig
}

func initBackupConfig(cmd *cobra.Command) *BackupConfig {
	utils.SetEnv("STORAGE_PATH", backupDestination)
	utils.GetEnv(cmd, "cron-expression", "BACKUP_CRON_EXPRESSION")
	utils.GetEnv(cmd, "period", "BACKUP_CRON_EXPRESSION")
	utils.GetEnv(cmd, "path", "REMOTE_PATH")
	//Get flag value and set env
	remotePath := utils.GetEnvVariable("REMOTE_PATH", "SSH_REMOTE_PATH")
	storage = utils.GetEnv(cmd, "storage", "STORAGE")
	_, _ = cmd.Flags().GetString("mode")
	passphrase := os.Getenv("GPG_PASSPHRASE")
	fromFolder := true
	_ = utils.GetEnv(cmd, "path", "AWS_S3_PATH")
	file = utils.GetEnv(cmd, "file", "FILE_NAME")

	cronExpression := os.Getenv("BACKUP_CRON_EXPRESSION")
	backupPrefix := os.Getenv("BACKUP_PREFIX")
	if backupPrefix == "" {
		encryption = true
		backupPrefix = "backup"
	}
	if file != "" {
		fromFolder = false
	}

	if passphrase != "" {
		encryption = true
	}
	//Initialize data configs
	config := BackupConfig{}
	config.storage = storage
	config.prefix = backupPrefix
	config.encryption = encryption
	config.remotePath = remotePath
	config.passphrase = passphrase
	config.file = file
	config.fromFolder = fromFolder
	config.cronExpression = cronExpression
	return &config
}

type RestoreConfig struct {
	s3Path        string
	remotePath    string
	storage       string
	file          string
	bucket        string
	gpqPassphrase string
}

func initRestoreConfig(cmd *cobra.Command) *RestoreConfig {
	utils.SetEnv("STORAGE_PATH", backupDestination)
	utils.GetEnv(cmd, "path", "REMOTE_PATH")

	//Get flag value and set env
	s3Path := utils.GetEnv(cmd, "path", "AWS_S3_PATH")
	remotePath := utils.GetEnvVariable("REMOTE_PATH", "SSH_REMOTE_PATH")
	storage = utils.GetEnv(cmd, "storage", "STORAGE")
	file = utils.GetEnv(cmd, "file", "FILE_NAME")
	_, _ = cmd.Flags().GetString("mode")
	bucket := utils.GetEnvVariable("AWS_S3_BUCKET_NAME", "BUCKET_NAME")
	gpqPassphrase := os.Getenv("GPG_PASSPHRASE")
	//Initialize restore configs
	rConfig := RestoreConfig{}
	rConfig.s3Path = s3Path
	rConfig.remotePath = remotePath
	rConfig.storage = storage
	rConfig.bucket = bucket
	rConfig.file = file
	rConfig.storage = storage
	rConfig.gpqPassphrase = gpqPassphrase
	return &rConfig
}
