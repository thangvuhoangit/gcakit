// Package log configures a new logger for an application.
package logger

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	logrusadapter "logur.dev/adapter/logrus"
	"logur.dev/logur"
)

// NewLogger creates a new logger.
func NewLogrusLogger(config Config) logur.LoggerFacade {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	if runtime.GOOS == "windows" {
		fmt.Println("Windows OS detected")
		logger.SetOutput(colorable.NewColorableStdout())
	}

	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:             config.NoColor,
		EnvironmentOverrideColors: true,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "[TIME]",
			logrus.FieldKeyLevel: "[LEVEL]",
			logrus.FieldKeyMsg:   "[MESSAGE]",
			logrus.FieldKeyFunc:  "[CALLER]",
			logrus.FieldKeyFile:  "[LINE]",
		},
	})

	switch config.Format {
	case "logfmt":
		// Already the default

	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "[TIME]",
				logrus.FieldKeyLevel: "[LEVEL]",
				logrus.FieldKeyMsg:   "[MESSAGE]",
				logrus.FieldKeyFunc:  "[CALLER]",
				logrus.FieldKeyFile:  "[LINE]",
			},
		})
	}

	if level, err := logrus.ParseLevel(config.Level); err == nil {
		logger.SetLevel(level)
	}

	if config.Mode == "development" {
		logger.SetReportCaller(true)
	}

	return logrusadapter.New(logger)
}
