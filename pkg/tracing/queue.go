package tracing

import (
	"context"
	"net/http"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// Inject extracts distributed trace headers from the active New Relic transaction
// in ctx and returns them as a flat map for embedding in a queue message payload.
// Returns nil when no active transaction exists.
func Inject(ctx context.Context) map[string]string {
	txn := newrelic.FromContext(ctx)
	if txn == nil {
		return nil
	}
	hdrs := http.Header{}
	txn.InsertDistributedTraceHeaders(hdrs)
	if len(hdrs) == 0 {
		return nil
	}
	meta := make(map[string]string, len(hdrs))
	for k, vals := range hdrs {
		if len(vals) > 0 {
			meta[k] = vals[0]
		}
	}
	return meta
}

// Accept starts a new New Relic transaction named txnName and links it to the
// distributed trace carried in meta (injected by the producer via Inject).
// Returns the enriched context and a done func the caller must defer.
// Safe to call with a nil app or empty meta — returns ctx unchanged with a no-op done.
func Accept(ctx context.Context, app *newrelic.Application, txnName string, meta map[string]string) (context.Context, func()) {
	if app == nil {
		return ctx, func() { /* no-op: APM not initialized */ }
	}
	txn := app.StartTransaction(txnName)
	if len(meta) > 0 {
		hdrs := make(http.Header, len(meta))
		for k, v := range meta {
			hdrs.Set(k, v)
		}
		txn.AcceptDistributedTraceHeaders(newrelic.TransportQueue, hdrs)
	}
	return newrelic.NewContext(ctx, txn), txn.End
}
