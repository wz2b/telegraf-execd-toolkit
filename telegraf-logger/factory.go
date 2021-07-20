package telegraf_logger

import (
	"flag"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	stdlog "log"
	"os"
	"strings"
)

type TelegrafLoggerConfig struct {
	LogFile   string // may be file path, "stderr", or "stdout"
	LogLevel  string
	LogFormat string
	LogMetric string
}

func NewTelegrafLoggerConfiguration(flags flag.FlagSet) *TelegrafLoggerConfig {
	logFactory := &TelegrafLoggerConfig{
		LogFile:   "stderr",
		LogLevel:  "warn",
		LogFormat: "logfmt",
		LogMetric: "log",
	}

	flag.StringVar(&logFactory.LogFile, "log", "stderr", "log destination, can be \"stdout\", \"stderr\", or file path")
	flag.StringVar(&logFactory.LogLevel, "log-level", "error", "log destination, can be \"stderr\" (default), \"stdout\", or file path")
	flag.StringVar(&logFactory.LogLevel, "log-format", "line", "log format, can be \"logfmt\" (default), \"json\", or \"line\"")
	flag.StringVar(&logFactory.LogLevel, "log-metric", "log", "log metric name, used for line protocol")

	return logFactory
}

func (config *TelegrafLoggerConfig) Create() kitlog.Logger {
	var outputStream io.Writer

	switch config.LogFile {
	case "stdout":
		outputStream = os.Stdout
	case "stderr", "":
		outputStream = os.Stderr

	default:
		outputStream = &lumberjack.Logger{
			Filename:   config.LogFile,
			MaxSize:    10, // megabytes
			MaxBackups: 10,
			MaxAge:     30,   //days
			Compress:   true, // disabled by default
		}
	}

	var klog kitlog.Logger
	switch strings.ToLower(config.LogFormat) {
	default:
		fallthrough

	case "line":
		klog = newLineLogger(config, outputStream)

	case "logfmt":
		w := kitlog.NewSyncWriter(outputStream)
		klog = kitlog.NewLogfmtLogger(w)
		klog = kitlog.With(klog, "time", kitlog.DefaultTimestampUTC)

	case "json":
		w := kitlog.NewSyncWriter(outputStream)
		klog = kitlog.NewJSONLogger(w)
		klog = kitlog.With(klog, "time", kitlog.DefaultTimestampUTC)
	}


	// Now map the standard go logger to us
	stdlog.SetOutput(kitlog.NewStdlibAdapter(klog))
	stdlog.SetFlags(0)

	// next, apply the level based filter
	klog = level.NewFilter(klog, stringToLevelFilter(config.LogLevel))

	return klog
}

func stringToLevelFilter(levelString string) level.Option {
	switch strings.ToLower(levelString) {

	default:
		fallthrough
	case "warn":
		return level.AllowWarn()
	case "all":
		return level.AllowAll()
	case "debug":
		return level.AllowDebug()
	case "error":
		return level.AllowError()
	case "info":
		return level.AllowInfo()
	case "none":
		return level.AllowNone()
	}
}