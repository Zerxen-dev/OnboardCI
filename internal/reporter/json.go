package reporter

import (
	"encoding/json"

	"github.com/Zerxen-dev/onboardci/internal/runner"
)

// JSONReporter emits serialized JSON output.
type JSONReporter struct {
	Pretty bool
}

func NewJSONReporter(pretty bool) *JSONReporter {
	return &JSONReporter{Pretty: pretty}
}

func (r *JSONReporter) Format(report *runner.ExecutionReport) (string, error) {
	var bytes []byte
	var err error

	if r.Pretty {
		bytes, err = json.MarshalIndent(report, "", "  ")
	} else {
		bytes, err = json.Marshal(report)
	}

	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
