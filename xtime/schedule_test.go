package xtime

import (
	"context"
	"sync"
	"testing"
	"time"
)

const inaccuracy = 17 * time.Millisecond

func TestSchedule(t *testing.T) {
	t.Parallel()
	type args struct {
		p time.Duration
		o time.Duration
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: `without offset`,
			args: args{
				p: time.Second,
				o: 0,
			},
		},
		{
			name: `with offset`,
			args: args{
				p: time.Second,
				o: time.Second / 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w sync.WaitGroup
			w.Add(2)

			var triggers []time.Time
			ctx, cancel := context.WithCancel(context.Background())
			Schedule(ctx, tt.args.p, tt.args.o, func(t time.Time) {
				triggers = append(triggers, t)
				w.Done()
			})

			w.Wait()
			cancel()

			first := triggers[0]
			second := triggers[1]

			gotPeriod := second.Sub(first)
			if expected := tt.args.p - inaccuracy; gotPeriod < expected {
				t.Fatalf("Schedule(%s, %s, func) period expect >= %d ns, got %d ns", tt.args.p, tt.args.o, expected, gotPeriod)
			}
			if expected := tt.args.p + inaccuracy; gotPeriod > expected {
				t.Fatalf("Schedule(%s, %s, func) period expect <= %d ns, got %d ns", tt.args.p, tt.args.o, expected, gotPeriod)
			}
			if min := first.Add(tt.args.p - inaccuracy); second.Before(min) {
				t.Fatalf("Schedule(%s, %s, func) period expect > %s, got %s", tt.args.p, tt.args.o, min, second)
			}
			i := first.Nanosecond() % int(tt.args.p)
			gotOffset := time.Duration(i)
			if expected := tt.args.o + inaccuracy; gotOffset > expected {
				t.Fatalf("Schedule(%s, %s, func) offset expect <= %d ns, got %d ns", tt.args.p, tt.args.o, expected, gotOffset)
			}
			if expected := tt.args.o - inaccuracy; gotOffset < expected {
				t.Fatalf("Schedule(%s, %s, func) offset expect >= %d ns, got %d ns", tt.args.p, tt.args.o, expected, gotOffset)
			}
		})
	}
}
