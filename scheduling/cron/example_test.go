package cron_test

import (
	"context"
	"fmt"
	"github.com/nerg4l/goperiment/scheduling/cron"
	"log"
	"sync"
	"time"
)

type countCmd struct {
	mu sync.Mutex // guards n
	n  int
}

func (h *countCmd) Exec(ctx context.Context, time time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.n++
	fmt.Printf("count is %d\n", h.n)
}

func ExampleCron() {
	cc := new(countCmd)
	cron.Schedule("* * * * *", cc)
	log.Fatal(cron.Run())
}
