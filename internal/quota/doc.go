// Package quota enforces per-source byte and entry quotas over a rolling
// time window.
//
// # Overview
//
// A [Limiter] is created with a [Config] that specifies maximum bytes,
// maximum entries, and the duration of the rolling window. Each call to
// [Limiter.Allow] atomically increments the counters for the named source
// and returns [ErrQuotaExceeded] when either limit is breached. Counters
// are automatically reset once the window expires.
//
// # Usage
//
//	l, _ := quota.New(quota.Config{
//		MaxBytes:   5 * 1024 * 1024, // 5 MiB per minute
//		MaxEntries: 5_000,
//		Window:     time.Minute,
//	})
//
//	if err := l.Allow(sourceID, int64(len(body))); err != nil {
//		http.Error(w, "quota exceeded", http.StatusTooManyRequests)
//		return
//	}
package quota
