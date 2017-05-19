package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"time"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

const (
	donutOriginKey = "origin"
)

var (
	accessToken        = flag.String("token", "{your_access_token}", "")
	port               = flag.Int("port", 80, "")
	collectorHost      = flag.String("collector_host", "localhost", "")
	collectorPort      = flag.Int("collector_port", 9997, "")
	tracerType         = flag.String("tracer_type", "lightstep", "")
	orderProcesses     = flag.Int("order", 6, "")
	restockerProcesses = flag.Int("restock", 3, "")
	cleanerProcesses   = flag.Int("clean", 1, "")
)

func SleepGaussian(d time.Duration) {
	time.Sleep(d + time.Duration((float64(d)/3)*rand.NormFloat64()))
}

func currentDir() string {
	// return "github.com/bhs/opentracing-sandbox/donutsalon/"
	return "./"
}
func rootHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("").ParseFiles(currentDir() + "single_page.go.html")
	if err != nil {
		panic(err)
	}
	err = t.ExecuteTemplate(w, "single_page.go.html", nil)
	if err != nil {
		panic(err)
	}

}

type TracerGenerator func(component string) opentracing.Tracer

func main() {
	flag.Parse()
	var tracerGen TracerGenerator
	if *tracerType == "lightstep" {
		tracerGen = func(component string) opentracing.Tracer {
			return lightstep.NewTracer(lightstep.Options{
				AccessToken: *accessToken,
				Collector: lightstep.Endpoint{
					Host: *collectorHost,
					Port: *collectorPort,
				},
				UseGRPC: true,
				Tags: opentracing.Tags{
					lightstep.ComponentNameKey: component,
				},
			})
		}
	} else if *tracerType == "zipkin" {
		fmt.Println("BHS Z")
		tracerGen = func(component string) opentracing.Tracer {
			collector, _ := zipkin.NewHTTPCollector(
				fmt.Sprintf("http://donutsalon.com:9411/api/v1/spans"))
			tracer, _ := zipkin.NewTracer(
				zipkin.NewRecorder(collector, false, "127.0.0.1:0", component))
			return tracer
		}
		t := tracerGen("foo")
		sp := t.StartSpan("blah")
		sp.Finish()
	} else {
		panic(*tracerType)
	}
	ds := newDonutService(tracerGen)

	// Make fake queries in the background.
	backgroundProcess(*orderProcesses, ds, runFakeUser)
	backgroundProcess(*restockerProcesses, ds, runFakeRestocker)
	backgroundProcess(*cleanerProcesses, ds, runFakeCleaner)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/make_donut", ds.handleRequest)
	http.HandleFunc("/status", ds.handleState)
	http.HandleFunc("/clean", ds.handleClean)
	http.HandleFunc("/restock", ds.handleRestock)
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(currentDir()+"public/"))))
	fmt.Println("Starting on :", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	fmt.Println("Exiting", err)
}

func backgroundProcess(max int, ds *DonutService, f func(flavor string, ds *DonutService)) {
	for i := 0; i < max; i++ {
		var flavor string
		switch i % 3 {
		case 0:
			flavor = "cinnamon"
		case 1:
			flavor = "old-fashioned"
		case 2:
			flavor = "sprinkles"
		}
		go f(flavor, ds)
	}
}

func runFakeUser(flavor string, ds *DonutService) {
	for {
		SleepGaussian(250 * time.Millisecond)
		span := ds.tracer.StartSpan(fmt.Sprintf("background_order[%s]", flavor))
		ds.makeDonut(span.Context(), flavor)
		span.Finish()
	}
}

func runFakeRestocker(flavor string, ds *DonutService) {
	for {
		SleepGaussian(500 * time.Millisecond)
		span := ds.tracer.StartSpan(fmt.Sprintf("background_restocker[%s]", flavor))
		ds.restock(span.Context(), flavor)
		span.Finish()
	}
}

func runFakeCleaner(flavor string, ds *DonutService) {
	for {
		SleepGaussian(time.Second)
		span := ds.tracer.StartSpan("background_cleaner")
		ds.cleanFryer(span.Context())
		span.Finish()
	}
}

func startSpanFronContext(name string, tracer opentracing.Tracer, ctx context.Context) opentracing.Span {
	var parentSpanContext opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentSpanContext = parent.Context()
	}
	return tracer.StartSpan(name, opentracing.ChildOf(parentSpanContext))
}
