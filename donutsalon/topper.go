package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	opentracing "github.com/opentracing/opentracing-go"
)

type Topper struct {
	tracer    opentracing.Tracer
	lock      *SmartLock
	donutType string
	duration  time.Duration
}

func newTopper(tracerGen TracerGenerator, donutType string, duration time.Duration) *Topper {
	return &Topper{
		tracer:    tracerGen("donut-topper"),
		donutType: donutType,
		duration:  duration,
		lock:      NewSmartLock(false),
	}
}

func (t *Topper) SprinkleTopping(ctx context.Context) {
	var parentSpanContext opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentSpanContext = parent.Context()
	}
	span := t.tracer.StartSpan(fmt.Sprint("sprinkle_topping: ", t.donutType), opentracing.ChildOf(parentSpanContext))
	defer span.Finish()
	t.lock.Lock(span)
	defer t.lock.Unlock()
	span.LogEvent(fmt.Sprint("starting donut topping: ", span.BaggageItem(donutOriginKey)))
	SleepGaussian(t.duration)
}
