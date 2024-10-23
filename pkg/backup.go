// Package pkg /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package pkg

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/jkaninda/encryptor"
	"github.com/jkaninda/go-storage/pkg/ftp"
	"github.com/jkaninda/go-storage/pkg/local"
	"github.com/jkaninda/go-storage/pkg/s3"
	"github.com/jkaninda/go-storage/pkg/ssh"
	"github.com/jkaninda/volume-backup/utils"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"time"
)

func StartBackup(cmd *cobra.Command) {
	intro()
	//Initialize data configs
	config := initBackupConfig(cmd)

	if config.cronExpression == "" {
		BackupTask(config)
	} else {
		if utils.IsValidCronExpression(config.cronExpression) {
			scheduledMode(config)
		} else {
			utils.Fatal("Cron expression is not valid: %s", config.cronExpression)
		}
	}

}

// Run in scheduled mode
func scheduledMode(config *BackupConfig) {
	utils.Info("Running in Scheduled mode")
	utils.Info("Backup cron expression:  %s", config.cronExpression)
	utils.Info("Storage type %s ", config.storage)

	//Test data
	utils.Info("Testing data configurations...")
	BackupTask(config)
	utils.Info("Testing data configurations...done")
	utils.Info("Creating data job...")
	// Create a new cron instance
	c := cron.New()

	_, err := c.AddFunc(config.cronExpression, func() {
		BackupTask(config)
	})
	if err != nil {
		return
	}
	// Start the cron scheduler
	c.Start()
	utils.Info("Creating backup job...done")
	utils.Info("Backup job started")
	defer c.Stop()
	select {}
}
func BackupTask(config *BackupConfig) {
	utils.Info("Starting backup task...")
	//Generate file name
	backupFileName := fmt.Sprintf("%s_%s.tar.gz", config.prefix, time.Now().Format("20060102_150405"))
	if !config.fromFolder {
		backupFileName = fmt.Sprintf("%s_%s.tar.gz", config.file, time.Now().Format("20060102_150405"))
	}
	config.backupFileName = backupFileName
	switch config.storage {
	case "local":
		localBackup(config)
	case "s3":
		s3Backup(config)
	case "ssh", "remote":
		sshBackup(config)
	case "ftp":
		ftpBackup(config)
	default:
		localBackup(config)
	}
}
func intro() {
	utils.Info("Starting Volume Backup...")
	utils.Info("Copyright (c) 2024 Jonas Kaninda ")
}

// BackupData backup data
func BackupData(config *BackupConfig) {
	utils.Info("Starting data backup...")
	if !config.fromFolder {
		err := compressFile(config.file, config.backupFileName)
		if err != nil {
			utils.Fatal("Error compressing file, error %v", err)
		}
	} else {
		err := utils.CopyDir(dataPath, dataTmpPath)
		if err != nil {
			utils.Fatal("Error copying file, error %v", err)
		}
		err = compressFolder(dataTmpPath, config.backupFileName)
		if err != nil {
			utils.Fatal("Error creating file, error %v", err)
		}
	}
	// Backup data
	utils.Info("Backing up data...")
	utils.Info("Data has been backed up")

}
func localBackup(config *BackupConfig) {
	utils.Info("Backup data to local storage")
	startTime = time.Now().Format(utils.TimeFormat())
	BackupData(config)
	finalFileName := config.backupFileName
	if config.encryption && config.passphrase != "" {
		encryptBackup(config.backupFileName, config.passphrase)
		finalFileName = fmt.Sprintf("%s.%s", config.backupFileName, gpgExtension)
	}
	utils.Info("Backup name is %s", finalFileName)
	//Get backup info
	fileInfo, err := os.Stat(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		utils.Error("Error:", err)
	}
	backupSize = fileInfo.Size()

	//moveToBackup(finalFileName, backupDestination)
	localStorage := local.NewStorage(local.Config{
		LocalPath:  tmpPath,
		RemotePath: backupDestination,
	})
	err = localStorage.Copy(finalFileName)
	if err != nil {
		utils.Fatal("Error copying file, error %v", err)
	}
	//Send notification
	utils.NotifySuccess(&utils.NotificationData{
		File:           finalFileName,
		BackupSize:     backupSize,
		Storage:        config.storage,
		BackupLocation: filepath.Join(config.remotePath, finalFileName),
		StartTime:      startTime,
		EndTime:        time.Now().Format(utils.TimeFormat()),
	})
	//Delete old data
	if config.prune {
		err := localStorage.Prune(config.backupRetention)
		if err != nil {
			utils.Error("Error pruning file, error %v", err)
			return
		}
	}
	//Delete temp
	deleteDataTemp()
	deleteTemp()
}

func s3Backup(config *BackupConfig) {
	utils.Info("Backup data to s3 storage")
	startTime = time.Now().Format(utils.TimeFormat())
	awsConfig := initAWSConfig()
	//Backup data
	BackupData(config)
	finalFileName := config.backupFileName
	if config.encryption {
		encryptBackup(config.backupFileName, config.passphrase)
		finalFileName = fmt.Sprintf("%s.%s", config.backupFileName, "gpg")
	}
	utils.Info("Uploading backup archive to remote storage S3 ... ")

	utils.Info("Backup name is %s", finalFileName)
	s3Storage, err := s3.NewStorage(s3.Config{
		Endpoint:       awsConfig.endpoint,
		Bucket:         awsConfig.bucket,
		AccessKey:      awsConfig.accessKey,
		SecretKey:      awsConfig.secretKey,
		Region:         awsConfig.region,
		DisableSsl:     awsConfig.disableSsl,
		ForcePathStyle: awsConfig.forcePathStyle,
		RemotePath:     awsConfig.remotePath,
		LocalPath:      tmpPath,
	})
	if err != nil {
		utils.Fatal("Error creating s3 storage: %s", err)
	}
	err = s3Storage.Copy(finalFileName)
	if err != nil {
		utils.Fatal("Error copying file, error %v", err)
	}
	//Get backup info
	fileInfo, err := os.Stat(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		utils.Error("Error:", err)
	}
	backupSize = fileInfo.Size()

	//Delete data file from tmp folder
	err = utils.DeleteFile(filepath.Join(tmpPath, config.backupFileName))
	if err != nil {
		fmt.Println("Error deleting file: ", err)

	}
	// Delete old data
	if config.prune {
		err := s3Storage.Prune(config.backupRetention)
		if err != nil {
			utils.Fatal("Error deleting old backup from S3: %s ", err)
		}
	}
	utils.Done("Uploading backup archive to remote storage S3 ... done ")
	//Send notification
	utils.NotifySuccess(&utils.NotificationData{
		File:           finalFileName,
		BackupSize:     backupSize,
		Storage:        config.storage,
		BackupLocation: filepath.Join(config.remotePath, finalFileName),
		StartTime:      startTime,
		EndTime:        time.Now().Format(utils.TimeFormat()),
	})
	//Delete temp
	deleteDataTemp()
	deleteTemp()
}
func sshBackup(config *BackupConfig) {
	utils.Info("Backup data to Remote server")
	startTime = time.Now().Format(utils.TimeFormat())

	//Backup data
	BackupData(config)
	finalFileName := config.backupFileName
	if config.encryption {
		encryptBackup(config.backupFileName, config.passphrase)
		finalFileName = fmt.Sprintf("%s.%s", config.backupFileName, "gpg")
	}
	utils.Info("Uploading backup archive to remote storage ... ")
	utils.Info("Backup name is %s", finalFileName)
	sshConfig, err := loadSSHConfig()
	if err != nil {
		utils.Fatal("Error loading ssh config: %s", err)
	}
	sshStorage, err := ssh.NewStorage(ssh.Config{
		Host:         sshConfig.hostName,
		Port:         sshConfig.port,
		User:         sshConfig.user,
		Password:     sshConfig.password,
		IdentifyFile: sshConfig.identifyFile,
		RemotePath:   config.remotePath,
		LocalPath:    tmpPath,
	})
	if err != nil {
		utils.Fatal("Error creating SSH storage: %s", err)
	}
	err = sshStorage.Copy(finalFileName)
	if err != nil {
		utils.Fatal("Error copying file, error %v", err)
	}

	//Get backup info
	fileInfo, err := os.Stat(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		utils.Error("Error:", err)
	}
	backupSize = fileInfo.Size()

	//Delete data file from tmp folder
	err = utils.DeleteFile(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		utils.Error("Error deleting file: %v", err)

	}
	if config.prune {
		err := sshStorage.Prune(config.backupRetention)
		if err != nil {
			utils.Error("Error pruning file, error %v", err)
			return
		}
	}

	utils.Done("Uploading backup archive to remote storage ... done ")
	//Send notification
	utils.NotifySuccess(&utils.NotificationData{
		File:           finalFileName,
		BackupSize:     backupSize,
		Storage:        config.storage,
		BackupLocation: filepath.Join(config.remotePath, finalFileName),
		StartTime:      startTime,
		EndTime:        time.Now().Format(utils.TimeFormat()),
	})
	//Delete temp
	deleteDataTemp()
	deleteTemp()
}
func ftpBackup(config *BackupConfig) {
	utils.Info("Backup data to the remote FTP server")
	startTime = time.Now().Format(utils.TimeFormat())

	//Backup database
	BackupData(config)
	finalFileName := config.backupFileName
	if config.encryption {
		encryptBackup(config.backupFileName, config.passphrase)
		finalFileName = fmt.Sprintf("%s.%s", config.backupFileName, "gpg")
	}
	utils.Info("Uploading backup archive to the remote FTP server ... ")
	utils.Info("Backup name is %s", finalFileName)
	ftpConfig := initFtpConfig()
	ftpStorage, err := ftp.NewStorage(ftp.Config{
		Host:       ftpConfig.host,
		Port:       ftpConfig.port,
		User:       ftpConfig.user,
		Password:   ftpConfig.password,
		RemotePath: config.remotePath,
		LocalPath:  tmpPath,
	})
	if err != nil {
		utils.Fatal("Error creating FTP storage: %s", err)
	}
	err = ftpStorage.Copy(finalFileName)
	if err != nil {
		utils.Fatal("Error uploading file to the remote FTP server: %s ", err)
	}
	//Get backup info
	fileInfo, err := os.Stat(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		utils.Error("Error:", err)
	}
	backupSize = fileInfo.Size()

	//Delete data file from tmp folder
	err = utils.DeleteFile(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		utils.Error("Error deleting file: %v", err)

	}
	if config.prune {
		err := ftpStorage.Prune(config.backupRetention)
		if err != nil {
			utils.Error("Error pruning file, error %v", err)
			return
		}

	}

	utils.Done("Uploading backup archive to the remote FTP server ... done ")
	//Send notification
	utils.NotifySuccess(&utils.NotificationData{
		File:           finalFileName,
		BackupSize:     backupSize,
		Storage:        config.storage,
		BackupLocation: filepath.Join(config.remotePath, finalFileName),
		StartTime:      startTime,
		EndTime:        time.Now().Format(utils.TimeFormat()),
	})
	//Delete temp
	deleteDataTemp()
	deleteTemp()
}

func encryptBackup(backupFileName, gpqPassphrase string) {
	backupFile, err := os.ReadFile(filepath.Join(tmpPath, backupFileName))
	outputFile := fmt.Sprintf("%s.%s", filepath.Join(tmpPath, backupFileName), gpgExtension)
	if err != nil {
		utils.Fatal("Error reading backup file: %s ", err)
	}
	err = encryptor.Encrypt(backupFile, outputFile, gpqPassphrase)
	if err != nil {
		utils.Fatal("Error during encrypting backup %v", err)
	}

}

// Compresses a folder into a .tar file
func compressFolder(sourceFolder, fileName string) error {
	// Create the output tar file
	outFile, err := os.Create(filepath.Join(tmpPath, fileName))
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Create a gzip writer
	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	// Create a tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk through the source folder
	err = filepath.Walk(sourceFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path to maintain the folder structure
		relPath, err := filepath.Rel(sourceFolder, path)
		if err != nil {
			return err
		}

		// Create a header for the tar file
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		// Update the header name to the relative file path
		header.Name = relPath

		// Write the header to the tar file
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If it's a directory, there's no need to write any file content
		if info.IsDir() {
			return nil
		}

		// Open the file to be written to the tar file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Copy the file data to the tar file
		if _, err := io.Copy(tarWriter, file); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// Compresses a file into a .tar file
func compressFile(sourceFile, fileName string) error {
	//Exist file
	if !utils.FileExists(filepath.Join(dataPath, sourceFile)) {
		utils.Error("file %s does not exist  ", filepath.Join(dataPath, sourceFile))
		return fmt.Errorf("file %s does not exist", sourceFile)
	}
	// Create the output file
	outFile, err := os.Create(filepath.Join(tmpPath, fileName))
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Create a gzip writer
	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	// Create a tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Open the source file to be added to the archive
	file, err := os.Open(filepath.Join(dataPath, sourceFile))
	if err != nil {
		return err
	}
	defer file.Close()

	// Get the file's information (name, size, permissions, etc.)
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a tar header from the file info
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// Update the header name to the base name of the file
	header.Name = info.Name()

	// Write the header to the tar archive
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	// Copy the file content into the tar archive
	if _, err := io.Copy(tarWriter, file); err != nil {
		return err
	}

	return nil
}
