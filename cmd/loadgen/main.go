package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/pjj-22/sd-go/internal/caller"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	req := caller.JSON(http.MethodPost, "http://localhost:8080/hello", nil)

	callers := []*caller.Caller{
		caller.New("alice", 500*time.Millisecond, req),
		caller.New("bob", 1250*time.Millisecond, req),
		caller.New("carlton", 5000*time.Millisecond, req),
	}

	results := make(chan caller.Result)
	var wg sync.WaitGroup

	for _, caller := range callers {
		wg.Add(1)
		go func(c *caller.Caller) {
			defer wg.Done()
			c.Run(ctx, results)
		} (caller)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	type tally struct {
		ok, limited, failed int
	}

	stats := map[string]*tally{} 

	/**
	type Result struct {
		UserID string
		Status int
		Err    error
		At     time.Time
	*/

	for result := range results {
		t, ok := stats[result.UserID]
		if !ok {
			t = &tally{}
			stats[result.UserID] = t
		}

		switch {
		case result.Err != nil:
			t.failed++
		case result.Status == http.StatusTooManyRequests:
			t.limited++
		default:
			t.ok++
		}
	}

	fmt.Println()
	for _, c := range callers {
		t, ok := stats[c.UserID]
		if !ok {
			t = &tally{}
		}
		fmt.Printf("%-8s ok=%-5d limited=%-5d failed=%d\n", c.UserID, t.ok, t.limited, t.failed)
	}
}