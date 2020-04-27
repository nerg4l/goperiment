package cron

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
)

type Executer interface {
	Exec(context.Context, time.Time)
}

type ExecuterFunc func(context.Context, time.Time)

func (f ExecuterFunc) Exec(ctx context.Context, now time.Time) {
	f(ctx, now)
}

const (
	firstMinute = 0
	lastMinute  = 59

	firstHour = 0
	lastHour  = 23

	firstDay = 1
	lastDay  = 31

	firstMonth = 1
	lastMonth  = 12

	firstWeekday = 0
	lastWeekday  = 7
)

// TODO remove when #format.shortDayNames becomes public
var shortDayNames = []string{
	"Sun",
	"Mon",
	"Tue",
	"Wed",
	"Thu",
	"Fri",
	"Sat",
}

// TODO remove when #format.shortMonthNames becomes public
var shortMonthNames = []string{
	"Jan",
	"Feb",
	"Mar",
	"Apr",
	"May",
	"Jun",
	"Jul",
	"Aug",
	"Sep",
	"Oct",
	"Nov",
	"Dec",
}

const (
	minuteStar cronFlag = 1 << iota
	hourStar
	dayStar
	weekdayStar
)

type cronFlag uint

// Cron runs specified function at scheduled time.
type Cron struct {
	minutes, hours, days, months, weekdays *big.Int

	flags cronFlag
}

// New parses a string to create an Cron.
//
// ┌───────────── minute (0 - 59)
// │ ┌───────────── hour (0 - 23)
// │ │ ┌───────────── day of the month (1 - 31)
// │ │ │ ┌───────────── month (1 - 12)
// │ │ │ │ ┌───────────── day of the week (0 - 7) (0 - Sunday to  6 - Saturday;
// │ │ │ │ │                                   7 is also Sunday for more predictability)
// │ │ │ │ │
// │ │ │ │ │
// * * * * * command to execute
//
// The following nonstandard predefined scheduling definitions can be used:
//
// * @yearly: Run once a year at midnight of 1 January
// * @annually: Run once a year at midnight of 1 January
// * @monthly: Run once a month at midnight of the first day of the month
// * @weekly: Run once a week at midnight on Sunday morning
// * @daily: Run once a day at midnight
// * @midnight: Run once a day at midnight
// * @hourly: Run once an hour at the beginning of the hour
func New(s string) (c *Cron, err error) {
	var ch rune
	reader := strings.NewReader(s)
	if ch, _, err = reader.ReadRune(); err != nil {
		return nil, errors.New("cron: cannot parse empty string as definition")
	}

	c = &Cron{}

	if ch == '@' {
		switch s {
		case "@yearly", "@annually": // 0 0 1 1 *
			c.minutes = big.NewInt(0)
			c.hours = big.NewInt(0)
			c.days = big.NewInt(0)
			c.months = big.NewInt(0)
			c.weekdays = newCronRangeBetween(firstWeekday, lastWeekday).Bits()
			c.flags = weekdayStar
		case "@monthly": // 0 0 1 * *
			c.minutes = big.NewInt(0)
			c.hours = big.NewInt(0)
			c.days = big.NewInt(0)
			c.months = newCronRangeBetween(firstMonth, lastMonth).Bits()
			c.weekdays = newCronRangeBetween(firstWeekday, lastWeekday).Bits()
			c.flags = weekdayStar
		case "@weekly": // 0 0 * * 0
			c.minutes = big.NewInt(0)
			c.hours = big.NewInt(0)
			c.days = newCronRangeBetween(firstDay, lastDay).Bits()
			c.months = newCronRangeBetween(firstMonth, lastMonth).Bits()
			c.weekdays = big.NewInt(0)
			c.flags = dayStar
		case "@daily", "@midnight": // 0 0 * * *
			c.minutes = big.NewInt(0)
			c.hours = big.NewInt(0)
			c.days = newCronRangeBetween(firstDay, lastDay).Bits()
			c.months = newCronRangeBetween(firstMonth, lastMonth).Bits()
			c.weekdays = newCronRangeBetween(firstWeekday, lastWeekday).Bits()
			c.flags = dayStar | weekdayStar
		case "@hourly": // 0 * * * *
			c.minutes = big.NewInt(0)
			c.hours = newCronRangeBetween(firstHour, lastHour).Bits()
			c.days = newCronRangeBetween(firstDay, lastDay).Bits()
			c.months = newCronRangeBetween(firstMonth, lastMonth).Bits()
			c.weekdays = newCronRangeBetween(firstWeekday, lastWeekday).Bits()
			c.flags = hourStar | dayStar | weekdayStar
		default:
			return nil, fmt.Errorf("cron: cannot parse %s as predefined scheduling definition", s)
		}
	} else {
		var b *big.Int
		cs := cronScanner{scanner: reader, ch: ch}
		if cs.ch == '*' {
			c.flags |= minuteStar
		}
		if b, err = cs.ScanList(firstMinute, lastMinute, nil); err != nil {
			return nil, fmt.Errorf("cron: cannot parse minute part of scheduling definition: %w", err)
		}
		c.minutes = b
		if cs.ch == '*' {
			c.flags |= hourStar
		}
		if b, err = cs.ScanList(firstHour, lastHour, nil); err != nil {
			return nil, fmt.Errorf("cron: cannot parse hour part of scheduling definition: %w", err)
		}
		c.hours = b
		if cs.ch == '*' {
			c.flags |= dayStar
		}
		if b, err = cs.ScanList(firstDay, lastDay, nil); err != nil {
			return nil, fmt.Errorf("cron: cannot parse day part of scheduling definition: %w", err)
		}
		c.days = b
		if b, err = cs.ScanList(firstMonth, lastMonth, shortMonthNames); err != nil {
			return nil, fmt.Errorf("cron: cannot parse month part of scheduling definition: %w", err)
		}
		c.months = b
		if cs.ch == '*' {
			c.flags |= weekdayStar
		}
		if b, err = cs.ScanList(firstWeekday, lastWeekday, shortDayNames); err != nil && err != io.EOF {
			return nil, fmt.Errorf("cron: cannot parse weekday part of scheduling definition: %w", err)
		}
		c.weekdays = b
	}

	if c.weekdays.Bit(0) == 1 || c.weekdays.Bit(7) == 1 {
		c.weekdays.SetBit(c.weekdays, 0, 1)
		c.weekdays.SetBit(c.weekdays, 7, 1)
	}

	return c, nil
}

func (c Cron) hasFlag(cf cronFlag) bool {
	return c.flags&cf > 0
}

func (c Cron) Check(t time.Time) bool {
	h, m, _ := t.Clock()
	if !c.hasFlag(minuteStar) && c.minutes.Bit(m-firstMinute) == 0 {
		return false
	}
	if !c.hasFlag(hourStar) && c.hours.Bit(h-firstHour) == 0 {
		return false
	}
	if c.months.Bit(t.Day()-firstMonth) == 0 {
		return false
	}
	// Commands are executed when the 'minute', 'hour',
	// and 'month of the year' fields match the current
	// time, and at least one of the two 'day' fields
	// ('day of month', or 'day of week') match the
	// current time
	if !c.hasFlag(dayStar) && !c.hasFlag(weekdayStar) &&
		c.days.Bit(t.Day()-firstDay) == 0 && c.weekdays.Bit(int(t.Weekday())-firstWeekday) == 0 {
		return false
	}

	return true
}

// cronRange contains details about a range
type cronRange struct {
	from, to, step int

	// when a range starts from a non-zero value
	// an offset is needed to be able to deduct
	// from the range
	offset int
}

// newCronRangeBetween creates a cronRange where low and high used
// for the boundaries of the range and step value is set to one.
//
// The bits of the returned value will be all 1.
func newCronRangeBetween(low, high int) cronRange {
	return cronRange{from: low, to: high, step: 1, offset: low}
}

// Bits outputs a range as a bitset
//
// Zero value for cr.step will cause panic
func (cr cronRange) Bits() *big.Int {
	b := big.NewInt(0)
	f := cr.from - cr.offset
	t := cr.to - cr.offset
	// set all elements from `f` to `t`, stepping by `step`
	s := cr.step
	if s == 0 {
		panic("step can not be zero")
	}
	for i := f; i <= t; i += s {
		b.SetBit(b, i, 1)
	}
	return b
}

type cronScanner struct {
	scanner io.RuneScanner

	ch rune
}

func (cs *cronScanner) readRune() (err error) {
	cs.ch, _, err = cs.scanner.ReadRune()
	return
}

func (cs *cronScanner) unreadRune() (err error) {
	return cs.scanner.UnreadRune()
}

// ScanList scans a list of cron ranges
// and moves the scanner to the next list.
//
// A list is a sequence of comma separated
// cron ranges.
func (cs *cronScanner) ScanList(low, high int, names []string) (b *big.Int, err error) {
	b = big.NewInt(0)
	for {
		if r, err := cs.scanRange(low, high, names); err == io.EOF {
			return b.Add(b, r.Bits()), err
		} else if err != nil {
			return b, err
		} else {
			b = b.Add(b, r.Bits())
		}
		if cs.ch == ',' {
			if err = cs.readRune(); err != nil {
				return b, err
			}
		} else {
			break
		}
	}

	// skip to some blanks, then skip over the blanks.
	for !unicode.IsSpace(cs.ch) {
		if err = cs.readRune(); err != nil {
			return b, err
		}
	}
	for unicode.IsSpace(cs.ch) {
		if err = cs.readRune(); err != nil {
			return b, err
		}
	}

	return b, err
}

// scanRange scans a cron range.
//
// A range is a `number [ "-" number ] [ "/" number ]`
func (cs *cronScanner) scanRange(low, high int, names []string) (r cronRange, err error) {
	r.step = 1
	r.offset = low

	if cs.ch == '*' {
		// '*' means "first-last" but can still be modified by /step
		r.from, r.to = low, high
		if err = cs.readRune(); err != nil {
			return r, err
		}
	} else {
		if r.from, err = cs.scanNumber(low, names, ",- \t\n"); err != nil {
			return r, err
		}

		if cs.ch != '-' {
			// not a range, it's a single number
			r.to = r.from
		} else {
			// eat the dash
			if err = cs.readRune(); err != nil {
				return r, fmt.Errorf("cron: dash must be folowed by a value: %w", err)
			}

			// get the number following the dash
			if r.to, err = cs.scanNumber(low, names, "/, \t\n"); err != nil {
				return r, err
			} else if r.from > r.to {
				return r, errors.New("cron: end of range must be greater than start of range")
			}
		}
	}

	// check for step size
	if cs.ch == '/' {
		// eat the slash
		if err = cs.readRune(); err != nil {
			return r, err
		}

		// get the step size -- note: we don't pass the
		// names here, because the number is not an
		// element id, it's a step size.  'low' is
		// sent as a 0 since there is no offset either.
		if r.step, err = cs.scanNumber(0, nil, ", \t\n"); err != nil {
			return r, err
		} else if r.step == 0 {
			return r, errors.New("cron: step must be greater than 0")
		}
	}

	if r.to < low {
		return r, fmt.Errorf("cron: range end must be gerater than %d", low)
	}
	if r.to > high {
		return r, fmt.Errorf("cron: range end must be lower than %d", high)
	}
	if r.step > high {
		return r, fmt.Errorf("cron: step must be lower than %d", high)
	}

	return r, err
}

func (cs *cronScanner) scanNumber(low int, names []string, terms string) (num int, err error) {
	var sb strings.Builder

	// first look for a number
	for unicode.IsDigit(cs.ch) {
		sb.WriteRune(cs.ch)
		if err = cs.readRune(); err == io.EOF {
			break
		} else if err != nil {
			return num, err
		}
	}
	if sb.Len() != 0 {
		// got a number, check for valid terminator
		if !strings.ContainsRune(terms, cs.ch) && err != io.EOF {
			_ = cs.unreadRune()
			return 0, fmt.Errorf("cron: invalid terminator `%v`", cs.ch)
		}
		i, _ := strconv.Atoi(sb.String())
		return i, nil
	}

	// no numbers, look for a string if we have any
	if len(names) != 0 {
		for unicode.IsLetter(cs.ch) {
			sb.WriteRune(cs.ch)
			if err = cs.readRune(); err == io.EOF {
				break
			} else if err != nil {
				return num, err
			}
		}
		if sb.Len() != 0 && (strings.ContainsRune(terms, cs.ch) || err == io.EOF) {
			for i, name := range names {
				if strings.EqualFold(name, sb.String()) {
					return i + low, nil
				}
			}
		}
	}

	_ = cs.unreadRune()
	return 0, io.EOF
}

type atomicBool int32

func (b *atomicBool) isSet() bool { return atomic.LoadInt32((*int32)(b)) != 0 }
func (b *atomicBool) setTrue()    { atomic.StoreInt32((*int32)(b), 1) }

// ErrCrontabClosed is returned by the Crontab's Run method after a call to Shutdown or Close.
var ErrCrontabClosed = errors.New("cron: Crontab closed")

// NewCrontab allocates and returns a new Crontab.
func NewCrontab() *Crontab { return new(Crontab) }

// DefaultCrontab is the default Crontab used by Run.
var DefaultCrontab = &defaultCrontab

var defaultCrontab Crontab

type Runner struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup

	inShutdown atomicBool
	mu         sync.Mutex

	ct *Crontab
}

func (r *Runner) Run() error {
	if r.inShutdown.isSet() {
		return ErrCrontabClosed
	}

	r.mu.Lock()
	var ctx context.Context
	ctx, r.cancel = context.WithCancel(context.Background())
	r.mu.Unlock()

	if r.ct == nil {
		r.ct = DefaultCrontab
	}

	for {
		select {
		case <-ctx.Done():
			break
		case now := <-time.After(time.Minute):
			r.ct.Do(ctx, now)
		}
	}
}

// Immediate
func (r *Runner) Close() error {
	r.inShutdown.setTrue()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		r.cancel()
	}
	return nil
}

// Graceful
func (r *Runner) Shutdown(ctx context.Context) error {
	r.inShutdown.setTrue()
	r.mu.Lock()
	defer r.mu.Unlock()
	done := make(chan bool)
	if r.cancel != nil {
		r.cancel()

		go func() {
			r.wg.Wait()
			close(done)
		}()
	} else {
		close(done)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func Run() error {
	r := Runner{}
	return r.Run()
}

type Crontab struct {
	mu      sync.Mutex
	entries []*Entry
}

// Schedule registers the function for the given pattern.
// If a pattern is incorrect, Schedule panics.
func (ct *Crontab) Schedule(s string, executer Executer) {
	c, err := New(s)
	if err != nil {
		panic(err.Error())
	}
	if executer == nil {
		panic("cron: nil executer")
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	e := Entry{cron: c, executer: executer}
	ct.entries = append(ct.entries, &e)
}

func Schedule(s string, e Executer) {
	DefaultCrontab.Schedule(s, e)
}

func (ct *Crontab) Do(ctx context.Context, now time.Time) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	for _, e := range ct.entries {
		e := e
		go func() {
			if e.Check(now) {
				e.Exec(ctx, now)
			}
		}()
	}
}

type Entry struct {
	cron     *Cron
	executer Executer
}

func NewEntry(cron *Cron, executer Executer) *Entry {
	return &Entry{cron: cron, executer: executer}
}

func (e *Entry) Check(now time.Time) bool {
	return e.cron.Check(now)
}

func (e *Entry) Exec(ctx context.Context, now time.Time) {
	e.executer.Exec(ctx, now)
}
