// Copyright 2024 Brisk, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "brisk",
	Short: "Brisk: ⚡Lightning Fast Tests",
	Long:  "Brisk: ⚡Lightning Fast Tests.\nLearn more at https://brisktest.com",

	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return RunBrisk(cmd.Context())

	},
}

func ExecuteContext(ctx context.Context) {

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		fmt.Println("exiting")
		os.Exit(1)
	}

}

func initFlags(rootCmd *cobra.Command) {

	rootCmd.PersistentFlags().StringP("config", "c", "brisk.json", "project config file")
	rootCmd.PersistentFlags().StringP("credentials", "a", "$HOME/.config/brisk/config.toml", "brisk credentials file")
	rootCmd.PersistentFlags().BoolP("watch", "w", true, "should brisk watch for local changes")
	// cfgFile := os.Getenv("BRISK_CONFIG")
	err := viper.BindPFlag("PROJECT_CONFIG_FILE", rootCmd.PersistentFlags().Lookup("config"))
	if err != nil {
		fmt.Println("Error binding config flag")
	}

	err = viper.BindPFlag("CREDENTIALS_CONFIG", rootCmd.PersistentFlags().Lookup("credentials"))
	if err != nil {
		fmt.Println("Error binding credentials flag")
	}

	err = viper.BindPFlag("watch", rootCmd.PersistentFlags().Lookup("watch"))
	if err != nil {
		fmt.Println("Error binding watch flag")
	}

}
func init() {
	initFlags(rootCmd)

}
