package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	opentracing "github.com/opentracing/opentracing-go"
)

type Mixer struct {
	tracer   opentracing.Tracer
	lock     *SmartLock
	duration time.Duration
}

func newMixer(tracerGen TracerGenerator, duration time.Duration) *Mixer {
	return &Mixer{
		tracer:   tracerGen("donut-mixer"),
		duration: duration,
		lock:     NewSmartLock(false),
	}
}

func (m *Mixer) MixBatter(ctx context.Context) {
	var parentSpanContext opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentSpanContext = parent.Context()
	}
	span := m.tracer.StartSpan("mix_batter", opentracing.ChildOf(parentSpanContext))
	defer span.Finish()
	m.lock.Lock(span)
	defer m.lock.Unlock()
	span.LogEvent(fmt.Sprint("starting to mix: ", span.BaggageItem(donutOriginKey)))
	SleepGaussian(m.duration)
}
