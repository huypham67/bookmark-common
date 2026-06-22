package tracing

import (
	"context"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// StartSegment starts a named segment on the active New Relic transaction in ctx
// and returns a done func the caller must defer.
// Safe to call when no transaction is active — done becomes a no-op.
func StartSegment(ctx context.Context, name string) func() {
	txn := newrelic.FromContext(ctx)
	if txn == nil {
		return func() { /* no-op: no active transaction in context */ }
	}
	seg := txn.StartSegment(name)
	return seg.End
}
