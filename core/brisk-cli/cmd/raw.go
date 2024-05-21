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

// rawCmd represents the raw command
var rawCmd = &cobra.Command{
	Use:   "raw",
	Short: "Used to initialize a new Brisk project with no Framework",
	Long:  `Raw will create a new Brisk project in the current directory. It creates a brisk.json file which you can use to configure your project. It makes no assumption about the framework.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		ctx, span := otel.Tracer(name).Start(ctx, "Raw")
		defer span.End()

		ctx, err := cli_utils.CliAddAuthToCtx(ctx)
		if err != nil {
			fmt.Println(err)
			return
		}
		ctx, err = cli_utils.CliAddTraceKeyToCtx(ctx, "")
		if err != nil {
			fmt.Println(err)
			return
		}

		if projects.CheckForBriskJsonFile(ctx, viper.GetString("PROJECT_CONFIG_FILE")) {
			fmt.Println("Project already exists.")

			fmt.Printf("%s already exists in the current directory.\n", viper.GetString("PROJECT_CONFIG_FILE"))
			return
		}
		fmt.Println("Creating a new Raw project...")

		err = projects.CreateRawProject(ctx)

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("New Raw project created successfully. See %v for more information.\n", viper.GetString("PROJECT_CONFIG_FILE"))
	},
}

func init() {
	initCmd.AddCommand(rawCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rawCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rawCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
