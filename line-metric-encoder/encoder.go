package line_metric_encoder

import (
	"bytes"
	"github.com/influxdata/line-protocol"
	"io"
)

// A non-reentrant encoder that can be called multiple times
type MetricEncoder struct {
	buf        *bytes.Buffer
	serializer *protocol.Encoder
}

// Create a metric encoder (called by the when it needs a new instance)
func NewMetricEncoder() *MetricEncoder {
	buf := bytes.Buffer{}
	serializer := protocol.NewEncoder(&buf)
	serializer.SetMaxLineBytes(-1)

	return &MetricEncoder{
		buf:        &buf,
		serializer: serializer,
	}
}

// Encode a metric to bytes
func (enc *MetricEncoder) Encode(m protocol.Metric) []byte {
	enc.buf.Reset()
	enc.serializer.Encode(m)
	return enc.buf.Bytes()
}



// Encode metric and write to an io.Writer
func (enc *MetricEncoder) Write(m protocol.Metric, writer io.Writer) (int, error) {
	enc.buf.Reset()
	enc.serializer.Encode(m)
	return writer.Write(enc.buf.Bytes())
}

// Create a new metric attached to this encoder pool
func (enc *MetricEncoderPool) NewMetric(metricName string) *WrappedMetric {
		return createMetric(metricName, enc)
}