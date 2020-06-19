# periskop-go

[![Build Status](https://api.cirrus-ci.com/github/soundcloud/periskop-go.svg)](https://cirrus-ci.com/github/soundcloud/periskop-go)

[Periskop](https://github.com/soundcloud/periskop) requires collecting and aggregating exceptions on the client side,
as well as exposing them via an HTTP endpoint using a well defined format.

This library provides low level collection and rendering capabilities

## Usage

```
go get github.com/soundcloud/periskop-go
```

### Example

```go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/soundcloud/periskop-go"
)

func faultyJSONParser() error {
	var dat map[string]interface{}
	// will return "unexpected end of JSON input"
	return json.Unmarshal([]byte(`{"id":`), &dat)
}

func main() {
	c := periskop.NewErrorCollector()

	// Without context
	c.Report(faultyJSONParser())

	// With HTTP context
	var body string := "some body"
	c.ReportWithHTTPContext(faultyJSONParser(), &periskop.HTTPContext{
		RequestMethod:  "GET",
		RequestURL:     "http://example.com",
		RequestHeaders: map[string]string{"Cache-Control": "no-cache"},
		RequestBody:	&body // optional request body, nil if not present
	})

	// With http.Request
	req, err := http.NewRequest("GET", "http://example.com", nil)
	c.ReportWithHTTPRequest(err, req)

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
	c.ReportWithHTTPRequest(err, req, "example-request-error")
}
```
__Note:__ With this method you are also aggregating by _error class_ which means that for the previous example the aggregation key is `*url.Error@example-request-error`.

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md)
