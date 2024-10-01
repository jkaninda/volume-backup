// Package pkg /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package pkg

import (
	"github.com/jkaninda/volume-backup/utils"
	"os"
	"path/filepath"
	"time"
)

func copyToTmp(sourcePath string, backupFileName string) {
	//Copy data from storage to /tmp
	err := utils.CopyFile(filepath.Join(sourcePath, backupFileName), filepath.Join(tmpPath, backupFileName))
	if err != nil {
		utils.Fatal("Error copying file %s %v", backupFileName, err)

	}
}
func moveToBackup(backupFileName string, destinationPath string) {
	//Copy data from tmp folder to storage destination
	err := utils.CopyFile(filepath.Join(tmpPath, backupFileName), filepath.Join(destinationPath, backupFileName))
	if err != nil {
		utils.Fatal("Error copying file %s %v", backupFileName, err)

	}
	//Delete data file from tmp folder
	err = utils.DeleteFile(filepath.Join(tmpPath, backupFileName))
	if err != nil {
		utils.Error("Error deleting file: %s", err)

	}
	utils.Done("Data has been backed up and copied to  %s", filepath.Join(destinationPath, backupFileName))
}
func deleteOldBackup(retentionDays int) {
	utils.Info("Deleting old backups...")
	// Define the directory path
	backupDir := backupDestination + "/"
	// Get current time
	currentTime := time.Now()
	// Delete file
	deleteFile := func(filePath string) error {
		err := os.Remove(filePath)
		if err != nil {
			utils.Fatal("Error:", err)
		} else {
			utils.Done("File %s deleted successfully", filePath)
		}
		return err
	}

	// Walk through the directory and delete files modified more than specified days ago
	err := filepath.Walk(backupDir, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if it's a regular file and if it was modified more than specified days ago
		if fileInfo.Mode().IsRegular() {
			timeDiff := currentTime.Sub(fileInfo.ModTime())
			if timeDiff.Hours() > 24*float64(retentionDays) {
				err := deleteFile(filePath)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		utils.Fatal("Error:", err)
		return
	}
	utils.Done("Deleting old backups...done")
}
func deleteTemp() {
	utils.Info("Deleting %s ...", tmpPath)
	err := filepath.Walk(tmpPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if the current item is a file
		if !info.IsDir() {
			// Delete the file
			err = os.Remove(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		utils.Error("Error deleting files: %v", err)
	} else {
		utils.Info("Deleting %s ... done", tmpPath)
	}
}
