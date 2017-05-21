package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	opentracing "github.com/opentracing/opentracing-go"
)

type Fryer struct {
	tracer   opentracing.Tracer
	lock     *SmartLock
	duration time.Duration
	oilLevel int
}

func newFryer(tracerGen TracerGenerator, duration time.Duration) *Fryer {
	return &Fryer{
		tracer:   tracerGen("donut-fryer"),
		duration: duration,
		lock:     NewSmartLock(true),
	}
}

func (f *Fryer) FryDonut(ctx context.Context) {
	span := startSpanFronContext("fry_donut", f.tracer, ctx)
	defer span.Finish()

	f.lock.Lock(span)
	defer f.lock.Unlock()

	span.LogEvent(fmt.Sprint("starting to fry: ", span.BaggageItem(donutOriginKey)))
	SleepGaussian(f.duration+time.Duration(f.oilLevel)*time.Millisecond, f.lock.QueueLength())
	f.oilLevel++
}

func (f *Fryer) ChangeOil(ctx context.Context) {
	span := startSpanFronContext("change_oil", f.tracer, ctx)
	defer span.Finish()

	f.lock.Lock(span)
	defer f.lock.Unlock()

	SleepGaussian(f.duration*10, f.lock.QueueLength())
	f.oilLevel = f.oilLevel / 2
}

func (f *Fryer) OilLevel() int {
	span := f.tracer.StartSpan("oil_level")
	defer span.Finish()

	return f.oilLevel
}
