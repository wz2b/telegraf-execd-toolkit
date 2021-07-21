package line_metric_encoder

import (
	protocol "github.com/influxdata/line-protocol"
	"io"
	"os"
	"time"
)

type WrappedMetric struct {
	protocol.MutableMetric
	encoderPool *MetricEncoderPool
}

func createMetric(name string, encoderPool *MetricEncoderPool) *WrappedMetric {
	var tags map[string]string
	var fields map[string]interface{}
	metric, _ := protocol.New(name, tags, fields, time.Now())

	return &WrappedMetric{
		MutableMetric: metric,
		encoderPool: encoderPool,
	}
}

func (m *WrappedMetric) WithTime(tm time.Time) *WrappedMetric {
	m.SetTime(tm)
	return m
}

func (m *WrappedMetric) WithTag(key string, value string) *WrappedMetric {
	m.AddTag(key, value)
	return m
}

func (m *WrappedMetric) WithField(key string, value interface{}) *WrappedMetric {
	// Check if the value is is an error.  If so, convert it to a string
	valueAsError, valueIsError := value.(error)
	if valueIsError {
		value = valueAsError.Error()
	}

	m.AddField(key, value)
	return m
}

func (m *WrappedMetric) Print() (int, error) {
	return m.Write(os.Stdout)
}

func (m *WrappedMetric) Write(out io.Writer) (int, error) {
	encoder := m.encoderPool.Get()
	defer m.encoderPool.PutBack(encoder)
	return encoder.Write(m, out)
}

func (m *WrappedMetric) Encode() []byte {
	encoder := m.encoderPool.Get()
	defer m.encoderPool.PutBack(encoder)

	bytes := encoder.Encode(m)

	return bytes
}