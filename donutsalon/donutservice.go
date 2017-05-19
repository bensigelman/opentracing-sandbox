package main

import (
	"encoding/json"
	"fmt"
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

	p := struct {
		Flavor string `json:"flavor"`
	}{}
	unmarshalJSON(r.Body, &p)
	if p.Flavor == "" {
		panic("flavor not set")
	}

	span := ds.tracer.StartSpan(fmt.Sprintf("order_donut[%s]", p.Flavor), ext.RPCServerOption(clientContext))
	defer span.Finish()

	err := ds.makeDonut(span.Context(), p.Flavor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
	span := ds.tracer.StartSpan("cleaner")
	ds.cleanFryer(span.Context())
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (ds *DonutService) handleRestock(w http.ResponseWriter, r *http.Request) {
	flavor := r.FormValue("flavor")
	span := ds.tracer.StartSpan(fmt.Sprintf("restocker[%s]", flavor))
	ds.restock(span.Context(), flavor)
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
