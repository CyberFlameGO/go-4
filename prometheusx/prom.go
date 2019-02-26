package prometheusx

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/pprof"
	"testing"

	"github.com/m-lab/go/httpx"
	"github.com/m-lab/go/rtx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/prometheus/util/promlint"
)

// MustStartPrometheus starts the prometheus server. If the address passed in is
// ":0" then a random open port will be chosen and the .Addr element of the
// returned server will be udpated to reflect the actual port. The server will
// run until the passed-in context is canceled.
func MustStartPrometheus(addr string) *http.Server {
	// Prometheus with some extras.
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/metrics", promhttp.Handler())

	// Start up the http server.
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	rtx.Must(httpx.ListenAndServeAsync(server), "Could not start metric server")

	// Return the server
	return server
}

// LintMetrics will ensure that the names of the passed-in promethus metrics
// follow all best practices. If the passed-in testing.T is nil, then all lint
// errors are just log messages. If a real testing.T is passed in, then lint
// errors cause test failures.
func LintMetrics(t *testing.T) {
	srv := MustStartPrometheus(":0")
	defer srv.Shutdown(context.Background())

	metricReader, err := http.Get("http://" + srv.Addr + "/metrics")
	rtx.Must(err, "Could not GET metrics")
	metricBytes, err := ioutil.ReadAll(metricReader.Body)
	rtx.Must(err, "Could not read metrics")

	metricsLinter := promlint.New(bytes.NewBuffer(metricBytes))
	problems, err := metricsLinter.Lint()
	rtx.Must(err, "Could not lint metrics")

	for _, p := range problems {
		msg := fmt.Sprintf("Bad metric %v: %v", p.Metric, p.Text)
		if t == nil {
			log.Println(msg)
		} else {
			t.Error(msg)
		}
	}
}
