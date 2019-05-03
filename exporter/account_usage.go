package exporter

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"encoding/json"
)

const (
	namespace = "logentries" //for Prometheus metrics.
)

// Structs
type jsonData struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	PeriodUsage int    `json:"period_usage"`
}

type AccountStruct struct {
	URI     string
	APIKEY  string
	SERVICE string
	mutex   sync.Mutex
	client  *http.Client

	Id          *prometheus.Desc
	periodUsage *prometheus.Desc
}

// NewExporter returns an initialized Exporter.
func AccountGetUsage(uri string, apikey string, service string) *AccountStruct {
	return &AccountStruct{
		URI:     uri,
		APIKEY:  apikey,
		SERVICE: service,
		periodUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "period_usage_daily"),
			"Account Usage Size in bytes",
			[]string{"account"},
			nil),
		client: &http.Client{},
	}
}

// Describe describes all the metrics ever exported by the logentries exporter. It
// implements prometheus.Collector.
func (e *AccountStruct) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.periodUsage
}

// Collect fetches the stats from configured location and delivers them
// as Prometheus metrics.
func (e *AccountStruct) collect(ch chan<- prometheus.Metric) error {
	log.Debugln("---------------------- SCRAPER ACCOUNT SIZE ----------------------------------")

	// Get current date
	currentTime := time.Now().Local().Format("2006-01-02")
	log.Debugln("Current Date:", currentTime)
	safeAccountID := url.QueryEscape(e.URI)
	var parseUrl string
	// Create parse url per service
	if e.SERVICE == "logentries" {
		parseUrl = fmt.Sprintf("https://rest.logentries.com/usage/accounts/%s?from=%s&to=%s", safeAccountID, currentTime, currentTime)
		log.Debugln(parseUrl)
	} else if e.SERVICE == "rapid7" {
		parseUrl = fmt.Sprintf("https://eu.rest.logs.insight.rapid7.com/usage/organizations?from=%s&to=%s", currentTime, currentTime)
		log.Debugln(parseUrl)
	}

	// Build the request
	url := parseUrl
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("x-api-key", e.APIKEY)
	if err != nil {
		log.Fatal("NewRequest: ", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Debugln("Account Status Code:", resp.StatusCode)
	} else {
		log.Errorln("Account Status Code:", resp.StatusCode)
	}
	defer resp.Body.Close()
	var record jsonData
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		log.Fatal(err)
	}

	ch <- prometheus.MustNewConstMetric(e.periodUsage, prometheus.GaugeValue, float64(record.PeriodUsage), record.Name)

	return nil
}

// Collect fetches the stats from configured logentries location and delivers them
// as Prometheus metrics.
// It implements prometheus.Collector.
func (e *AccountStruct) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	if err := e.collect(ch); err != nil {
		log.Errorf("Error scraping logentries: %s", err)
	}
	return
}
