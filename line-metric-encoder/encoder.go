package line_metric_encoder

import (
	"bytes"
	"github.com/influxdata/line-protocol"
	"io"
)

type MetricEncoder struct {
	buf        *bytes.Buffer
	serializer *protocol.Encoder
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

func (enc *MetricEncoder) Encode(m protocol.Metric) []byte {
	enc.buf.Reset()
	enc.serializer.Encode(m)
	return enc.buf.Bytes()
}

func (enc *MetricEncoder) Write(m protocol.Metric, writer io.Writer) (int, error) {
	enc.buf.Reset()
	enc.serializer.Encode(m)
	return writer.Write(enc.buf.Bytes())
}

func (enc *MetricEncoderPool) NewMetric(metricName string) *WrappedMetric {
		return createMetric(metricName, enc)
}