package exporter

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"encoding/json"
)

// Structs
type responseLogsStruct struct {
	PerDayUsage struct {
		Period struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"period"`
		ReportInterval string `json:"report_interval"`
		UsageUnits     string `json:"usage_units"`
		Usage          []struct {
			Interval string `json:"interval"`
			LogUsage []struct {
				ID    string `json:"id"`
				Usage int    `json:"usage"`
			} `json:"log_usage"`
		} `json:"usage"`
	} `json:"per_day_usage"`
}

//
type responseLogStesStruct struct {
	Logs []struct {
		ID         string        `json:"id"`
		Name       string        `json:"name"`
		Tokens     []string      `json:"tokens"`
		Structures []interface{} `json:"structures"`
		UserData   struct {
		} `json:"user_data,omitempty"`
		SourceType      string      `json:"source_type"`
		TokenSeed       interface{} `json:"token_seed"`
		RetentionPeriod string      `json:"retention_period"`
		Links           []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
		LogsetsInfo []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Links []struct {
				Rel  string `json:"rel"`
				Href string `json:"href"`
			} `json:"links"`
		} `json:"logsets_info"`
	} `json:"logs"`
}

// LogsStruct is return for Colletor prometheus
type LogsStruct struct {
	APIKEY        string
	mutex         sync.Mutex
	client        *http.Client
	periodUsage   *prometheus.Desc
	periodUsageUp *prometheus.Desc
}

// LogsGetUsage returns an initialized Exporter.
func LogsGetUsage(apikey string) *LogsStruct {
	return &LogsStruct{
		APIKEY: apikey,
		periodUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "log_usage_daily"),
			"Log Usage Size in bytes (d-1)",
			[]string{"logID", "logName", "logSet"},
			nil),
		periodUsageUp: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "log_usage_up"),
			"Was the last scrape of log usage Size successful (0-successful / 1-Fail)",
			[]string{},
			nil),
		client: &http.Client{},
	}
}

// Describe describes all the metrics ever exported by the logentries exporter. It
// implements prometheus.Collector.
func (e *LogsStruct) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.periodUsage
}

// Collect fetches the stats from configured location and delivers them
// as Prometheus metrics.
func (e *LogsStruct) collect(ch chan<- prometheus.Metric) error {
	log.Debugln("---------------------- SCRAPER LOGS SIZE ----------------------------------")

	// Create parse url to capture all sizes logs
	parseLogsURL := "https://eu.rest.logs.insight.rapid7.com/usage/organizations/logs?time_range=yesterday&interval=day"
	log.Debugln(parseLogsURL)
	responseLogs := requestHTTP(parseLogsURL, e.APIKEY)
	defer responseLogs.Body.Close()
	var recordLogsStruct responseLogsStruct
	if err := json.NewDecoder(responseLogs.Body).Decode(&recordLogsStruct); err != nil {
		log.Fatal(err)
	}

	// Create parse url to capture all logs in account
	parseLogSetsURL := "https://eu.rest.logs.insight.rapid7.com/management/logs"
	log.Debugln(parseLogSetsURL)
	responseLogStes := requestHTTP(parseLogSetsURL, e.APIKEY)
	defer responseLogStes.Body.Close()
	var recordLogStesStruct responseLogStesStruct
	if err := json.NewDecoder(responseLogStes.Body).Decode(&recordLogStesStruct); err != nil {
		log.Fatal(err)
	}

	// Check if return dont is empty
	if len(recordLogsStruct.PerDayUsage.Usage) == 0 {
		ch <- prometheus.MustNewConstMetric(e.periodUsageUp, prometheus.GaugeValue, 0)
	} else {
		for _, logs := range recordLogsStruct.PerDayUsage.Usage {
			for _, logUsage := range logs.LogUsage {
				for _, logName := range recordLogStesStruct.Logs {
					for _, logSetName := range logName.LogsetsInfo {
						if logUsage.ID == logName.ID {
							log.Debugln("LogStet: ", logSetName.Name)
							log.Debugln("ID: ", logUsage.ID)
							log.Debugln("Name: ", logName.Name)
							log.Debugln("Size: ", logUsage.Usage)
							ch <- prometheus.MustNewConstMetric(e.periodUsage, prometheus.GaugeValue, float64(logUsage.Usage), logUsage.ID, logName.Name, logSetName.Name)
						}
					}
				}
			}
		}
	}
	return nil
}

// Collect fetches the stats from configured logentries location and delivers them
// as Prometheus metrics.
// It implements prometheus.Collector.
func (e *LogsStruct) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	if err := e.collect(ch); err != nil {
		log.Errorf("Error scraping logentries: %s", err)
	}
	return
}
