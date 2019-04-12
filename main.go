package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/logentries_exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

// declare variables for logentries metrics
var (
	listeningAddress = flag.String("telemetry.address", ":9234", "Address on which to expose metrics.")
	metricsPath      = flag.String("telemetry.endpoint", "/metrics", "Path under which to expose metric.")
	logentriesID     = flag.String("logentriesID", "", "ID Account to logentries metrics")
	apikey           = flag.String("apikey", "", "APIKEY to connect logentries metrics")
	showVersion      = flag.Bool("version", false, "Print version information.")
)

func main() {
	log.Infoln("Starting logentries_exporter", version.Info())

	flag.Parse()

	if *logentriesID == "" && *apikey == "" {
		log.Fatal("Cannot specify both logentriesID and apikey")
	}

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("logentries_exporter"))
		os.Exit(0)
	}

	exporter := exporter.AccountGetUsage(*logentriesID, *apikey)
	prometheus.MustRegister(exporter)
	prometheus.MustRegister(version.NewCollector("logentries_exporter"))

	// setup and start webserver
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>Logentries Exporter</title></head>
		<body>
		<h1>Logentries Exporter</h1>
		<p><a href="` + *metricsPath + `">Metrics</a></p>
		</body>
		</html>
		`))
	})
	log.Infoln("Build context", version.BuildContext())

	log.Infof("Starting Server in: %s", *listeningAddress)
	log.Fatal(http.ListenAndServe(*listeningAddress, nil))
}
