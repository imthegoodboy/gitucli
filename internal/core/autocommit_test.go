package core

import (
	"testing"
	"time"
)

func TestNextClockDelay(t *testing.T) {
	now := time.Date(2026, 5, 7, 10, 30, 0, 0, time.Local)

	delay, err := NextClockDelay(now, "10:45")
	if err != nil {
		t.Fatal(err)
	}
	if delay != 15*time.Minute {
		t.Fatalf("unexpected same-day delay: %s", delay)
	}

	delay, err = NextClockDelay(now, "10:15")
	if err != nil {
		t.Fatal(err)
	}
	if delay != 23*time.Hour+45*time.Minute {
		t.Fatalf("unexpected next-day delay: %s", delay)
	}
}

