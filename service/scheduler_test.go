package service

import (
	"context"
	"testing"

	"github.com/robfig/cron/v3"
	"github.com/weirwei/rss2email/constants"
)

func TestAddSchedule(t *testing.T) {
	ctx := context.Background()
	config := Config{
		FeedURL:      "https://example.com/feed",
		Subscription: constants.SubscriptionID("demo"),
	}

	t.Run("startup does not register", func(t *testing.T) {
		c := cron.New()
		addSchedule(ctx, c, config, constants.ScheduleTypeStartup, "")
		if got := len(c.Entries()); got != 0 {
			t.Fatalf("expected 0 entries, got %d", got)
		}
	})

	t.Run("live registers", func(t *testing.T) {
		c := cron.New()
		addSchedule(ctx, c, config, constants.ScheduleTypeLive, "")
		if got := len(c.Entries()); got != 1 {
			t.Fatalf("expected 1 entries, got %d", got)
		}
	})

	t.Run("cron with empty spec does not register", func(t *testing.T) {
		c := cron.New()
		addSchedule(ctx, c, config, constants.ScheduleTypeCron, "")
		if got := len(c.Entries()); got != 0 {
			t.Fatalf("expected 0 entries, got %d", got)
		}
	})

	t.Run("cron with invalid spec does not register", func(t *testing.T) {
		c := cron.New()
		addSchedule(ctx, c, config, constants.ScheduleTypeCron, "invalid")
		if got := len(c.Entries()); got != 0 {
			t.Fatalf("expected 0 entries, got %d", got)
		}
	})

	t.Run("cron with valid spec registers", func(t *testing.T) {
		c := cron.New()
		addSchedule(ctx, c, config, constants.ScheduleTypeCron, "0 */2 * * *")
		if got := len(c.Entries()); got != 1 {
			t.Fatalf("expected 1 entries, got %d", got)
		}
	})
}
