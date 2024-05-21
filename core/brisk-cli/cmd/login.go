/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"brisk-supervisor/brisk-cli/dotfiles"
	. "brisk-supervisor/shared"
	brisk_user "brisk-supervisor/shared/brisk_user"
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Brisk via web browser",
	Long:  `Login will open a browser and allow you to log in to your Brisk account on brisktest.com.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return login(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func login(ctx context.Context) error {
	// we don't have a user here so there is no auth added to the context

	ctx, span := otel.Tracer(name).Start(ctx, "Login")
	defer span.End()

	nonce, randomErr := RandomHex(32)

	if randomErr != nil {
		return randomErr
	}

	creds, err := brisk_user.Login(ctx, nonce)

	if err != nil {
		fmt.Println(err)
		return err
	}

	dotfiles.AddToConfig(ctx, "apiKey", creds.ApiKey)
	dotfiles.AddToConfig(ctx, "apiToken", creds.ApiToken)
	err = dotfiles.WriteConfig(ctx)
	if err != nil {
		fmt.Println(err)

	}

	return err
}
