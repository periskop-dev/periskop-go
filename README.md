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
	c.ReportWithHTTPContext(faultyJSONParser(), &periskop.HTTPContext{
		RequestMethod:  "GET",
		RequestURL:     "http://example.com",
		RequestHeaders: map[string]string{"Cache-Control": "no-cache"},
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

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md)
