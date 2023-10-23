package schedule_test

import (
	"context"
	"github.com/nerg4l/goperiment/schedule"
	"time"
)

func ExampleSchedule() {
	ctx, cancel := context.WithCancel(context.Background())

	// Every hour.
	hourly := schedule.NewSchedule(ctx, time.Hour, 0)

	// Once a day at noon.
	daily := schedule.NewSchedule(ctx, 24*time.Hour, 12*time.Hour)

	// Twice a day at 03:40 (00:00 + 03:40) and 15:40 (12:00 + 03:40).
	precise := schedule.NewSchedule(ctx, 12*time.Hour, 3*time.Hour+40*time.Minute)

	go func() {
		// Mimic cancellation
		<-time.After(time.Minute)
		cancel()
	}()

	for {
		select {
		case <-hourly.C:
			// Do your job here
		case t := <-daily.C:
			// Filters can be added for the time.
			if t.Weekday() == time.Wednesday {
				continue
			}
		// Do your job here
		case <-precise.C:
			// Do your job here
		case <-ctx.Done():
			// Graceful shutdown
			return
		}
	}
}
