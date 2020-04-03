package periskop

import (
	"errors"
	"testing"
)

var aggregationKeyCases = []struct {
	expectedAggregationKey string
	errorMessage           string
	stacktrace             []string
}{
	{"testingError@af88b146", "error", []string{""}},
	{"testingError@b8a5702f", "timeout error", []string{"line 0:", "error in testingError"}},
	{"testingError@b8a5702f", "index error", []string{"line 0:", "error in testingError"}},
	{"testingError@e1441017", "index error", []string{"line 0:", "index error", "line 1:", "line 4:", "checkTest()"}},
	{"testingError@df41ce5a", "index error", []string{"line 0:", "index error", "line 1:", "line 5:", "checkTest()"}},
	{"testingError@df41ce5a", "index error", []string{"line 0:", "index error", "line 1:", "line 5:", "checkFunc()"}},
}

func newMockErrorWithContext(stacktrace []string) errorWithContext {
	errorInstance := newErrorInstance(errors.New("divisin by zero"), "testingError", stacktrace)
	return newErrorWithContext(errorInstance, SeverityError, HTTPContext{})
}

func TestTypes_aggregationKey(t *testing.T) {
	for _, tt := range aggregationKeyCases {
		t.Run(tt.expectedAggregationKey, func(t *testing.T) {
			errorWithContext := newMockErrorWithContext(tt.stacktrace)
			resultAggregationKey := errorWithContext.aggregationKey()
			if resultAggregationKey != tt.expectedAggregationKey {
				t.Errorf("error in aggregationKey, expected: %s, got %s", tt.expectedAggregationKey, resultAggregationKey)
			}
		})
	}
}

func TestTypes_addError(t *testing.T) {
	errorWithContext := newMockErrorWithContext([]string{""})
	errorAggregate := newAggregatedError("error@hash", SeverityWarning)
	errorAggregate.addError(errorWithContext)
	if errorAggregate.TotalCount != 1 {
		t.Errorf("expected one error")
	}
	for i := 0; i < MaxErrors; i++ {
		errorAggregate.addError(errorWithContext)
	}
	if errorAggregate.TotalCount != MaxErrors+1 {
		t.Errorf("expected %v total errors", MaxErrors+1)
	}
	if len(errorAggregate.LatestErrors) != MaxErrors {
		t.Errorf("expected %v latest errors", MaxErrors)
	}
}
