package exporter

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"encoding/json"
)

const (
	namespace = "logentries" //for Prometheus metrics.
)

var (
	y int
	m time.Month
)

// Structs
type jsonData struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PeriodUsage int    `json:"period_usage"`
}

// AccountStruct is return for Colletor prometheus
type AccountStruct struct {
	APIKEY      string
	REGION      string
	mutex       sync.Mutex
	client      *http.Client
	ID          *prometheus.Desc
	periodUsage *prometheus.Desc
}

// AccountGetUsage returns an initialized Exporter.
func AccountGetUsage(apikey string, region string) *AccountStruct {
	return &AccountStruct{
		APIKEY: apikey,
		REGION: region,
		periodUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "size_month_size_total"),
			"Account size month in bytes",
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

func monthInterval(y int, m time.Month) (firstDay, lastDay time.Time) {
	firstDay = time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	lastDay = time.Date(y, m+1, 1, 0, 0, 0, -1, time.UTC)
	return firstDay, lastDay
}

// Collect fetches the stats from configured location and delivers them
// as Prometheus metrics.
func (e *AccountStruct) collect(ch chan<- prometheus.Metric) error {
	log.Debugln("---------------------- SCRAPER ACCOUNT SIZE ----------------------------------")

	// Get first and last day of month
	y, m, _ = time.Now().Date()
	firstDayMonth, lastDayMonth := monthInterval(y, m)
	log.Debugf("Dates: %s - %s\n", firstDayMonth.Format("2006-01-02"), lastDayMonth.Format("2006-01-02"))

	// Create parse url per service
	parseURL := fmt.Sprintf("https://%s.rest.logs.insight.rapid7.com/usage/organizations?from=%s&to=%s", 
		e.REGION, firstDayMonth.Format("2006-01-02"), lastDayMonth.Format("2006-01-02"))
	log.Debugln(parseURL)

	responseAccount := requestHTTP(parseURL, e.APIKEY)
	defer responseAccount.Body.Close()
	var record jsonData
	if err := json.NewDecoder(responseAccount.Body).Decode(&record); err != nil {
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
