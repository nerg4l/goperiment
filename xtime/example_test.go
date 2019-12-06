package xtime_test

import (
	"context"
	"github.com/nerg4l/goperiment/xtime"
	"time"
)

func ExampleSchedule() {
	// Run f every hour.
	xtime.Schedule(context.Background(), time.Hour, 0, func(t time.Time) {
		// Do your job here
	})

	// Run f once a day at noon.
	xtime.Schedule(context.Background(), 24*time.Hour, 12*time.Hour, func(t time.Time) {
		// Do your job here
	})

	// Run f twice a day at 03:40 (00:00 + 03:40) and 15:40 (12:00 + 03:40).
	xtime.Schedule(context.Background(), 12*time.Hour, 3*time.Hour+40*time.Minute, func(t time.Time) {
		// Do your job here
	})

	// Filter can be implemented inside f.
	xtime.Schedule(context.Background(), time.Hour, 0, func(t time.Time) {
		if t.Weekday() == time.Wednesday {
			return
		}
		// Do your job here
	})

	// Schedule can be cancelled with a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	xtime.Schedule(ctx, time.Hour, 0, func(t time.Time) {
		// Do your job here
	})
	cancel()
}
