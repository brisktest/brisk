/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	constants "brisk-supervisor/shared/constants"
	"fmt"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Outputs the current version.",
	Long:  `Outputs the current version.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		_, span := otel.Tracer(name).Start(ctx, "Version")
		defer span.End()

		fmt.Printf("brisk version %v", constants.VERSION)
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
