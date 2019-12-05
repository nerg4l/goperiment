package xtime

import (
	"context"
	"time"
)

// Schedule calls function f with period of p offsetted by o.
// Similarly to cron a function schedules to one minute will
// be executed every whole minute,  every minute after
// initialization.
//
// The parameter for f is the current time. Use this to
// implement extra filters eg. do not run on specific weekday
//
// It is not recommended to have a greater offset than the
// duration except if you want to skip the first n execution.
//
// F is executed in a goroutine which means multiple job can
// be executed in the same time.
func Schedule(ctx context.Context, p time.Duration, o time.Duration, f func(time.Time)) {
	next := time.Now().Truncate(p).Add(o)
	if next.Before(time.Now()) {
		next = next.Add(p)
	}

	t := time.NewTimer(next.Sub(time.Now()))

	go func() {
		for {
			select {
			case v := <-t.C:
				next = next.Add(p)
				t.Reset(next.Sub(time.Now()))
				go f(v)
			case <-ctx.Done():
				if !t.Stop() {
					<-t.C
				}
				return
			}
		}
	}()
}
