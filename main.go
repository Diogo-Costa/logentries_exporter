package main

import (
	"flag"
	"net/http"

	"github.com/logentries_exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"

	log "github.com/sirupsen/logrus"
)

// declare variables for logentries metrics
var (
	listeningAddress = flag.String("telemetry.address", ":9582", "Address on which to expose metrics.")
	metricsPath      = flag.String("telemetry.endpoint", "/metrics", "Path under which to expose metric.")
	apikey           = flag.String("apikey", "", "APIKEY to connect logentries metrics.")
	region           = flag.String("region", "us", "Region logentries (us, eu, ca or au).")
	isDebug          = flag.String("isDebug", "false", "Output verbose debug information.")
)

func main() {
	flag.Parse()

	if *isDebug == "true" {
		log.SetLevel(log.DebugLevel)
		log.Debugln("Enabling debug output")
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if *apikey == "" {
		log.Fatal("Cannot specify both apikey")
	}

	log.Infoln("Starting logentries_exporter", version.Info())

	// Scraper AccountUsage
	prometheus.MustRegister(exporter.AccountGetUsage(*apikey,*region))

	// Scraper LogsUsage
	prometheus.MustRegister(exporter.LogsGetUsage(*apikey,*region))

	// Scraper Exporter version
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

	log.Infoln("Starting Server in port", *listeningAddress)
	log.Fatal(http.ListenAndServe(*listeningAddress, nil))
}
