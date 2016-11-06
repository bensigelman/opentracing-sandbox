package main

import (
	"flag"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"time"
)

const (
	donutTypeKey = "donut_type"
)

var (
	accessToken = flag.String("token", "{your_access_token}", "")
	port        = flag.Int("port", 80, "")
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

func main() {
	flag.Parse()
	ds := newDonutService()

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
			span := ds.tracer.StartSpan("background_donut")
			span.SetBaggageItem(donutTypeKey, flavor)
			ds.makeDonut(span.Context(), flavor)
			span.Finish()
		}
	}()

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/make_donut", ds.handleRequest)
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("github.com/bensigelman/opentracing-sandbox/donutsalon/public"))))
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}
