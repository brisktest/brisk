/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"brisk-supervisor/brisk-cli/cli_utils"
	"brisk-supervisor/brisk-cli/projects"
	. "brisk-supervisor/shared/logger"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  `This command will list all projects for the current logged in user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		ctx, span := otel.Tracer(name).Start(ctx, "List")
		defer span.End()
		ctx, err := cli_utils.CliAddAuthToCtx(ctx)
		if err != nil {
			return err
		}
		ctx, err = cli_utils.CliAddTraceKeyToCtx(ctx, "")
		if err != nil {
			return err
		}

		projects, err := projects.ListAll(ctx)

		if err != nil {
			Logger(ctx).Errorf("error listing projects - %v", err)
			fmt.Printf("Error listing projects :%v \n", err)
			return err
		}

		w := new(tabwriter.Writer)

		w = w.Init(os.Stdout, 0, 8, 1, '\t', 0)

		fmt.Fprintln(w, "Name\tFramework\tProject Token\t Concurrency \t")
		fmt.Fprintln(w, "\t\t\t")
		fmt.Println("")
		for _, project := range projects {
			// Format in tab-separated columns with a tab stop of 8.
			s := fmt.Sprintf("%s\t%s\t%s\t%d\t", project.Name, project.Framework, project.ProjectToken, project.Concurrency)
			fmt.Fprintln(w, s)

		}
		w.Flush()
		fmt.Println("")
		return nil
	},
}

func init() {
	projectCmd.AddCommand(listCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
