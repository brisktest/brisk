/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Used to manage the config file",
	Long: `This command is used to manage the config file.

Brisk uses two config files. One for credentials and general app settings 
and the other for specific project settings.

The general config defaults to ~/.config/brisk/config.toml and stores credential information.
brisk.json contains project specific information and is in the root of the project directory.
`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("config called")
	// },
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
