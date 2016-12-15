package main

import (
	"flag"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"time"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	donutOriginKey = "donut_origin"
)

var (
	accessToken = flag.String("token", "{your_access_token}", "")
	port        = flag.Int("port", 80, "")
	tracerType  = flag.String("tracer_type", "lightstep", "")
)

func SleepGaussian(d time.Duration) {
	time.Sleep(d + time.Duration((float64(d)/3)*rand.NormFloat64()))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("").ParseFiles("github.com/bensigelman/opentracing-sandbox/donutsalon/single_page.go.html")
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
					Host: "collector-grpc.lightstep.com",
					Port: 443,
				},
				UseGRPC: true,
				Tags: opentracing.Tags{
					lightstep.ComponentNameKey: component,
				},
			})
		}
	} else if *tracerType == "zipkin" {
		tracerGen = func(component string) opentracing.Tracer {
			collector, _ := zipkin.NewHTTPCollector(
				fmt.Sprintf("http://donutsalon.com:9411/api/v1/spans"))
			tracer, _ := zipkin.NewTracer(
				zipkin.NewRecorder(collector, false, "127.0.0.1:0", component))
			return tracer
		}
	} else {
		panic(*tracerType)
	}
	ds := newDonutService(tracerGen)

	// Make fake queries in the background.
	go func() {
		for _ = range time.Tick(time.Millisecond * 500) {
			var flavor string
			switch rand.Int() % 5 {
			case 0:
				flavor = "cinnamon"
			case 1:
				flavor = "old-fashioned"
			case 2:
				flavor = "chocolate"
			case 3:
				flavor = "glazed"
			case 4:
				flavor = "cruller"
			}
			span := ds.tracer.StartSpan("bulk_donut")
			span.SetTag("service", "background-service")
			span.SetBaggageItem(donutOriginKey, flavor+" (daemon-donuts)")
			ds.makeDonut(span.Context(), flavor)
			span.Finish()
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/make_donut", ds.handleRequest)
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("github.com/bensigelman/opentracing-sandbox/donutsalon/public"))))
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}
