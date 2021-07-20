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

func NewTelegrafLoggerConfiguration(useFlags bool) (*TelegrafLoggerConfig, error) {
	logFactory := &TelegrafLoggerConfig{
		LogFile:   "stderr",
		LogLevel:  "warn",
		LogFormat: "logfmt",
		LogMetric: "log",
	}

	if useFlags {
		flagset := flag.NewFlagSet("main", flag.ContinueOnError)

		flagset.StringVar(&logFactory.LogFile, "log", "stderr", "log destination, can be \"stdout\", \"stderr\", or file path")
		flagset.StringVar(&logFactory.LogLevel, "log-level", "error", "log destination, can be \"stderr\" (default), \"stdout\", or file path")
		flagset.StringVar(&logFactory.LogFormat, "log-format", "line", "log format, can be \"logfmt\" (default), \"json\", or \"line\"")
		flagset.StringVar(&logFactory.LogMetric, "log-metric", "log", "log metric name, used for line protocol")

		err := flagset.Parse(os.Args[1:])

		if err != nil {
			return nil, err
		}
	}

	return logFactory, nil
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
		klog = NewLineLogger(config, outputStream)

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
