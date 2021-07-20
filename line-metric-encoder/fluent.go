package line_metric_encoder

type FieldBuilder struct {
	wrappedMetric *WrappedMetric
	key           string
}

type TagBuilder struct {
	wrappedMetric *WrappedMetric
	key           string
}

func (wrapped *WrappedMetric) BuildField(key string) *FieldBuilder {
	return &FieldBuilder{
		wrappedMetric: wrapped,
		key:           key,
	}
}

func (m *FieldBuilder) ValueIfNoErr(value interface{}, err error) *WrappedMetric {
	if err != nil {
		return m.wrappedMetric.WithField(m.key, value)
	}
	return m.wrappedMetric
}

func (m *FieldBuilder) Value(value interface{}) *WrappedMetric {
	return m.wrappedMetric.WithField(m.key, value)
}

func (w *WrappedMetric) BuildTag(key string) *FieldBuilder {
	return &FieldBuilder{
		wrappedMetric: w,
		key:           key,
	}
}
func (m *TagBuilder) Value(value string) *WrappedMetric {
	return m.wrappedMetric.WithField(m.key, value)
}