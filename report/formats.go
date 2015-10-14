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
	Counter:        "%s.%s.count %d %d\n",
	HistogramCount: "%s.%s.count %d %d\n",
	Gauge:          "%s.%s.value %d %d\n",
	GaugeFloat64:   "%s.%s.value %f %d\n",
	Min:            "%s.%s.min %d %d\n",
	Max:            "%s.%s.max %d %d\n",
	Mean:           "%s.%s.mean %.2f %d\n",
	Stddev:         "%s.%s.std-dev %.2f %d\n",
	Percentile:     "%s.%s.%s-percentile %.2f %d\n",
	Rate1:          "%s.%s.one-minute %.2f %d\n",
	Rate5:          "%s.%s.five-minute %.2f %d\n",
	Rate15:         "%s.%s.fifteen-minute %.2f %d\n",
}

// An alternate export format that formats percentile paths more like twitter's ostrich.
var OstrichFormats = ExportFormatStrings{
	Counter:        "%s.%s.count %d %d\n",
	HistogramCount: "%s.%s.count %d %d\n",
	Gauge:          "%s.%s.value %d %d\n",
	GaugeFloat64:   "%s.%s.value %f %d\n",
	Min:            "%s.%s.min %d %d\n",
	Max:            "%s.%s.max %d %d\n",
	Mean:           "%s.%s.mean %.2f %d\n",
	Stddev:         "%s.%s.std-dev %.2f %d\n",
	Percentile:     "%s.%s.percentiles.p%s %.2f %d\n",
	Rate1:          "%s.%s.one-minute %.2f %d\n",
	Rate5:          "%s.%s.five-minute %.2f %d\n",
	Rate15:         "%s.%s.fifteen-minute %.2f %d\n",
}

var ExportFormats = GoMetricsFormats
