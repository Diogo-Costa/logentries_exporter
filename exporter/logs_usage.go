package exporter

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	"encoding/json"
)

// Structs to getListLogs
type ListLogs struct {
	Logs []struct {
		LogsetsInfo []struct {
			ID    string `json:"id"`
			Links []struct {
				Href string `json:"href"`
				Rel  string `json:"rel"`
			} `json:"links"`
			Name string `json:"name"`
		} `json:"logsets_info"`
		Name     string `json:"name"`
		UserData struct {
			LeAgentFilename string `json:"le_agent_filename"`
			LeAgentFollow   string `json:"le_agent_follow"`
		} `json:"user_data"`
		Tokens     []interface{} `json:"tokens"`
		SourceType string        `json:"source_type"`
		TokenSeed  interface{}   `json:"token_seed"`
		Structures []interface{} `json:"structures"`
		ID         string        `json:"id"`
	} `json:"logs"`
}

// json data structure for logentries
type jsonLog struct {
	Usage struct {
		ID     string `json:"id"`
		Period struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"period"`
		DailyUsage []struct {
			Day      string `json:"day"`
			LogUsage string `json:"usage"`
		} `json:"daily_usage"`
	} `json:"usage"`
}

type LogStruct struct {
	URI     string
	APIKEY  string
	LOGID   string
	LOGNAME string
	mutex   sync.Mutex
	client  *http.Client

	up       *prometheus.Desc
	ID       *prometheus.Desc
	logName  *prometheus.Desc
	logUsage *prometheus.Desc
}

// LogGetUsage returns an initialized Exporter.
func LogGetUsage(uri string, apikey string) *LogStruct {
	return &LogStruct{
		URI:    uri,
		APIKEY: apikey,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could logentries be reached",
			nil,
			nil),
		logName: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "logName"),
			"Log Name",
			[]string{"account"},
			nil),
		logUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "log_usage_daily"),
			"Log Usage Size in bytes",
			[]string{"logname", "logset", "logid"},
			nil),
		client: &http.Client{},
	}
}

// Describe describes all the metrics ever exported by the logentries exporter. It
// implements prometheus.Collector.
func (e *LogStruct) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.logName
	ch <- e.logUsage
}

// Collect fetches the stats from configured location and delivers them
// as Prometheus metrics.
func (e *LogStruct) collect(ch chan<- prometheus.Metric) error {
	urlList := fmt.Sprintf("https://rest.logentries.com/management/logs")
	req, err := http.NewRequest("GET", urlList, nil)
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
		log.Debugln("Status Code:", resp.StatusCode)
	} else {
		log.Errorln("Status Code:", resp.StatusCode)
	}
	defer resp.Body.Close()
	var recordLogs ListLogs
	if err := json.NewDecoder(resp.Body).Decode(&recordLogs); err != nil {
		log.Fatal(err)
	}
	for rateList, logsList := range recordLogs.Logs {
		log.Infoln("---------------------- NEW REQUEST LOG ----------------------------------")
		for _, list := range logsList.LogsetsInfo {
			log.Debugln("RateLimit:", rateList)
			// Get current date
			currentTime := time.Now().Local().AddDate(0, 0, -1).Format("2006-01-02")
			safeAccountID := url.QueryEscape(e.URI)
			// Build the request
			url := fmt.Sprintf("https://rest.logentries.com/usage/accounts/%s/logs/%s/?from=%s&to=%s", safeAccountID, logsList.ID, currentTime, currentTime)
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
				log.Debugln("Status Code:", resp.StatusCode)
			} else {
				log.Errorln("Status Code:", resp.StatusCode)
			}
			defer resp.Body.Close()

			var record jsonLog
			if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
				log.Fatal(err)
			}
			log.Infoln("ID:", logsList.ID)
			log.Infoln("Name:", logsList.Name)
			log.Infoln("LogSet:", list.Name)
			if len(record.Usage.DailyUsage) != 0 {
				for _, logs := range record.Usage.DailyUsage {
					// Convert string in float64
					logSize, _ := strconv.ParseFloat(logs.LogUsage, 64)
					log.Debugln("SIZE:", logSize)
					ch <- prometheus.MustNewConstMetric(e.logUsage, prometheus.GaugeValue, float64(logSize), logsList.Name, list.Name, logsList.ID)
				}
			} else {
				log.Debugln("DEBUG:", record.Usage.DailyUsage)
				ch <- prometheus.MustNewConstMetric(e.logUsage, prometheus.GaugeValue, float64(0), logsList.Name, list.Name, logsList.ID)
			}
			if rateList == 400 {
				time.Sleep(240 * time.Second)
			} else if rateList == 800 {
				time.Sleep(240 * time.Second)
			} else if rateList == 1200 {
				time.Sleep(180 * time.Second)
			}
			rateList++
		}
	}
	return nil
}

// Collect fetches the stats from configured logentries location and delivers them
// as Prometheus metrics.
// It implements prometheus.Collector.
func (e *LogStruct) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	if err := e.collect(ch); err != nil {
		log.Errorf("Error scraping logentries: %s", err)
	}
	return
}
