package scheduling_test

import (
	"context"
	"github.com/nerg4l/goperiment/scheduling"
	"time"
)

func ExampleSchedule() {
	// Run f every hour.
	scheduling.Schedule(context.Background(), time.Hour, 0, func(ctx context.Context, t time.Time) {
		// Do your job here
	})

	// Run f once a day at noon.
	scheduling.Schedule(context.Background(), 24*time.Hour, 12*time.Hour, func(ctx context.Context, t time.Time) {
		// Do your job here
	})

	// Run f twice a day at 03:40 (00:00 + 03:40) and 15:40 (12:00 + 03:40).
	scheduling.Schedule(context.Background(), 12*time.Hour, 3*time.Hour+40*time.Minute, func(ctx context.Context, t time.Time) {
		// Do your job here
	})

	// Filter can be implemented inside f.
	scheduling.Schedule(context.Background(), time.Hour, 0, func(ctx context.Context, t time.Time) {
		if t.Weekday() == time.Wednesday {
			return
		}
		// Do your job here
	})

	// Schedule can be cancelled with a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	scheduling.Schedule(ctx, time.Hour, 0, func(ctx context.Context, t time.Time) {
		// Do your job here
	})
	cancel()
}
