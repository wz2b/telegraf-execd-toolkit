package tlogger

import (
	"fmt"
	"io"
	mpool "telegraf-execd-toolkit/pkg/line-metric-encoder"
)

type lineLogger struct {
	config TelegrafLoggerConfig

	writer io.Writer
	pool   *mpool.MetricEncoderPool
}

func newLineLogger(config *TelegrafLoggerConfig, outputStream io.Writer) *lineLogger {
	logger := &lineLogger{
		config: *config,
		writer: outputStream,
		pool:   mpool.NewMetricEncoderPool(),
	}

	return logger
}

func (l *lineLogger) Log(keyvals ...interface{}) error {
	enc := l.pool.Get()
	defer l.pool.Put(enc)

	enc.Begin("log")

	nargs := len(keyvals)
	for i := 0; i < nargs; i += 2 {
		key, keyIsString := keyvals[i].(string)
		if !keyIsString {
			// We should always be handed a string as the key, but if for some oddball reason
			// someone sent us something weird, use Sprint to do our best to convert it to one.
			key = fmt.Sprint(key)
		}
		value := keyvals[i+1]
		enc.AddField(key, value)
	}
	_, err := enc.Write(l.writer)

	return err
}
