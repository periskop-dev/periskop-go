package periskop

import (
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Severity is the definition of different severities
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
	MaxTraces       int      = 4
	MaxErrors       int      = 10
)

type payload struct {
	AggregatedErrors []*aggregatedError `json:"aggregated_errors"`
}

type aggregatedError struct {
	AggregationKey string             `json:"aggregation_key"`
	TotalCount     int                `json:"total_count"`
	Severity       Severity           `json:"severity"`
	LatestErrors   []errorWithContext `json:"latest_errors"`
	mux            sync.Mutex
}

func newAggregatedError(aggregationKey string, severity Severity) aggregatedError {
	return aggregatedError{
		AggregationKey: aggregationKey,
		TotalCount:     0,
		Severity:       severity,
	}
}

func (e *aggregatedError) addError(errWithContext errorWithContext) {
	e.mux.Lock()
	if len(e.LatestErrors) >= MaxErrors {
		// dequeue
		e.LatestErrors = e.LatestErrors[1:]
	}
	e.LatestErrors = append(e.LatestErrors, errWithContext)
	e.TotalCount++
	e.mux.Unlock()
}

// HTTPContext holds info of the HTTP context when an error is produced
type HTTPContext struct {
	RequestMethod  string            `json:"request_method"`
	RequestURL     string            `json:"request_url"`
	RequestHeaders map[string]string `json:"request_headers"`
}

type errorWithContext struct {
	Error       errorInstance `json:"error"`
	UUID        uuid.UUID     `json:"uuid"`
	Timestamp   time.Time     `json:"timestamp"`
	Severity    Severity      `json:"severity"`
	HTTPContext HTTPContext   `json:"http_context"`
}

func newErrorWithContext(errInstance errorInstance, severity Severity, httpCtx HTTPContext) errorWithContext {
	return errorWithContext{
		Error:       errInstance,
		UUID:        uuid.New(),
		Timestamp:   time.Now().UTC(),
		Severity:    severity,
		HTTPContext: httpCtx,
	}
}

type errorInstance struct {
	Class      string         `json:"class"`
	Message    string         `json:"message"`
	Stacktrace []string       `json:"stacktrace"`
	Cause      *errorInstance `json:"cause"`
}

func newErrorInstance(err error, errType string, stacktrace []string) errorInstance {
	return errorInstance{
		Message:    err.Error(),
		Class:      errType,
		Stacktrace: stacktrace,
	}
}

// aggregationKey generates a hash for errorWithContext using the last MaxTraces
func (e *errorWithContext) aggregationKey() string {
	stacktraceHead := e.Error.Stacktrace
	if len(stacktraceHead) > MaxTraces {
		stacktraceHead = stacktraceHead[:MaxTraces]
	}
	stacktraceHeadHash := hash(e.Error.Message + strings.Join(stacktraceHead, ""))
	return fmt.Sprintf("%s@%s", e.Error.Class, stacktraceHeadHash)
}

func hash(s string) string {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		fmt.Printf("error hashing string '%s': %s\n", s, err)
	}
	return fmt.Sprintf("%x", h.Sum32())
}
