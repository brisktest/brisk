/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"brisk-supervisor/brisk-cli/cmd"
	"brisk-supervisor/brisk-cli/dotfiles"
	"brisk-supervisor/shared/constants"
	"brisk-supervisor/shared/honeycomb"
	"context"
	"fmt"

	"github.com/bugsnag/bugsnag-go"
	"go.opentelemetry.io/otel"
)

const name = "brisk-cli/main.go"

func main() {
	bugsnag.Configure(bugsnag.Configuration{
		APIKey:  constants.BUGSNAG_API_KEY,
		Logger:  BugsnagLogger{},
		AppType: "cli",
	})
	shutdown := honeycomb.InitTracer()
	defer shutdown()
	ctx := context.Background()
	dotfiles.InitCLIViper(ctx)

	ctx, span := otel.Tracer(name).Start(ctx, "BriskCLI")
	defer span.End()

	cmd.ExecuteContext(ctx)
	fmt.Println()
}

type BugsnagLogger struct {
}

func (b BugsnagLogger) Printf(format string, v ...interface{}) {
	// we already log enough
}
