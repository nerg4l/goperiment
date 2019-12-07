package xtime

import (
	"context"
	"time"
)

// Schedule calls function f with period of p offsetted by o.
// Similarly to cron a function schedules to one minute will
// be executed every whole minute, not every minute after
// initialisation.
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
	go func() {
		first := time.Now().Truncate(p).Add(o)
		if first.Before(time.Now()) {
			first = first.Add(p)
		}
		select {
		case v := <-time.After(first.Sub(time.Now())):
			go f(v)
		case <-ctx.Done():
			return
		}
		t := time.NewTicker(p)
		for {
			select {
			case v := <-t.C:
				go f(v)
			case <-ctx.Done():
				t.Stop()
				return
			}
		}
	}()
}
