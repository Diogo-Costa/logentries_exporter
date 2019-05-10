package exporter

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Build the request
func requestHTTP(url, apikey string) *http.Response {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("x-api-key", apikey)
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

	return resp
}
