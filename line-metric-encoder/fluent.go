package line_metric_encoder

// Field Builder
type FieldBuilder struct {
	wrappedMetric *WrappedMetric
	key           string
}

// Tag Builder
type TagBuilder struct {
	wrappedMetric *WrappedMetric
	key           string
}

// Create a field builder
func (wrapped *WrappedMetric) BuildField(key string) *FieldBuilder {
	return &FieldBuilder{
		wrappedMetric: wrapped,
		key:           key,
	}
}

// Emit a field
func (m *FieldBuilder) Value(value interface{}) *WrappedMetric {
	return m.wrappedMetric.WithField(m.key, value)
}

// Emit a field if err != nil
func (m *FieldBuilder) ValueIfNoErr(value interface{}, err error) *WrappedMetric {
	if err != nil {
		return m.Value(value)
	}
	return m.wrappedMetric
}

// Create a tag builder
func (w *WrappedMetric) BuildTag(key string) *TagBuilder {
	return &TagBuilder{
		wrappedMetric: w,
		key:           key,
	}
}

// Emit a tag
func (m *TagBuilder) Value(value string) *WrappedMetric {
	return m.Value(value)
}

// Emit a tag if err != nil
func (m *TagBuilder) ValueIfNoErr(value string, err error) *WrappedMetric {
	if err != nil {
		return m.wrappedMetric.WithTag(m.key, value)
	}
	return m.wrappedMetric
}
