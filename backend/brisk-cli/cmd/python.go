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

// pythonCmd represents the python command
var pythonCmd = &cobra.Command{
	Use:   "python",
	Short: "Used to initialize a new Python Brisk project",
	Long:  `Python will create a new Python Brisk project in the current directory. It creates a brisk.json file which you can use to configure your project.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		ctx, span := otel.Tracer(name).Start(ctx, "Python")
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
			fmt.Printf("%s already exists in the current directory. \n", viper.GetString("PROJECT_CONFIG_FILE"))
			return
		}
		fmt.Println("Creating a new Python project...")

		err = projects.CreatePythonProject(ctx)

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("New Python project created successfully. See %v for more information.\n", viper.GetString("PROJECT_CONFIG_FILE"))
	},
}

func init() {
	initCmd.AddCommand(pythonCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pythonCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pythonCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
