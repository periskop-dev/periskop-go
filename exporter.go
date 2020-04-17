package periskop

import (
	"encoding/json"
)

// ErrorExporter exposes collected errors
type ErrorExporter struct {
	collector *ErrorCollector
}

// NewErrorExporter creates a new ErrorExporter
func NewErrorExporter(collector *ErrorCollector) ErrorExporter {
	return ErrorExporter{
		collector: collector,
	}
}

// Export exports all collected errors in json format
func (e *ErrorExporter) Export() (string, error) {
	payload := e.collector.getAggregatedErrors()
	res, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(res), nil
}
