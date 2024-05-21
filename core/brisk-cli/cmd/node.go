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

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Used to initialize a new Node project",
	Long:  `Node will create a new Node project in the current directory. It creates a brisk.json file which you can use to configure your project.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		ctx, span := otel.Tracer(name).Start(ctx, "Node")
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
		fmt.Println("Creating a new Node project...")

		err = projects.CreateJestProject(ctx)

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("New Node project created successfully. See %v for more information. \n", viper.GetString("PROJECT_CONFIG_FILE"))
	},
}

func init() {
	initCmd.AddCommand(nodeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nodeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
