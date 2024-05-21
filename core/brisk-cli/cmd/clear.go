/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Used to clear your Brisk workers",
	Long:  `Clear will allow you to clear your Brisk workers. It is the same as pressing 'r' in the Brisk cli but you can do it without a run.`,
	Run: func(cmd *cobra.Command, args []string) {

		err := ClearWorkers(cmd.Context())
		if err != nil {
			fmt.Println("error clearing workers ", err)
			return
		}

	},
}

func init() {
	workersCmd.AddCommand(clearCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clearCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clearCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
