package periskop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

func (e *ErrorExporter) export() ([]byte, error) {
	payload := e.collector.getAggregatedErrors()
	res, err := json.Marshal(payload)
	if err != nil {
		return []byte{}, err
	}
	return res, nil
}

// Export exports all collected errors in json format
func (e *ErrorExporter) Export() (string, error) {
	res, err := e.export()
	return string(res), err
}

func (e *ErrorExporter) PushToGateway(addr string) error {
	fmt.Println(e.Export())
	exportedData, err := e.export()
	if err == nil {
		_, err := http.Post(addr+"/errors", "application/json", bytes.NewBuffer(exportedData))
		if err == nil {
			return nil
		}
		//defer resp.Body.Close()
	}
	return err
}
