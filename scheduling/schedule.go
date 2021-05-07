package scheduling

import (
	"context"
	"time"
)

// Schedule calls function f with period of p offsetted by o.
//
// Similarly to cron a function with a period of two minutes
// will be executed every even minute, not every two minutes
// after initialisation.
//
// The parameter for f is the given context and the current
// time. Use them to detect cancellation and to implement
// extra filters eg. do not run on specific dates.
//
// The first n execution can be skipped with an offset
// greater then than the duration.
//
// Schedule operates on the time as an absolute duration since
// the zero time. Thus, Schedule(_, Hour, _, _) may call f at
// a non-zero minute, depending on the time's Location.
func Schedule(ctx context.Context, p time.Duration, o time.Duration, f func(context.Context, time.Time)) {
	// Position the first execution
	first := time.Now().Truncate(p).Add(o)
	if first.Before(time.Now()) {
		first = first.Add(p)
	}
	firstC := time.After(first.Sub(time.Now()))
	// Receiving from a nil channel blocks forever
	t := &time.Ticker{C: nil}
	for {
		select {
		case v := <-firstC:
			// The ticker has to be started before f
			t = time.NewTicker(p)
			f(ctx, v)
		case v := <-t.C:
			f(ctx, v)
		case <-ctx.Done():
			t.Stop()
			return
		}
	}
}
