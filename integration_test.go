package periskop

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func errFunc() error {
	var dat map[string]interface{}
	// will return "unexpected end of JSON input"
	return json.Unmarshal([]byte(`{"id":`), &dat)
}

func parseJSON(exportedErrors string) payload {
	p := payload{}
	json.Unmarshal([]byte(exportedErrors), &p)
	return p
}

func TestHandler(t *testing.T) {
	body := "some body"
	c := NewErrorCollector()
	c.Report(errFunc())
	c.ReportWithHTTPContext(errFunc(), &HTTPContext{
		RequestMethod:  "GET",
		RequestURL:     "http://example.com",
		RequestHeaders: map[string]string{"Cache-Control": "no-cache"},
		RequestBody:    &body,
	})

	e := NewErrorExporter(&c)
	h := NewHandler(e)
	req, err := http.NewRequest("GET", "/-/exceptions", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	p := parseJSON(rr.Body.String())
	if p.AggregatedErrors[0].TotalCount != 2 {
		t.Errorf("wrong number of exceptions collected: %s", rr.Body.String())
	}
}

func TestConcurrency(t *testing.T) {
	const maxGoRoutines = 10
	const maxIterations = 20

	c := NewErrorCollector()
	e := NewErrorExporter(&c)
	var wg sync.WaitGroup
	wg.Add(maxGoRoutines)
	for i := 0; i < maxGoRoutines; i++ {
		go func() {
			defer wg.Done()

			for i := 0; i < maxIterations; i++ {
				c.Report(errFunc())
			}
		}()
		e.Export()
	}
	wg.Wait()

	s, _ := e.Export()
	p := parseJSON(s)
	if p.AggregatedErrors[0].TotalCount != maxGoRoutines*maxIterations {
		t.Errorf("num total errors expected %d, got %d", maxGoRoutines*maxIterations,
			p.AggregatedErrors[0].TotalCount)
	}
}
