package line_metric_encoder

import (
	"bytes"
	"github.com/influxdata/line-protocol"
	"io"
	"time"
)

type MetricEncoder struct {
	buf        *bytes.Buffer
	serializer *protocol.Encoder
	metric     protocol.MutableMetric
}

func NewMetricEncoder() *MetricEncoder {
	buf := bytes.Buffer{}
	serializer := protocol.NewEncoder(&buf)
	serializer.SetMaxLineBytes(-1)

	return &MetricEncoder{
		buf:        &buf,
		serializer: serializer,
	}
}

func (enc *MetricEncoder) Begin(metricName string) *MetricEncoder{
	now := time.Now()
	var tags map[string]string
	var fields map[string]interface{}

	enc.buf.Reset()
	enc.metric, _ = protocol.New(metricName, tags, fields, now)

	return enc
}

func (enc *MetricEncoder) WithTime(tm time.Time) *MetricEncoder{
	enc.metric.SetTime(tm)
	return enc
}

func (enc *MetricEncoder) AddTag(key string, value string) *MetricEncoder {
	enc.metric.AddTag(key, value)
	return enc
}

func (enc *MetricEncoder) AddField(key string, value interface{}) *MetricEncoder {
	// Check if the value is is an error.  If so, convert it to a string
	valueAsError, valueIsError := value.(error)
	if valueIsError {
		value = valueAsError.Error()
	}

	enc.metric.AddField(key, value)
	return enc
}

func (enc *MetricEncoder) Encode() []byte {
	enc.buf.Reset()
	enc.serializer.Encode(enc.metric)
	return enc.buf.Bytes()
}

func (enc *MetricEncoder) Write(writer io.Writer) (int, error) {
	enc.buf.Reset()
	enc.serializer.Encode(enc.metric)
	return writer.Write(enc.buf.Bytes())
}