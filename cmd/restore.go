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

var RestoreCmd = &cobra.Command{
	Use:     "restore",
	Short:   "Restore backup operation",
	Example: utils.RestoreExample,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.StartRestore(cmd)
	},
}

func init() {
	//Restore
	RestoreCmd.PersistentFlags().StringP("storage", "s", "local", "Storage. local or s3")
	RestoreCmd.PersistentFlags().StringP("path", "P", "", "AWS S3 path without file name. eg: /custom_path or ssh remote path `/home/foo/data`")
	RestoreCmd.PersistentFlags().StringP("file", "f", "", "File name of database")

}
