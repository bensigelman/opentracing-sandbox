package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"math"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

const (
	donutOriginKey   = "origin"
	maxQueueDuration = float64(8 * time.Second)
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

func SleepGaussian(d time.Duration, queueLength float64) {
	cappedDuration := float64(d)
	if queueLength > 5 {
		cappedDuration = math.Min(cappedDuration, maxQueueDuration/(queueLength-5))
	}
	//	noise := (float64(cappedDuration) / 3) * rand.NormFloat64()
	time.Sleep(time.Duration(cappedDuration))
}

func currentDir() string {
	// return "github.com/bhs/opentracing-sandbox/donutsalon/"
	return "./"
}

type TracerGenerator func(component string) opentracing.Tracer

func main() {
	flag.Parse()
	var tracerGen TracerGenerator
	switch *tracerType {
	case "lightstep":
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
	case "zipkin":
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
	default:
		panic(*tracerType)
	}
	ds := newDonutService(tracerGen)

	// Make fake queries in the background.
	backgroundProcess(*orderProcesses, ds, runFakeUser)
	// backgroundProcess(*restockerProcesses, ds, runFakeRestocker)
	// backgroundProcess(*cleanerProcesses, ds, runFakeCleaner)

	http.HandleFunc("/", ds.handleRoot)
	http.HandleFunc("/make_donut", ds.handleOrder)
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
			flavor = "chocolate"
		case 2:
			flavor = "sprinkles"
		}
		go f(flavor, ds)
	}
}

func runFakeUser(flavor string, ds *DonutService) {
	for {
		SleepGaussian(1500*time.Millisecond, 1)
		span := ds.tracer.StartSpan(fmt.Sprintf("background_order[%s]", flavor))
		ds.makeDonut(span.Context(), flavor)
		span.Finish()
	}
}

func runFakeRestocker(flavor string, ds *DonutService) {
	for {
		SleepGaussian(1500*time.Millisecond, 1)
		span := ds.tracer.StartSpan(fmt.Sprintf("background_restocker[%s]", flavor))
		ds.restock(span.Context(), flavor)
		span.Finish()
	}
}

func runFakeCleaner(flavor string, ds *DonutService) {
	for {
		SleepGaussian(time.Second, 1)
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
