/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// workersCmd represents the workers command
var workersCmd = &cobra.Command{
	Use:   "workers",
	Short: "Used to interact with your Brisk workers",
	Long:  `Workers will allow you to interact with your Brisk workers. You can clear your workers`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	// fmt.Println("workers called")
	// },
}

func init() {
	rootCmd.AddCommand(workersCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// workersCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// workersCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
