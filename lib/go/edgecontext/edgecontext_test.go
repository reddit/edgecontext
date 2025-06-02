package edgecontext_test

import (
	"testing"

	"github.com/reddit/edgecontext/lib/go/edgecontext/edgecontexttest"
)

func TestDefaultBackend(t *testing.T) {
	runner := edgecontexttest.NewRunner(t, edgecontexttest.WithDefaultBackend())
	runner.Run(t)
}
