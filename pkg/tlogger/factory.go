package tlogger

import (
	"flag"
	kitlog "github.com/go-kit/kit/log"
	"strings"
)

type TelegrafLoggerConfig struct {
	LogFile   string // may be file path, "stderr", or "stdout"
	LogLevel  string
	LogFormat string
	LogMetric string
}

func NewTelegrafLoggerFactory(flags flag.FlagSet) *TelegrafLoggerConfig {
	logFactory := &TelegrafLoggerConfig{
		LogFile:   "stderr",
		LogLevel:  "warn",
		LogFormat: "logfmt",
		LogMetric: "log",
	}

	flag.StringVar(&logFactory.LogFile, "log", "stderr", "log destination, can be \"stdout\", \"stderr\", or file path")
	flag.StringVar(&logFactory.LogLevel, "log-level", "error", "log destination, can be \"stderr\" (default), \"stdout\", or file path")
	flag.StringVar(&logFactory.LogLevel, "log-format", "line", "log format, can be \"logfmt\" (default) or \"line\"")
	flag.StringVar(&logFactory.LogLevel, "log-metric", "log", "log metric name, used for line protocol")

	return logFactory
}

func (config *TelegrafLoggerConfig) Create() kitlog.Logger {

	switch strings.ToLower(config.LogFormat) {
	case "line":
		return newLineLogger(config)

	}

}
