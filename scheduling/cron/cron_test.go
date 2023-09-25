package cron

import (
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Parallel()

	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantC   *Cron
		wantErr bool
	}{
		{
			name:    `wrong predefined scheduling definition`,
			args:    args{s: `@any`},
			wantErr: true,
		},
		{
			name:    `too big value for minute`,
			args:    args{s: `60 * * * *`},
			wantErr: true,
		},
		{
			name:    `too big value for hour`,
			args:    args{s: `* 24 * * *`},
			wantErr: true,
		},
		{
			name:    `too small value for day of month`,
			args:    args{s: `* * 0 * *`},
			wantErr: true,
		},
		{
			name:    `too big value for day of month`,
			args:    args{s: `* * 32 * *`},
			wantErr: true,
		},
		{
			name:    `too small value for month`,
			args:    args{s: `* * * 0 *`},
			wantErr: true,
		},
		{
			name:    `too big value for month`,
			args:    args{s: `* * * 13 *`},
			wantErr: true,
		},
		{
			name:    `too big value for day of week`,
			args:    args{s: `* * * * 8`},
			wantErr: true,
		},
		{
			name:    `invalid range of minutes, end lower than start`,
			args:    args{s: `2-0 * * * *`},
			wantErr: true,
		},
		{
			name:    `invalid range of minutes, missing end`,
			args:    args{s: `2- * * * *`},
			wantErr: true,
		},
		{
			name:    `invalid range of minutes, missing start`,
			args:    args{s: `-2 * * * *`},
			wantErr: true,
		},
		{
			name:    `invalid range of weekdays, missing end`,
			args:    args{s: `* * * * 1-`},
			wantErr: true,
		},
		{
			name:    `invalid terminator for single value`,
			args:    args{s: `* 2/1 * * *`},
			wantErr: true,
		},
		{
			name:    `invalid terminator for range values`,
			args:    args{s: `* 2-3-4 * * *`},
			wantErr: true,
		},

		{
			name: `yearly`,
			args: args{s: `@yearly`},
			wantC: &Cron{
				minutes:  0,
				hours:    0,
				days:     0,
				months:   0,
				weekdays: 1<<8 - 1,
				flags:    weekdayStar,
			},
		},
		{
			name: `monthly`,
			args: args{s: `@monthly`},
			wantC: &Cron{
				minutes:  0,
				hours:    0,
				days:     0,
				months:   1<<12 - 1,
				weekdays: 1<<8 - 1,
				flags:    weekdayStar,
			},
		},
		{
			name: `weekly`,
			args: args{s: `@weekly`},
			wantC: &Cron{
				minutes:  0,
				hours:    0,
				days:     1<<31 - 1,
				months:   1<<12 - 1,
				weekdays: 0,
				flags:    dayStar,
			},
		},
		{
			name: `daily`,
			args: args{s: `@daily`},
			wantC: &Cron{
				minutes:  0,
				hours:    0,
				days:     1<<31 - 1,
				months:   1<<12 - 1,
				weekdays: 1<<8 - 1,
				flags:    dayStar | weekdayStar,
			},
		},
		{
			name: `hourly`,
			args: args{s: `@hourly`},
			wantC: &Cron{
				minutes:  0,
				hours:    1<<24 - 1,
				days:     1<<31 - 1,
				months:   1<<12 - 1,
				weekdays: 1<<8 - 1,
				flags:    dayStar | weekdayStar,
			},
		},
		{
			name: `every minute`,
			args: args{s: `* * * * *`},
			wantC: &Cron{
				minutes:  1<<60 - 1,
				hours:    1<<24 - 1,
				days:     1<<31 - 1,
				months:   1<<12 - 1,
				weekdays: 1<<8 - 1,
				flags:    dayStar | weekdayStar,
			},
		},
		{
			name: `minute with step`,
			args: args{s: `*/2 * * * *`},
			wantC: &Cron{
				minutes:  0b10101010101010101010101010101010101010101010101010101010101,
				hours:    1<<24 - 1,
				days:     1<<31 - 1,
				months:   1<<12 - 1,
				weekdays: 1<<8 - 1,
				flags:    dayStar | weekdayStar,
			},
		},
		{
			name: `range of minutes`,
			args: args{s: `0-2 * * * *`},
			wantC: &Cron{
				minutes:  0b111,
				hours:    1<<24 - 1,
				days:     1<<31 - 1,
				months:   1<<12 - 1,
				weekdays: 1<<8 - 1,
				flags:    dayStar | weekdayStar,
			},
		},
		{
			name: `range of minutes with step`,
			args: args{s: `0-2/2 * * * *`},
			wantC: &Cron{
				minutes:  0b101,
				hours:    1<<24 - 1,
				days:     1<<31 - 1,
				months:   1<<12 - 1,
				weekdays: 1<<8 - 1,
				flags:    dayStar | weekdayStar,
			},
		},
		{
			name: `minute list`,
			args: args{s: `0,2,5 * * * *`},
			wantC: &Cron{
				minutes:  0b100101,
				hours:    1<<24 - 1,
				days:     1<<31 - 1,
				months:   1<<12 - 1,
				weekdays: 1<<8 - 1,
				flags:    dayStar | weekdayStar,
			},
		},
		{
			name: `range of minutes list`,
			args: args{s: `0-2/2,5-6 * * * *`},
			wantC: &Cron{
				minutes:  0b1100101,
				hours:    1<<24 - 1,
				days:     1<<31 - 1,
				months:   1<<12 - 1,
				weekdays: 1<<8 - 1,
				flags:    dayStar | weekdayStar,
			},
		},
		{
			name: `sunday can be 7`,
			args: args{s: `* * * * 7`},
			wantC: &Cron{
				minutes:  1<<60 - 1,
				hours:    1<<24 - 1,
				days:     1<<31 - 1,
				months:   1<<12 - 1,
				weekdays: 0b10000001,
				flags:    dayStar,
			},
		},
		{
			name: `month as string`,
			args: args{s: `* * * JAN *`},
			wantC: &Cron{
				minutes:  1<<60 - 1,
				hours:    1<<24 - 1,
				days:     1<<31 - 1,
				months:   0b1,
				weekdays: 1<<8 - 1,
				flags:    dayStar | weekdayStar,
			},
		},
		{
			name: `range of months as string`,
			args: args{s: `* * * JAN-FEB *`},
			wantC: &Cron{
				minutes:  1<<60 - 1,
				hours:    1<<24 - 1,
				days:     1<<31 - 1,
				months:   0b11,
				weekdays: 1<<8 - 1,
				flags:    dayStar | weekdayStar,
			},
		},
		{
			name: `weekday as string`,
			args: args{s: `* * * * SUN`},
			wantC: &Cron{
				minutes:  1<<60 - 1,
				hours:    1<<24 - 1,
				days:     1<<31 - 1,
				months:   1<<12 - 1,
				weekdays: 0b10000001,
				flags:    dayStar,
			},
		},
		{
			name: `range of weekdays as string`,
			args: args{s: `* * * * SUN-TUE`},
			wantC: &Cron{
				minutes:  1<<60 - 1,
				hours:    1<<24 - 1,
				days:     1<<31 - 1,
				months:   1<<12 - 1,
				weekdays: 0b10000111,
				flags:    dayStar,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotC, err := Parse(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(gotC, tt.wantC) {
				t.Errorf(`
Parse() gotC = &{minutes:%x hours:%x days:%x months:%x weekdays:%x flags:%b},
        want &{minutes:%x hours:%x days:%x months:%x weekdays:%x flags:%b}`,
					gotC.minutes, gotC.hours, gotC.days, gotC.months, gotC.weekdays, gotC.flags,
					tt.wantC.minutes, tt.wantC.hours, tt.wantC.days, tt.wantC.months, tt.wantC.weekdays, tt.wantC.flags)
			}
		})
	}
}

func TestCron_ScheduledFor(t *testing.T) {
	t.Run("every minute", func(t *testing.T) {
		c, err := Parse("* * * * *")
		if err != nil {
			t.Error(err)
			return
		}
		var want bool
		want = true
		if got := c.ScheduledFor(time.Date(1990, 1, 1, 0, 0, 0, 0, time.Local)); got != want {
			t.Errorf("ScheduledFor() = %v, want %v", got, want)
		}
		want = true
		if got := c.ScheduledFor(time.Date(1990, 1, 1, 0, 1, 0, 0, time.Local)); got != want {
			t.Errorf("ScheduledFor() = %v, want %v", got, want)
		}
	})
	t.Run("minute with step", func(t *testing.T) {
		c, err := Parse("*/2 * * * *")
		if err != nil {
			t.Error(err)
			return
		}
		var want bool
		want = true
		if got := c.ScheduledFor(time.Date(1990, 1, 1, 0, 0, 0, 0, time.Local)); got != want {
			t.Errorf("ScheduledFor() = %v, want %v", got, want)
		}
		want = false
		if got := c.ScheduledFor(time.Date(1990, 1, 1, 0, 1, 0, 0, time.Local)); got != want {
			t.Errorf("ScheduledFor() = %v, want %v", got, want)
		}
		want = true
		if got := c.ScheduledFor(time.Date(1990, 1, 1, 0, 2, 0, 0, time.Local)); got != want {
			t.Errorf("ScheduledFor() = %v, want %v", got, want)
		}
	})
	t.Run("range of minutes", func(t *testing.T) {
		c, err := Parse("0-1 * * * *")
		if err != nil {
			t.Error(err)
			return
		}
		var want bool
		want = true
		if got := c.ScheduledFor(time.Date(1990, 1, 1, 0, 0, 0, 0, time.Local)); got != want {
			t.Errorf("ScheduledFor() = %v, want %v", got, want)
		}
		want = true
		if got := c.ScheduledFor(time.Date(1990, 1, 1, 0, 1, 0, 0, time.Local)); got != want {
			t.Errorf("ScheduledFor() = %v, want %v", got, want)
		}
		want = false
		if got := c.ScheduledFor(time.Date(1990, 1, 1, 0, 2, 0, 0, time.Local)); got != want {
			t.Errorf("ScheduledFor() = %v, want %v", got, want)
		}
	})
	t.Run("range of minutes with step", func(t *testing.T) {
		c, err := Parse("0,2 * * * *")
		if err != nil {
			t.Error(err)
			return
		}
		var want bool
		want = true
		if got := c.ScheduledFor(time.Date(1990, 1, 1, 0, 0, 0, 0, time.Local)); got != want {
			t.Errorf("ScheduledFor() = %v, want %v", got, want)
		}
		want = false
		if got := c.ScheduledFor(time.Date(1990, 1, 1, 0, 1, 0, 0, time.Local)); got != want {
			t.Errorf("ScheduledFor() = %v, want %v", got, want)
		}
		want = true
		if got := c.ScheduledFor(time.Date(1990, 1, 1, 0, 2, 0, 0, time.Local)); got != want {
			t.Errorf("ScheduledFor() = %v, want %v", got, want)
		}
	})
}
