package xtime

import (
	"context"
	"time"
)

// Schedule calls function f with period of p offsetted by o.
// Similarly to cron a function with a period of two minutes
// will be executed every even minute, not every two minutes
// after initialisation.
//
// The parameter for f is the given context and the current
// time. Use them to detect cancellation and to implement
// extra filters eg. do not run on specific weekday.
//
// The first n execution can be skipped with an offset
// greater then than the duration.
//
// F is executed in a goroutine which means multiple job can
// be executed in the same time.
func Schedule(ctx context.Context, p time.Duration, o time.Duration, f func(context.Context, time.Time)) {
	go func() {
		first := time.Now().Truncate(p).Add(o)
		if first.Before(time.Now()) {
			first = first.Add(p)
		}
		select {
		case v := <-time.After(first.Sub(time.Now())):
			go f(ctx, v)
		case <-ctx.Done():
			return
		}
		t := time.NewTicker(p)
		for {
			select {
			case v := <-t.C:
				go f(ctx, v)
			case <-ctx.Done():
				t.Stop()
				return
			}
		}
	}()
}
