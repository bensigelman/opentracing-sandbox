package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	fryDuration = time.Millisecond * 300
	mixDuration = time.Millisecond * 80
	topDuration = time.Millisecond * 110
)

type DonutService struct {
	tracer    opentracing.Tracer
	mixer     *Mixer
	fryer     *Fryer
	tracerGen TracerGenerator

	toppersLock  sync.Mutex
	toppers      map[string]*Topper
	topperTracer opentracing.Tracer
}

func newDonutService(tracerGen TracerGenerator) *DonutService {
	return &DonutService{
		tracer:       tracerGen("donut-webserver"),
		mixer:        newMixer(tracerGen, mixDuration),
		fryer:        newFryer(tracerGen, fryDuration),
		toppers:      make(map[string]*Topper),
		tracerGen:    tracerGen,
		topperTracer: tracerGen("topper"),
	}
}

func (ds *DonutService) handleRequest(w http.ResponseWriter, r *http.Request) {
	carrier := opentracing.HTTPHeadersCarrier(r.Header)
	clientContext, _ := ds.tracer.Extract(opentracing.HTTPHeaders, carrier)

	type params struct {
		Flavor string `json:"flavor"`
	}

	p := params{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&p)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	flavor := "plain"
	if len(p.Flavor) > 0 {
		flavor = p.Flavor
	}
	ds.makeDonut(clientContext, flavor)

	w.Write([]byte("\n"))
}

func (ds *DonutService) makeDonut(parentSpanContext opentracing.SpanContext, flavor string) {
	donutSpan := ds.tracer.StartSpan("make_donut", ext.RPCServerOption(parentSpanContext))
	donutSpan.SetTag("service", "donut-webserver")
	defer donutSpan.Finish()
	ctx := opentracing.ContextWithSpan(context.Background(), donutSpan)

	ds.mixer.MixBatter(ctx)
	ds.fryer.FryDonut(ctx)

	ds.toppersLock.Lock()
	topper := ds.toppers[flavor]
	if topper == nil {
		topper = newTopper(ds.topperTracer, flavor, topDuration)
		ds.toppers[flavor] = topper
	}
	ds.toppersLock.Unlock()
	topper.SprinkleTopping(ctx)
}
