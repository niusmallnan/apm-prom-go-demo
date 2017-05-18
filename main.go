package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli"
)

var VERSION = "v0.0.0-dev"

func prometheusHandler() http.Handler {
	return prometheus.Handler()

}

func computeSum(body []byte) []byte {
	h := sha256.New()

	h.Write(body)
	hashed := h.Sum(nil)
	return hashed

}

func hashHandler(histogram *prometheus.HistogramVec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer r.Body.Close()
		code := 500

		defer func() { // Make sure we record a status.
			duration := time.Since(start)
			histogram.WithLabelValues(fmt.Sprintf("%d", code)).Observe(duration.Seconds())
		}()

		code = http.StatusBadRequest
		if r.Method == "POST" {
			code = http.StatusOK
			w.WriteHeader(code)
			body, _ := ioutil.ReadAll(r.Body)

			fmt.Printf("\"%s\"\n", string(body))

			hashed := computeSum(body)
			val := fmt.Sprintf("%x\n", hashed)
			w.Write([]byte(val))

		} else {
			w.WriteHeader(code)

		}

	}

}

func appMain(ctx *cli.Context) error {
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "hash_seconds",
		Help: "Time taken to create hashes",
	}, []string{"code"})

	r := mux.NewRouter()
	r.Handle("/metrics", prometheusHandler())
	r.Handle("/hash", hashHandler(histogram))

	prometheus.Register(histogram)

	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    8 * time.Second,
		WriteTimeout:   8 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        r,
	}
	logrus.Fatal(s.ListenAndServe())

}

func main() {
	app := cli.NewApp()
	app.Name = "apm-prom-go-demo"
	app.Version = VERSION
	app.Usage = "You need help!"
	app.Action = appMain
	app.Run(os.Args)
}
