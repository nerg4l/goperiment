package schedule

import (
	"context"
	"errors"
	"time"
)

// A Schedule holds a channel that is triggered every p period
// like a time.Ticker. Unlike time.Timer, it can have an offset,
// and it is aligned with the UNIX epoch time to make it predictable.
type Schedule struct {
	C    <-chan time.Time
	p, o time.Duration
}

// NewSchedule returns a new Schedule containing a channel that will send
// the current time on the channel after each tick. The period of the
// ticks is specified by the duration argument. The schedule will adjust
// the time interval or drop ticks to make up for slow receivers.
// The duration d must be greater than zero; if not, NewSchedule will
// panic.
//
// Similarly to cron a function with a period of two minutes
// will be executed every even minute, not every two minutes
// after initialisation.
//
// Cancel ctx to release associated resources.
func NewSchedule(ctx context.Context, p time.Duration, o time.Duration) *Schedule {
	if p <= 0 {
		panic(errors.New("non-positive interval for NewSchedule"))
	}

	ch := make(chan time.Time)
	s := Schedule{C: ch}
	// Position the first execution
	firstT := time.Now().Truncate(p).Add(o)
	if firstT.Before(time.Now()) {
		firstT = firstT.Add(p)
	}
	first := time.NewTimer(firstT.Sub(time.Now()))
	// Receiving from a nil channel blocks forever
	t := &time.Ticker{C: nil}
	go func() {
		for {
			select {
			case v := <-first.C:
				// The ticker has to be started before f
				t = time.NewTicker(p)
				ch <- v
			case v := <-t.C:
				ch <- v
			case <-ctx.Done():
				if t.C == nil {
					if !first.Stop() {
						<-first.C
					}
				} else {
					t.Stop()
				}
				return
			}
		}
	}()
	return &s
}

// Period returns the period of s.
func (s Schedule) Period() time.Duration {
	return s.p
}

// Offset returns the offset of s.
func (s Schedule) Offset() time.Duration {
	return s.o
}
