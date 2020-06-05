package periskop

import (
	"errors"
	"net/http"
	"testing"
)

func getFirstAggregatedErr(aggregatedErrors map[string]*aggregatedError) *aggregatedError {
	for _, value := range aggregatedErrors {
		return value
	}
	return nil
}

func TestCollector_addError(t *testing.T) {
	c := NewErrorCollector()
	err := errors.New("testing")
	c.addError(err, nil)

	if len(c.aggregatedErrors) != 1 {
		t.Errorf("expected one element")
	}

	c.addError(err, nil)
	if getFirstAggregatedErr(c.aggregatedErrors).TotalCount != 2 {
		t.Errorf("expected two elements")
	}
}

func TestCollector_Report(t *testing.T) {
	c := NewErrorCollector()
	err := errors.New("testing")
	c.Report(err)

	if len(c.aggregatedErrors) != 1 {
		t.Errorf("expected one element")
	}

	errorWithContext := getFirstAggregatedErr(c.aggregatedErrors).LatestErrors[0]
	if errorWithContext.Error.Message != err.Error() {
		t.Errorf("expected a propagated error")
	}

	if errorWithContext.Error.Class != "*errors.errorString" {
		t.Errorf("incorrect class name, got %s", errorWithContext.Error.Class)
	}

	if len(errorWithContext.Error.Stacktrace) == 0 {
		t.Errorf("expected a collected stack trace")
	}
}

func TestCollector_ReportWithHTTPContext(t *testing.T) {
	c := NewErrorCollector()
	body := "some body"
	err := errors.New("testing")
	httpContext := HTTPContext{
		RequestMethod:  "GET",
		RequestURL:     "http://example.com",
		RequestHeaders: map[string]string{"Cache-Control": "no-cache"},
		RequestBody:    &body,
	}
	c.ReportWithHTTPContext(err, &httpContext)

	if len(c.aggregatedErrors) != 1 {
		t.Errorf("expected one element")
	}

	errorWithContext := getFirstAggregatedErr(c.aggregatedErrors).LatestErrors[0]
	if errorWithContext.HTTPContext.RequestMethod != "GET" {
		t.Errorf("expected HTTP method GET")
	}

	if errorWithContext.Error.Class != "*errors.errorString" {
		t.Errorf("incorrect class name, got %s", errorWithContext.Error.Class)
	}
}

func TestCollector_ReportWithHTTPRequest(t *testing.T) {
	c := NewErrorCollector()
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = errors.New("testing")
	c.ReportWithHTTPRequest(err, req)

	if len(c.aggregatedErrors) != 1 {
		t.Errorf("expected one element")
	}

	errorWithContext := getFirstAggregatedErr(c.aggregatedErrors).LatestErrors[0]
	if errorWithContext.HTTPContext.RequestMethod != "GET" {
		t.Errorf("expected HTTP method GET")
	}

	if errorWithContext.HTTPContext.RequestBody != nil {
		t.Errorf("expected nil http request body but got %s", *errorWithContext.HTTPContext.RequestBody)
	}

	if errorWithContext.Error.Class != "*errors.errorString" {
		t.Errorf("incorrect class name, got %s", errorWithContext.Error.Class)
	}
}

func TestCollector_getAggregatedErrors(t *testing.T) {
	c := NewErrorCollector()
	err := errors.New("testing")
	c.addError(err, nil)

	aggregatedErr := getFirstAggregatedErr(c.aggregatedErrors)
	payload := c.getAggregatedErrors()
	if payload.AggregatedErrors[0].AggregationKey != aggregatedErr.AggregationKey {
		t.Errorf("keys for aggregated errors are different, expected: %s, got: %s",
			aggregatedErr.AggregationKey, payload.AggregatedErrors[0].AggregationKey)
	}
}

func TestCollector_getStackTrace(t *testing.T) {
	err := errors.New("testing")
	stacktrace := getStackTrace(err)
	if len(stacktrace) == 0 {
		t.Errorf("expected a  stacktrace")
	}
	lastFrame := stacktrace[len(stacktrace)-1]
	if lastFrame == "" {
		t.Errorf("got empty frame")
	}
}
