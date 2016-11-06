package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/context"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	fryDuration = time.Millisecond * 300
	mixDuration = time.Millisecond * 80
	topDuration = time.Millisecond * 110
)

type DonutService struct {
	tracer opentracing.Tracer
	mixer  *Mixer
	fryer  *Fryer

	toppersLock sync.Mutex
	toppers     map[string]*Topper
}

func newDonutService() *DonutService {
	return &DonutService{
		tracer: lightstep.NewTracer(lightstep.Options{
			AccessToken: *accessToken,
			Collector: lightstep.Endpoint{
				Host: "collector-grpc.lightstep.com",
				Port: 443,
			},
			UseGRPC: true,
			Tags: opentracing.Tags{
				lightstep.ComponentNameKey: "donut-webserver",
			},
		}),
		mixer:   newMixer(mixDuration),
		fryer:   newFryer(fryDuration),
		toppers: make(map[string]*Topper),
	}
}

func (ds *DonutService) handleRequest(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Flavor string `json:"flavor"`
	}

	carrier := opentracing.HTTPHeadersCarrier(r.Header)
	clientContext, _ := ds.tracer.Extract(opentracing.HTTPHeaders, carrier)

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
	defer donutSpan.Finish()
	ctx := opentracing.ContextWithSpan(context.Background(), donutSpan)

	ds.mixer.MixBatter(ctx)
	ds.fryer.FryDonut(ctx)

	ds.toppersLock.Lock()
	topper := ds.toppers[flavor]
	if topper == nil {
		topper = newTopper(flavor, topDuration)
		ds.toppers[flavor] = topper
	}
	ds.toppersLock.Unlock()
	topper.SprinkleTopping(ctx)
}
