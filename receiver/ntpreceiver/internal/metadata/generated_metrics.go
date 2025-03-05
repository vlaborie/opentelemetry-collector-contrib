// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/filter"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
)

type metricNtpOffset struct {
	data     pmetric.Metric // data buffer for generated metric.
	config   MetricConfig   // metric config provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills ntp.offset metric with initial data.
func (m *metricNtpOffset) init() {
	m.data.SetName("ntp.offset")
	m.data.SetDescription("Time difference between local and NTP server clocks")
	m.data.SetUnit("ns")
	m.data.SetEmptyGauge()
}

func (m *metricNtpOffset) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val int64) {
	if !m.config.Enabled {
		return
	}
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntValue(val)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricNtpOffset) updateCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricNtpOffset) emit(metrics pmetric.MetricSlice) {
	if m.config.Enabled && m.data.Gauge().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricNtpOffset(cfg MetricConfig) metricNtpOffset {
	m := metricNtpOffset{config: cfg}
	if cfg.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

// MetricsBuilder provides an interface for scrapers to report metrics while taking care of all the transformations
// required to produce metric representation defined in metadata and user config.
type MetricsBuilder struct {
	config                         MetricsBuilderConfig // config of the metrics builder.
	startTime                      pcommon.Timestamp    // start time that will be applied to all recorded data points.
	metricsCapacity                int                  // maximum observed number of metrics per resource.
	metricsBuffer                  pmetric.Metrics      // accumulates metrics data before emitting.
	buildInfo                      component.BuildInfo  // contains version information.
	resourceAttributeIncludeFilter map[string]filter.Filter
	resourceAttributeExcludeFilter map[string]filter.Filter
	metricNtpOffset                metricNtpOffset
}

// MetricBuilderOption applies changes to default metrics builder.
type MetricBuilderOption interface {
	apply(*MetricsBuilder)
}

type metricBuilderOptionFunc func(mb *MetricsBuilder)

func (mbof metricBuilderOptionFunc) apply(mb *MetricsBuilder) {
	mbof(mb)
}

// WithStartTime sets startTime on the metrics builder.
func WithStartTime(startTime pcommon.Timestamp) MetricBuilderOption {
	return metricBuilderOptionFunc(func(mb *MetricsBuilder) {
		mb.startTime = startTime
	})
}
func NewMetricsBuilder(mbc MetricsBuilderConfig, settings receiver.Settings, options ...MetricBuilderOption) *MetricsBuilder {
	mb := &MetricsBuilder{
		config:                         mbc,
		startTime:                      pcommon.NewTimestampFromTime(time.Now()),
		metricsBuffer:                  pmetric.NewMetrics(),
		buildInfo:                      settings.BuildInfo,
		metricNtpOffset:                newMetricNtpOffset(mbc.Metrics.NtpOffset),
		resourceAttributeIncludeFilter: make(map[string]filter.Filter),
		resourceAttributeExcludeFilter: make(map[string]filter.Filter),
	}
	if mbc.ResourceAttributes.NtpHost.MetricsInclude != nil {
		mb.resourceAttributeIncludeFilter["ntp.host"] = filter.CreateFilter(mbc.ResourceAttributes.NtpHost.MetricsInclude)
	}
	if mbc.ResourceAttributes.NtpHost.MetricsExclude != nil {
		mb.resourceAttributeExcludeFilter["ntp.host"] = filter.CreateFilter(mbc.ResourceAttributes.NtpHost.MetricsExclude)
	}

	for _, op := range options {
		op.apply(mb)
	}
	return mb
}

// NewResourceBuilder returns a new resource builder that should be used to build a resource associated with for the emitted metrics.
func (mb *MetricsBuilder) NewResourceBuilder() *ResourceBuilder {
	return NewResourceBuilder(mb.config.ResourceAttributes)
}

// updateCapacity updates max length of metrics and resource attributes that will be used for the slice capacity.
func (mb *MetricsBuilder) updateCapacity(rm pmetric.ResourceMetrics) {
	if mb.metricsCapacity < rm.ScopeMetrics().At(0).Metrics().Len() {
		mb.metricsCapacity = rm.ScopeMetrics().At(0).Metrics().Len()
	}
}

// ResourceMetricsOption applies changes to provided resource metrics.
type ResourceMetricsOption interface {
	apply(pmetric.ResourceMetrics)
}

type resourceMetricsOptionFunc func(pmetric.ResourceMetrics)

func (rmof resourceMetricsOptionFunc) apply(rm pmetric.ResourceMetrics) {
	rmof(rm)
}

// WithResource sets the provided resource on the emitted ResourceMetrics.
// It's recommended to use ResourceBuilder to create the resource.
func WithResource(res pcommon.Resource) ResourceMetricsOption {
	return resourceMetricsOptionFunc(func(rm pmetric.ResourceMetrics) {
		res.CopyTo(rm.Resource())
	})
}

// WithStartTimeOverride overrides start time for all the resource metrics data points.
// This option should be only used if different start time has to be set on metrics coming from different resources.
func WithStartTimeOverride(start pcommon.Timestamp) ResourceMetricsOption {
	return resourceMetricsOptionFunc(func(rm pmetric.ResourceMetrics) {
		var dps pmetric.NumberDataPointSlice
		metrics := rm.ScopeMetrics().At(0).Metrics()
		for i := 0; i < metrics.Len(); i++ {
			switch metrics.At(i).Type() {
			case pmetric.MetricTypeGauge:
				dps = metrics.At(i).Gauge().DataPoints()
			case pmetric.MetricTypeSum:
				dps = metrics.At(i).Sum().DataPoints()
			}
			for j := 0; j < dps.Len(); j++ {
				dps.At(j).SetStartTimestamp(start)
			}
		}
	})
}

// EmitForResource saves all the generated metrics under a new resource and updates the internal state to be ready for
// recording another set of data points as part of another resource. This function can be helpful when one scraper
// needs to emit metrics from several resources. Otherwise calling this function is not required,
// just `Emit` function can be called instead.
// Resource attributes should be provided as ResourceMetricsOption arguments.
func (mb *MetricsBuilder) EmitForResource(options ...ResourceMetricsOption) {
	rm := pmetric.NewResourceMetrics()
	ils := rm.ScopeMetrics().AppendEmpty()
	ils.Scope().SetName(ScopeName)
	ils.Scope().SetVersion(mb.buildInfo.Version)
	ils.Metrics().EnsureCapacity(mb.metricsCapacity)
	mb.metricNtpOffset.emit(ils.Metrics())

	for _, op := range options {
		op.apply(rm)
	}
	for attr, filter := range mb.resourceAttributeIncludeFilter {
		if val, ok := rm.Resource().Attributes().Get(attr); ok && !filter.Matches(val.AsString()) {
			return
		}
	}
	for attr, filter := range mb.resourceAttributeExcludeFilter {
		if val, ok := rm.Resource().Attributes().Get(attr); ok && filter.Matches(val.AsString()) {
			return
		}
	}

	if ils.Metrics().Len() > 0 {
		mb.updateCapacity(rm)
		rm.MoveTo(mb.metricsBuffer.ResourceMetrics().AppendEmpty())
	}
}

// Emit returns all the metrics accumulated by the metrics builder and updates the internal state to be ready for
// recording another set of metrics. This function will be responsible for applying all the transformations required to
// produce metric representation defined in metadata and user config, e.g. delta or cumulative.
func (mb *MetricsBuilder) Emit(options ...ResourceMetricsOption) pmetric.Metrics {
	mb.EmitForResource(options...)
	metrics := mb.metricsBuffer
	mb.metricsBuffer = pmetric.NewMetrics()
	return metrics
}

// RecordNtpOffsetDataPoint adds a data point to ntp.offset metric.
func (mb *MetricsBuilder) RecordNtpOffsetDataPoint(ts pcommon.Timestamp, val int64) {
	mb.metricNtpOffset.recordDataPoint(mb.startTime, ts, val)
}

// Reset resets metrics builder to its initial state. It should be used when external metrics source is restarted,
// and metrics builder should update its startTime and reset it's internal state accordingly.
func (mb *MetricsBuilder) Reset(options ...MetricBuilderOption) {
	mb.startTime = pcommon.NewTimestampFromTime(time.Now())
	for _, op := range options {
		op.apply(mb)
	}
}
