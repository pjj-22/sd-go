package caller

import (
	"bytes"
	"context"
	"net/http"
	"time"
)

// Requester sends one request on behalf of userID and returns the status code
type Requester func(ctx context.Context, userID string) (int, error)

// JSON returns a Requester that sends body to url with method
func JSON(method, url string, body []byte) Requester {
	return func(ctx context.Context, userID string) (int, error) {
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		if err != nil {
			return 0, err
		}
		req.Header.Set("X-User-ID", userID)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()
		return resp.StatusCode, nil
	}
}

type Result struct {
	UserID string
	Status int
	Err    error
	At     time.Time
}

type Caller struct {
	UserID   string
	Interval time.Duration
	Request  Requester
}

func New(userID string, interval time.Duration, r Requester) *Caller {
	return &Caller{UserID: userID, Interval: interval, Request: r}
}

// Run sends a request every Interval until ctx is cancelled.
func (c *Caller) Run(ctx context.Context, results chan<- Result) {
	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			status, err := c.Request(ctx, c.UserID)
			results <- Result{UserID: c.UserID, Status: status, Err: err, At: time.Now()}
		}
	}
}