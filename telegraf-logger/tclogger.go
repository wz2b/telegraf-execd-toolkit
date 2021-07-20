package telegraf_logger

import (
	"fmt"
	mpool "github.com/wz2b/telegraf-execd-toolkit/line-metric-encoder"
	"io"
)

type lineLogger struct {
	config TelegrafLoggerConfig

	writer io.Writer
	pool   *mpool.MetricEncoderPool
}

//
// Create a line-protocol logger using its own metric encoder pool
//
func NewLineLogger(config *TelegrafLoggerConfig, outputStream io.Writer) *lineLogger {
	return NewLineLoggerUsingPool(config, outputStream, mpool.NewMetricEncoderPool())
}

//
// Create a line-protocol logger using a provided metric encoder pool
//
func NewLineLoggerUsingPool(config *TelegrafLoggerConfig, outputStream io.Writer, encoderPool *mpool.MetricEncoderPool) *lineLogger {
	logger := &lineLogger{
		config: *config,
		writer: outputStream,
		pool:   encoderPool,
	}

	return logger
}

func (l *lineLogger) Log(keyvals ...interface{}) error {
	m := l.pool.NewMetric(l.config.LogMetric)

	nargs := len(keyvals)
	for i := 0; i < nargs; i += 2 {
		key, keyIsString := keyvals[i].(string)
		if !keyIsString {
			// We should always be handed a string as the key, but if for some oddball reason
			// someone sent us something weird, use Sprint to do our best to convert it to one.
			key = fmt.Sprint(key)
		}
		value := keyvals[i+1]
		m.WithField(key, value)
	}
	_, err := m.Write(l.writer)

	return err
}
