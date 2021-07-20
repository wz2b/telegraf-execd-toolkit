package line_metric_encoder

import "sync"

type MetricEncoderPool struct {
	pool sync.Pool
}

func (m *MetricEncoderPool) Get() *MetricEncoder {
	return m.pool.Get().(*MetricEncoder)
}

func (m *MetricEncoderPool) PutBack(encoder interface{}){
	m.pool.Put(encoder)
}


func NewMetricEncoderPool() *MetricEncoderPool {

	pool := &MetricEncoderPool{
		pool: sync.Pool{
			New: func() interface{} { return NewMetricEncoder() },
		},
	}

	return pool
}
