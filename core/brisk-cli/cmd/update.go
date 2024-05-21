/*
Copyright © 2022 Brisk Inc <sean@brisktest.com>
*/
package cmd

import (
	"brisk-supervisor/brisk-cli/utilities"
	"fmt"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates to the latest version",
	Long: `Update brisk cli to the latest verison:

`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		ctx, span := otel.Tracer(name).Start(ctx, "Update")
		defer span.End()

		err := utilities.DoUpdateNow(ctx)
		if err != nil {
			span.RecordError(err)
			fmt.Println(err.Error())
			return
		}

	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
