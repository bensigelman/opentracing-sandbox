package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	opentracing "github.com/opentracing/opentracing-go"
)

type Topper struct {
	tracer    opentracing.Tracer
	lock      SmartLock
	donutType string
	duration  time.Duration
}

func newTopper(tracer opentracing.Tracer, donutType string, duration time.Duration) *Topper {
	return &Topper{
		tracer:    tracer,
		donutType: donutType,
		duration:  duration,
	}
}

func (t *Topper) SprinkleTopping(ctx context.Context) {
	var parentSpanContext opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentSpanContext = parent.Context()
	}
	span := t.tracer.StartSpan("sprinkle_topping", opentracing.ChildOf(parentSpanContext))
	span.SetTag("service", "donut-mixer")
	span.SetTag("flavor", t.donutType)
	defer span.Finish()
	t.lock.Lock(span)
	defer t.lock.Unlock()
	span.LogEvent(fmt.Sprint("starting donut topping: ", span.BaggageItem(donutOriginKey)))
	SleepGaussian(t.duration)
}
