package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"
)

type Mixer struct {
	tracer   opentracing.Tracer
	lock     SmartLock
	duration time.Duration
}

func newMixer(duration time.Duration) *Mixer {
	return &Mixer{
		tracer: lightstep.NewTracer(lightstep.Options{
			AccessToken: *accessToken,
			Collector: lightstep.Endpoint{
				Host: "collector-grpc.lightstep.com",
				Port: 443,
			},
			UseGRPC: true,
			Tags: opentracing.Tags{
				lightstep.ComponentNameKey: "donut-mixer",
			},
		}),
		duration: duration,
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
	span.LogEvent(fmt.Sprint("starting to mix: ", span.BaggageItem(donutTypeKey)))
	SleepGaussian(m.duration)
}
