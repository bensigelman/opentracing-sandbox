package main

import (
	"fmt"
	"math"
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

	f.lock.Lock(span)

	span.LogEvent(fmt.Sprint("starting to fry: ", span.BaggageItem(donutOriginKey)))
	SleepGaussian(f.duration+time.Duration(math.Min(500, float64(f.oilLevel)))*time.Millisecond, f.lock.QueueLength())
	f.oilLevel++

	f.lock.Unlock()
	span.Finish()
}

func (f *Fryer) ChangeOil(ctx context.Context) {
	span := startSpanFronContext("change_oil", f.tracer, ctx)

	f.lock.Lock(span)
	if f.oilLevel < 10 {
		SleepGaussian(f.duration*5, f.lock.QueueLength())
	} else {
		SleepGaussian(time.Second*10, 0)
	}
	f.oilLevel = f.oilLevel / 2

	f.lock.Unlock()
	span.Finish()
}

func (f *Fryer) OilLevel() int {
	span := f.tracer.StartSpan("oil_level")
	defer span.Finish()

	return f.oilLevel
}
