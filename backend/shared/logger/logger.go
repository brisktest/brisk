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

package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"runtime/debug"
	"strings"

	"github.com/spf13/viper"
	trace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type correlationIdType int
type traceIdType string

const (
	traceIdKey   traceIdType       = "trace-key"
	requestIdKey correlationIdType = iota
	sessionIdKey
	nomadAllocKey
)

var internalLogger *zap.SugaredLogger
var atom zap.AtomicLevel

func initLogger() {
	// a fallback/root logger for events without context

	var logTmp *zap.Logger

	defer func() {
		if os.Getenv("RELEASE_STAGE") != "" {
			internalLogger = internalLogger.With(zap.String("release-stage", os.Getenv("RELEASE_STAGE")))
		}
	}()

	if len(os.Getenv("BRISK_LOG_TO_CONSOLE")) > 0 {

		logTmp, atom = createConsoleCILogger()

		logLevel := viper.GetString("LOG_LEVEL")

		atom.SetLevel(getLogLevel(logLevel))
		internalLogger = logTmp.Sugar()
		return
	} else if len(os.Getenv("NO_LOCAL_LOG_FILE")) == 0 {
		logTmp, atom = createConsoleLogger()

		logLevel := viper.GetString("LOG_LEVEL")
		// fmt.Println("createConsoleLogger: log level is ", logLevel)
		atom.SetLevel(getLogLevel(logLevel))
		internalLogger = logTmp.Sugar()

	} else {
		logTmp, atom = createDockerLogger()
		logLevel := viper.GetString("LOG_LEVEL")
		// fmt.Println("createDockerLogger: log level is ", logLevel)
		atom.SetLevel(getLogLevel(logLevel))

		internalLogger = logTmp.Sugar()
		internalLogger = internalLogger.With(zap.Int("pid", os.Getpid()))
		internalLogger = internalLogger.With(zap.String("exe", path.Base(os.Args[0])))

		if len(os.Getenv("NOMAD_SHORT_ALLOC_ID")) > 0 {
			internalLogger = internalLogger.With(zap.String("alloc-id", os.Getenv("NOMAD_SHORT_ALLOC_ID")))
		}
		if len(os.Getenv("HOST_IP")) > 0 {
			internalLogger = internalLogger.With(zap.String("hostname", os.Getenv("HOST_IP")))
		}
		if len(os.Getenv("NOMAD_TASK_NAME")) > 0 {
			internalLogger = internalLogger.With(zap.String("appname", os.Getenv("NOMAD_TASK_NAME")))

		}

	}
}

func getLogLevel(logLevel string) zapcore.Level {
	switch strings.ToLower(logLevel) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	case "panic":
		return zapcore.PanicLevel
	default:
		return zapcore.ErrorLevel
	}
}

func createConsoleCILogger() (*zap.Logger, zap.AtomicLevel) {
	fmt.Println("creating console CI logger")
	atom := zap.NewAtomicLevel()

	// To keep the example deterministic, disable timestamps in the output.
	consoleCfg := zap.NewDevelopmentConfig()
	consoleCfg.Sampling = nil
	consoleCfg.OutputPaths = []string{"stdout"}
	logLevel := zapcore.Level(getLogLevel(viper.GetString("LOG_LEVEL")))
	atom.SetLevel(logLevel)
	consoleCfg.Level = atom

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleCfg.EncoderConfig),
		zapcore.Lock(os.Stdout),
		atom,
	))
	defer logger.Sync()

	return logger, atom
}

func createConsoleLogger() (*zap.Logger, zap.AtomicLevel) {
	atom := zap.NewAtomicLevel()
	logLevel := zapcore.Level(getLogLevel(viper.GetString("LOG_LEVEL")))
	atom.SetLevel(logLevel)
	// To keep the example deterministic, disable timestamps in the output.
	// consoleCfg := zap.NewProductionConfig()
	consoleCfg := zap.NewDevelopmentConfig()
	consoleCfg.Sampling = nil
	consoleCfg.Level = atom
	if len(viper.GetString("LOG_FILE")) > 0 {
		consoleCfg.OutputPaths = []string{viper.GetString("LOG_FILE")}
	}
	// outPutFile, err := os.OpenFile(consoleCfg.OutputPaths[0], os.O_RDWR|os.O_CREATE|os.O_APPEND, 060 	0)
	logger, err := consoleCfg.Build()

	if err != nil {
		log.Fatal(err)
	}

	// logger := zap.New(zapcore.NewCore(
	// 	zapcore.NewConsoleEncoder(consoleCfg.EncoderConfig),
	// 	zapcore.Lock(outPutFile),
	// 	atom,
	// ))
	defer logger.Sync()

	return logger, atom
}

func createDockerLogger() (*zap.Logger, zap.AtomicLevel) {
	fmt.Println("creating docker logger")
	atom := zap.NewAtomicLevel()

	logLevel := zapcore.Level(getLogLevel(viper.GetString("LOG_LEVEL")))
	atom.SetLevel(logLevel)
	cfg := zap.NewDevelopmentConfig()
	// cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	cfg.Sampling = nil
	cfg.Level = atom
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	return logger, atom
}

func WithTraceId(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, traceIdKey, traceId)
}

func WithNomadAllocId(ctx context.Context, allocId string) context.Context {
	return context.WithValue(ctx, nomadAllocKey, allocId)
}

func SetLoggerLevel(logLevel string) {

	atom.SetLevel(getLogLevel(logLevel))
}

// Logger returns a zap logger with as much context as possible
func Logger(ctx context.Context) *BriskLogger {

	if internalLogger == nil {

		initLogger()
		internalLogger.Warn("logger is initialized - logger")
		// fmt.Println("Log - fmt")
	}

	newLogger := internalLogger
	if ctx != nil {
		if ctxRqId, ok := ctx.Value(requestIdKey).(string); ok {
			newLogger = newLogger.With(zap.String("rqId", ctxRqId))
		}
		if ctxTceId, ok := ctx.Value(traceIdKey).(string); ok {
			newLogger = newLogger.With(zap.String("trace-key", ctxTceId))
		}
		if ctxSessionId, ok := ctx.Value(sessionIdKey).(string); ok {
			newLogger = newLogger.With(zap.String("sessionId", ctxSessionId))
		}

		if ctxNomadAlloc, ok := ctx.Value(nomadAllocKey).(string); ok {
			newLogger = newLogger.With(zap.String("alloc-id", ctxNomadAlloc))
		}
		span := trace.SpanFromContext(ctx)
		if span != nil {
			newLogger = newLogger.With(zap.String("trace-id", span.SpanContext().TraceID().String()))
			newLogger = newLogger.With(zap.String("span-id", span.SpanContext().SpanID().String()))
		}

	}
	return &BriskLogger{*newLogger}
}

func (z *BriskLogger) PrintStackTrace() {
	z.Error(string(debug.Stack()))
}

type BriskLogger struct {
	zap.SugaredLogger
}
