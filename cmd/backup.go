// Package cmd /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package cmd

import (
	"github.com/jkaninda/volume-backup/pkg"
	"github.com/jkaninda/volume-backup/utils"
	"github.com/spf13/cobra"
)

var BackupCmd = &cobra.Command{
	Use:     "backup ",
	Short:   "Backup data",
	Example: utils.BackupExample,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.StartBackup(cmd)
	},
}

func init() {
	//Backup
	BackupCmd.PersistentFlags().StringP("storage", "s", "local", "Storage. local or s3")
	BackupCmd.PersistentFlags().StringP("path", "P", "", "AWS S3 path without file name. eg: /custom_path or ssh remote path `/home/foo/data`")
	BackupCmd.PersistentFlags().StringP("file", "f", "", "Backup a single file. eg: config.json")
	BackupCmd.PersistentFlags().StringP("cron-expression", "", "", "Backup cron expression")
	BackupCmd.PersistentFlags().BoolP("prune", "", false, "Delete old data, default disabled")

}
