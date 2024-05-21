/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"brisk-supervisor/brisk-cli/cli_utils"
	"brisk-supervisor/brisk-cli/projects"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

// jestCmd represents the jest command
var jestCmd = &cobra.Command{
	Use:   "jest",
	Short: "Create a new Jest project and inits a config file in the current directory",
	Long: `This command creates a new Jest project and inits a config file in the current directory.

It creates a default brisk.json file which you can edit with specific project settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		ctx, span := otel.Tracer(name).Start(ctx, "Jest")
		defer span.End()
		ctx, err := cli_utils.CliAddAuthToCtx(ctx)
		if err != nil {
			return err
		}
		ctx, err = cli_utils.CliAddTraceKeyToCtx(ctx, "")
		if err != nil {
			return err
		}

		if projects.CheckForBriskJsonFile(ctx, viper.GetString("PROJECT_CONFIG_FILE")) {
			return fmt.Errorf("project already initialized - %v already exists in the current directory", viper.GetString("PROJECT_CONFIG_FILE"))

		}
		fmt.Println("Creating a new Jest project...")

		err = projects.CreateJestProject(ctx)

		if err != nil {
			return err
		}
		fmt.Printf("New Jest project created successfully. See %v for more information. \n", viper.GetString("PROJECT_CONFIG_FILE"))
		return nil
	},
}

func init() {
	initCmd.AddCommand(jestCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// jestCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// jestCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
