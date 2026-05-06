// Package metrics provides lightweight atomic counters and an HTTP handler
// for exposing operational statistics of the logslice proxy.
//
// Usage:
//
//	// Increment counters as logs flow through the proxy.
//	 metrics.Global.Received.Add(1)
//	 metrics.Global.Forwarded.Add(1)
//
//	// Mount the metrics endpoint on your mux.
//	 mux.Handle("/metrics", metrics.Handler(metrics.Global))
package metrics
