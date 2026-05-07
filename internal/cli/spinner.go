package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/parth/gitucli/internal/textui"
)

func waitWithSpinner(ctx context.Context, w io.Writer, delay time.Duration, label string) error {
	if delay <= 0 {
		return nil
	}

	frames := []string{"|", "/", "-", "\\"}
	deadline := time.NewTimer(delay)
	ticker := time.NewTicker(120 * time.Millisecond)
	defer deadline.Stop()
	defer ticker.Stop()

	i := 0
	for {
		select {
		case <-ctx.Done():
			clearLine(w)
			return ctx.Err()
		case <-deadline.C:
			clearLine(w)
			fmt.Fprintf(w, "%s schedule reached\n", textui.Success("[OK]"))
			return nil
		case <-ticker.C:
			fmt.Fprintf(w, "\r%s %s %s", textui.Muted("[WAIT]"), frames[i%len(frames)], label)
			i++
		}
	}
}

func clearLine(w io.Writer) {
	fmt.Fprint(w, "\r                                                                                                                        \r")
}
