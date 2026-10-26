package reporter

import (
	"github.com/Zerxen-dev/onboardci/internal/runner"
)

// Reporter formats execution reports into desired representation.
type Reporter interface {
	Format(report *runner.ExecutionReport) (string, error)
}
