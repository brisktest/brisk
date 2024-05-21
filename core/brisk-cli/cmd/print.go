/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"brisk-supervisor/brisk-cli/dotfiles"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

// printCmd represents the print command
var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Prints the current config",
	Long: `Prints out the current config.

The general config defaults to ~/.config/brisk/config.toml and stores credential information and other app wide information.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		ctx, span := otel.Tracer(name).Start(ctx, "Print Config")
		defer span.End()

		dotfiles.PrintConfig(ctx)
	},
}

func init() {
	configCmd.AddCommand(printCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// printCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// printCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
