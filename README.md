# Logentries Exporter for Prometheus
Simple server that scrapes `logentries` metrics endpoint and exports them as Prometheus metrics.

## Flags/Arguments
```
  --telemetry.address string
    	Address on which to expose metrics. (default -> ":9578")
  --metricsPath string
      Path under which to expose metrics. (default -> "/metrics")
  --logentriesID string
      ID Logentries account for scraper. (required)
  --apikey string
      ApiKey to connect logentries metrics. (required)
  --service string
      Select service used in scraper (required) (options: `logentries` or `rapid7`)
  --isDebug bool
    	Output verbose debug information. (default -> "false")
```

## Collectors
The exporter collects the following metrics:

**Metrics:**
```
# HELP logentries_period_usage_daily Account Usage Size in bytes.
# TYPE logentries_period_usage_daily gauge
logentries_period_usage_daily{account="Your account name"} XXXXXX
...
```

## Building and running
```
$ go build
$ ./logentries_exporter --logentriesID xxxx-xxxx-xxxx-xxxx --apikey xxxx-xxxx-xxxx-xxxx --service XXXX
```

## Contribute
Feel free to open an issue or PR if you have suggestions or ideas about what to add.

## TODO
- Create suite test 
- Create functions to check size per log/logset.