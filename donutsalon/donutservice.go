package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/context"

	"io"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	fryDuration = time.Millisecond * 100
	payDuration = time.Millisecond * 200
	topDuration = time.Millisecond * 110
)

type DonutService struct {
	tracer    opentracing.Tracer
	payer     *Payer
	fryer     *Fryer
	tracerGen TracerGenerator

	toppersLock sync.Mutex
	toppers     map[string]*Topper
}

func newDonutService(tracerGen TracerGenerator) *DonutService {
	return &DonutService{
		tracer:    tracerGen("donut-webserver"),
		payer:     NewPayer(tracerGen, payDuration),
		fryer:     newFryer(tracerGen, fryDuration),
		toppers:   make(map[string]*Topper),
		tracerGen: tracerGen,
	}
}

func (ds *DonutService) handleRequest(w http.ResponseWriter, r *http.Request) {
	carrier := opentracing.HTTPHeadersCarrier(r.Header)
	clientContext, _ := ds.tracer.Extract(opentracing.HTTPHeaders, carrier)

	type params struct {
		Flavor string `json:"flavor"`
	}

	p := params{}
	unmarshalJSON(r.Body, &p)
	flavor := "plain"
	if len(p.Flavor) > 0 {
		flavor = p.Flavor
	}
	err := ds.makeDonut(clientContext, flavor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("\n"))
}

func (ds *DonutService) handleState(w http.ResponseWriter, r *http.Request) {
	state := struct {
		OilLevel  int
		Inventory map[string]int
	}{
		OilLevel:  ds.fryer.OilLevel(),
		Inventory: ds.inventory(),
	}
	data, err := json.Marshal(state)
	panicErr(err)

	w.Write(data)
}

func (ds *DonutService) handleClean(w http.ResponseWriter, r *http.Request) {
	carrier := opentracing.HTTPHeadersCarrier(r.Header)
	clientContext, _ := ds.tracer.Extract(opentracing.HTTPHeaders, carrier)
	ds.cleanFryer(clientContext)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (ds *DonutService) handleRestock(w http.ResponseWriter, r *http.Request) {
	carrier := opentracing.HTTPHeadersCarrier(r.Header)
	clientContext, _ := ds.tracer.Extract(opentracing.HTTPHeaders, carrier)
	flavor := r.FormValue("flavor")
	ds.restock(clientContext, flavor)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (ds *DonutService) makeDonut(parentSpanContext opentracing.SpanContext, flavor string) error {
	donutSpan := ds.tracer.StartSpan("make_donut", ext.RPCServerOption(parentSpanContext))
	defer donutSpan.Finish()
	ctx := opentracing.ContextWithSpan(context.Background(), donutSpan)

	ds.payer.BuyDonut(ctx)
	ds.fryer.FryDonut(ctx)

	ds.toppersLock.Lock()
	topper := ds.toppers[flavor]
	if topper == nil {
		topper = newTopper(ds.tracerGen, flavor, topDuration)
		ds.toppers[flavor] = topper
	}
	ds.toppersLock.Unlock()
	return topper.SprinkleTopping(ctx)
}

func (ds *DonutService) cleanFryer(parentSpanContext opentracing.SpanContext) {
	donutSpan := ds.tracer.StartSpan("clean_fryer", ext.RPCServerOption(parentSpanContext))
	defer donutSpan.Finish()
	ctx := opentracing.ContextWithSpan(context.Background(), donutSpan)

	ds.fryer.ChangeOil(ctx)
}

func (ds *DonutService) inventory() map[string]int {
	inventory := make(map[string]int)

	ds.toppersLock.Lock()
	defer ds.toppersLock.Unlock()
	for flavor, topper := range ds.toppers {
		inventory[flavor] = topper.Quantity()
	}

	return inventory
}

func (ds *DonutService) restock(parentSpanContext opentracing.SpanContext, flavor string) {
	donutSpan := ds.tracer.StartSpan("restock_ingredients", ext.RPCServerOption(parentSpanContext))
	defer donutSpan.Finish()
	ctx := opentracing.ContextWithSpan(context.Background(), donutSpan)

	ds.toppersLock.Lock()
	defer ds.toppersLock.Unlock()

	topper := ds.toppers[flavor]
	if topper == nil {
		topper = newTopper(ds.tracerGen, flavor, topDuration)
		ds.toppers[flavor] = topper
	}

	topper.Restock(ctx)
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func unmarshalJSON(body io.ReadCloser, data interface{}) {
	defer body.Close()
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&data)
	panicErr(err)
}
