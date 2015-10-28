package report

type ExportFormatStrings struct {
	Counter        string
	HistogramCount string
	Gauge          string
	GaugeFloat64   string
	Min            string
	Max            string
	Mean           string
	Stddev         string
	Percentile     string
	Rate1          string
	Rate5          string
	Rate15         string
}

var GoMetricsFormats = ExportFormatStrings{
	Counter:        "%s.%s.count %d %s\n",
	HistogramCount: "%s.%s.count %d %s\n",
	Gauge:          "%s.%s.value %d %s\n",
	GaugeFloat64:   "%s.%s.value %f %s\n",
	Min:            "%s.%s.min %d %s\n",
	Max:            "%s.%s.max %d %s\n",
	Mean:           "%s.%s.mean %.2f %s\n",
	Stddev:         "%s.%s.std-dev %.2f %s\n",
	Percentile:     "%s.%s.%s-percentile %.2f %s\n",
	Rate1:          "%s.%s.one-minute %.2f %s\n",
	Rate5:          "%s.%s.five-minute %.2f %s\n",
	Rate15:         "%s.%s.fifteen-minute %.2f %s\n",
}

// An alternate export format that formats percentile paths more like twitter's ostrich.
var OstrichFormats = ExportFormatStrings{
	Counter:        "%s.%s.count %d %s\n",
	HistogramCount: "%s.%s.count %d %s\n",
	Gauge:          "%s.%s.value %d %s\n",
	GaugeFloat64:   "%s.%s.value %f %s\n",
	Min:            "%s.%s.min %d %s\n",
	Max:            "%s.%s.max %d %s\n",
	Mean:           "%s.%s.mean %.2f %s\n",
	Stddev:         "%s.%s.std-dev %.2f %s\n",
	Percentile:     "%s.%s.percentiles.p%s %.2f %s\n",
	Rate1:          "%s.%s.one-minute %.2f %s\n",
	Rate5:          "%s.%s.five-minute %.2f %s\n",
	Rate15:         "%s.%s.fifteen-minute %.2f %s\n",
}

var ExportFormats = GoMetricsFormats
