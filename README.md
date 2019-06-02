# Logentries/Rapid7 Exporter for Prometheus
Simple server that scrapes `logentries/rapid7` metrics endpoint and exports them as Prometheus metrics.

## Flags/Arguments
```
  --telemetry.address string
        Address on which to expose metrics. (default -> ":9582")
  --metricsPath string
        Path under which to expose metrics. (default -> "/metrics")
  --apikey string
        ApiKey to connect logentries metrics. (required)
  --isDebug bool
        Output verbose debug information. (default -> "false")
```

## Collectors
The exporter collects the following metrics:

**Metrics:**
```
# HELP logentries_size_month_size_total Account size month in bytes.
# TYPE logentries_size_month_size_total gauge
logentries_size_month_size_total{account="Your account name"} XXXXXX
# HELP logentries_log_usage_daily Log Usage Size in bytes (d-1).
# TYPE logentries_log_usage_daily gauge
logentries_log_usage_daily{logID="XXXXXX-XXX-XXX-XXX-XXXXXXXXX",logName="XXX-XXXXXXX",logSet="XXXXXXX"} XXXXX
...
```

## Building and running
```
$ go build
$ ./logentries_exporter --apikey xxxx-xxxx-xxxx-xxxx
```

## Contribute
Feel free to open an issue or PR if you have suggestions or ideas about what to add.
