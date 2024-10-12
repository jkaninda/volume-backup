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
	"github.com/jkaninda/volume-backup/utils"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
)

func StartRestore(cmd *cobra.Command) {
	intro()
	restoreConf := initRestoreConfig(cmd)

	switch restoreConf.storage {
	case "s3":
		restoreFromS3(restoreConf.file, restoreConf.bucket, restoreConf.s3Path)
	case "local":
		utils.Info("Restore data from local")
		copyToTmp(backupDestination, restoreConf.file)
		RestoreData(restoreConf.file)
	case "ssh":
		restoreFromRemote(restoreConf.file, restoreConf.remotePath)
	case "ftp":
		restoreFromFTP(restoreConf.file, restoreConf.remotePath)
	default:
		utils.Info("Restore data from local")
		copyToTmp(backupDestination, restoreConf.file)
		RestoreData(restoreConf.file)
	}
}

func restoreFromS3(file, bucket, s3Path string) {
	utils.Info("Restore data from s3")
	err := DownloadFile(tmpPath, file, bucket, s3Path)
	if err != nil {
		utils.Fatal("Error download file from s3 %s %v ", file, err)
	}
	RestoreData(file)
}
func restoreFromRemote(file, remotePath string) {
	utils.Info("Restore data from remote server")
	err := CopyFromRemote(file, remotePath)
	if err != nil {
		utils.Fatal("Error download file from remote server: %s %v", filepath.Join(remotePath, file), err)
	}
	RestoreData(file)
}
func restoreFromFTP(file, remotePath string) {
	utils.Info("Restore data from FTP server")
	err := CopyFromFTP(file, remotePath)
	if err != nil {
		utils.Fatal("Error download file from FTP server: %s %v", filepath.Join(remotePath, file), err)
	}
	RestoreData(file)
}

// RestoreData restore database
func RestoreData(file string) {
	gpgPassphrase := os.Getenv("GPG_PASSPHRASE")
	if file == "" {
		utils.Fatal("Error, file required")
	}
	extension := filepath.Ext(filepath.Join(tmpPath, file))
	if extension == ".gpg" {
		if gpgPassphrase == "" {
			utils.Fatal("Error: GPG passphrase is required, your file seems to be a GPG file.\nYou need to provide GPG keys. GPG_PASSPHRASE environment variable is required.")

		} else {
			//Decrypt file
			err := encryptor.Decrypt(filepath.Join(tmpPath, file), RemoveLastExtension(filepath.Join(tmpPath, file)), gpgPassphrase)
			if err != nil {
				utils.Fatal("Error decrypting file %s %v", file, err)
			}
			//Update file name
			file = RemoveLastExtension(file)
		}
	}

	if utils.FileExists(filepath.Join(tmpPath, file)) {
		utils.Info("Restoring backup...")
		if filepath.Ext(filepath.Join(tmpPath, file)) == ".tar" {
			err := extractTar(filepath.Join(tmpPath, file))
			if err != nil {
				utils.Fatal("Error extracting file %s %v", file, err)
			}
		} else {
			err := extractTarGz(filepath.Join(tmpPath, file))
			if err != nil {
				utils.Fatal("Error extracting file %s %v", file, err)
			}
		}
		utils.Info("Backup has been restored.")

	} else {
		utils.Fatal("File not found in %s", fmt.Sprintf("%s/%s", tmpPath, file))
	}

}

// Extracts a .tar archive to the specified output directory
func extractTar(archivePath string) error {
	utils.Info("Extracting backup...")
	// Open the .tar archive for reading
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a tar reader
	tarReader := tar.NewReader(file)

	// Iterate through the files in the tar archive
	for {
		// Read the next header from the tar file
		header, err := tarReader.Next()
		if err == io.EOF {
			// End of tar archive
			break
		}
		if err != nil {
			return err
		}

		// Determine the output path for the current file or directory
		outputPath := filepath.Join(dataPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create the directory if it doesn't exist
			if err := os.MkdirAll(outputPath, os.FileMode(header.Mode)); err != nil {
				return err
			}

		case tar.TypeReg:
			// Create all parent directories if necessary
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return err
			}

			// Create the file and set the appropriate permissions
			outFile, err := os.Create(outputPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			// Copy the file contents from the tar archive to the output file
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}

			// Set the file permissions
			if err := os.Chmod(outputPath, os.FileMode(header.Mode)); err != nil {
				return err
			}

		default:
			// Handle other file types (e.g., symlinks, etc.) if needed
			fmt.Printf("Skipping file type: %c in file %s\n", header.Typeflag, header.Name)
		}
	}
	utils.Info("Extracting backup...done")

	return nil
}

// Extracts a .tar.gz archive to the specified output directory
func extractTarGz(archivePath string) error {
	utils.Info("Extracting backup...")

	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		outputPath := filepath.Join(dataPath, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directories
			if err := os.MkdirAll(outputPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create files and write data
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return err
			}

			outFile, err := os.Create(outputPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}

			if err := os.Chmod(outputPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		default:
			fmt.Printf("Skipping unsupported file type: %c in file %s\n", header.Typeflag, header.Name)
		}
	}
	utils.Info("Extracting backup...done")

	return nil
}
