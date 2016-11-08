package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	opentracing "github.com/opentracing/opentracing-go"
)

type Fryer struct {
	tracer   opentracing.Tracer
	lock     SmartLock
	duration time.Duration
}

func newFryer(tracerGen TracerGenerator, duration time.Duration) *Fryer {
	return &Fryer{
		tracer:   tracerGen("donut-fryer"),
		duration: duration,
	}
}

func (f *Fryer) FryDonut(ctx context.Context) {
	var parentSpanContext opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentSpanContext = parent.Context()
	}
	span := f.tracer.StartSpan("fry_donut", opentracing.ChildOf(parentSpanContext))
	defer span.Finish()
	f.lock.Lock(span)
	defer f.lock.Unlock()
	span.LogEvent(fmt.Sprint("starting to fry: ", span.BaggageItem(donutTypeKey)))
	SleepGaussian(f.duration)
}
