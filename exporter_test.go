package periskop

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func compareJSON(json0, json1 string) (bool, error) {
	var obj0 interface{}
	var obj1 interface{}

	var err error
	err = json.Unmarshal([]byte(json0), &obj0)
	if err != nil {
		return false, fmt.Errorf("Error mashalling json 0 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(json1), &obj1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling json 1 :: %s", err.Error())
	}

	return reflect.DeepEqual(obj0, obj1), nil
}

func TestExporter_Export(t *testing.T) {
	c := NewErrorCollector()
	uuid, _ := uuid.Parse("5d9893c6-51d6-11ea-8aad-f894c260afe5")
	c.uuid = uuid
	errWithContext := ErrorWithContext{
		Error: ErrorInstance{
			Class:      errors.New("testing").Error(),
			Stacktrace: []string{"line 12:", "syntax error"},
		},
		UUID:      uuid,
		Timestamp: time.Date(2020, 2, 17, 22, 42, 45, 0, time.UTC),
		Severity:  SeverityError,
		HTTPContext: &HTTPContext{
			RequestMethod:  "GET",
			RequestURL:     "http://example.com",
			RequestHeaders: map[string]string{"Cache-Control": "no-cache"},
			RequestBody:    nil,
		},
	}

	errorAggregate := aggregatedError{
		AggregationKey: "test",
		TotalCount:     1,
		Severity:       SeverityError,
		LatestErrors:   []ErrorWithContext{errWithContext},
		CreatedAt:      time.Date(2020, 2, 17, 22, 42, 45, 0, time.UTC),
	}
	var expected = `{
		"target_uuid": "5d9893c6-51d6-11ea-8aad-f894c260afe5",
		"aggregated_errors":[
		  {
			"aggregation_key":"test",
			"total_count":1,
			"severity":"error",
			"created_at":"2020-02-17T22:42:45Z",
			"latest_errors":[
			  {
				"error":{
				  "class":"testing",
				  "message":"",
				  "stacktrace":[
					"line 12:",
					"syntax error"
				  ],
				  "cause":null
				},
				"uuid":"5d9893c6-51d6-11ea-8aad-f894c260afe5",
				"timestamp":"2020-02-17T22:42:45Z",
				"severity":"error",
				"http_context":{
				  "request_method":"GET",
				  "request_url":"http://example.com",
				  "request_headers":{
					"Cache-Control":"no-cache"
				  },
				  "request_body": null
				}
			  }
			]
		  }
		]
	  }`
	c.aggregatedErrors["test"] = &errorAggregate
	e := NewErrorExporter(&c)
	data, err := e.Export()
	if err != nil {
		t.Errorf("error exporting exceptions: %v", err)
	}

	areEqual, err := compareJSON(data, expected)
	if err != nil {
		t.Errorf("error exporting exceptions: %v", err)
	}
	if !areEqual {
		t.Errorf("data did not match:\nexpected: %s\ngot: %s", expected, data)
	}
}
