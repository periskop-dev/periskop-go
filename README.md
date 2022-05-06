# periskop-go

[![Build Status](https://api.cirrus-ci.com/github/periskop-dev/periskop-go.svg)](https://cirrus-ci.com/github/periskop-dev/periskop-go)

[Periskop](https://github.com/periskop-dev/periskop) requires collecting and aggregating exceptions on the client side,
as well as exposing them via an HTTP endpoint using a well defined format.

This library provides low level collection and rendering capabilities

## Usage

```
go get github.com/periskop-dev/periskop-go
```

### Example

```go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/periskop-dev/periskop-go"
)

func faultyJSONParser() error {
	var dat map[string]interface{}
	// will return "unexpected end of JSON input"
	return json.Unmarshal([]byte(`{"id":`), &dat)
}

func main() {
	c := periskop.NewErrorCollector()

	// Without context
	c.ReportError(faultyJSONParser())

	// Optionally pass Severity of an error (supported by all report methods)
	c.ReportWithSeverity(faultyJSONParser(), periskop.SeverityInfo)

	// With HTTP context
	body := "some body"
	c.ReportWithHTTPContext(faultyJSONParser(), &periskop.HTTPContext{
		RequestMethod:  "GET",
		RequestURL:     "http://example.com",
		RequestHeaders: map[string]string{"Cache-Control": "no-cache"},
		RequestBody:    &body, // optional request body, nil if not present
	})

	// With http.Request
	req, err := http.NewRequest("GET", "http://example.com", nil)
	c.ReportWithHTTPRequest(err, req)

	// With a full error report
	c.Report(periskop.ErrorReport{
		err:      err,
		severity: SeverityWarning,
		httpCtx: &periskop.HTTPContext{
			RequestMethod:  "GET",
			RequestURL:     "http://example.com",
			RequestHeaders: map[string]string{"Cache-Control": "no-cache"},
			RequestBody:    &body,
		},
		errKey: "json-parsing", // Overrides the errors aggregation key (see more info below)
	})

	// With a full error report, but with http.Request instead of HTTP context
	c.Report(periskop.ErrorReport{
		err:         err,
		severity:    SeverityWarning,
		httpRequest: req,
		errKey:      "json-parsing",
	})

	// Call the exporter and HTTP handler to expose the
	// errors in /-/exceptions endpoints
	e := periskop.NewErrorExporter(&c)
	h := periskop.NewHandler(e)
	http.Handle("/-/exceptions", h)
	http.ListenAndServe(":8080", nil)
}
```

### Custom aggregation for reported errors

By default errors are aggregated by their _stack trace_ and _error message_. This might cause that errors that are the same (but with different message) are treated as different in Periskop:

```
*url.Error@efdca928 -> Get "http://example": dial tcp 10.10.10.1:10100: i/o timeout
*url.Error@824c748e -> Get "http://example": dial tcp 10.10.10.2:10100: i/o timeout
```

To avoid that, you can manually group errors specifying the error key that you want to use:

```go
func main() {
	c := periskop.NewErrorCollector()
	req, err := http.NewRequest("GET", "http://example.com", nil)
	c.Report(periskop.ErrorReport{
		err:         err,
		httpRequest: req,
		errKey:      "example-request-error",
	})
}
```

### Using push gateway

You can also use [pushgateway](https://github.com/periskop-dev/periskop-pushgateway) in case you want to push your metrics instead of using pull method. Use only in case you really need it (e.g a batch job) as it would degrade the performance of your application. In the following example, we assume that we deployed an instance of periskop-pushgateway on `http://localhost:6767`:

```go
package main

import (
	"encoding/json"
	"github.com/periskop-dev/periskop-go"
)

func faultyJSONParser() error {
	var dat map[string]interface{}
	// will return "unexpected end of JSON input"
	return json.Unmarshal([]byte(`{"id":`), &dat)
}

func reportAndPush(c *periskop.ErrorCollector, e *periskop.ErrorExporter, err error) error {
  c.ReportError(err)
  return e.PushToGateway("http://localhost:6767")
}

func main() {
	c := periskop.NewErrorCollector()
	e := periskop.NewErrorExporter(&c)

	reportAndPush(&c, &e, faultyJSONParser())
}
```

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md)
