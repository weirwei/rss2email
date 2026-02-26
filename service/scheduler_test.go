package service

import (
	"context"
	"testing"

	"github.com/robfig/cron/v3"
	"github.com/weirwei/rss2email/constants"
	"github.com/weirwei/rss2email/models"
)

func TestAddSchedule(t *testing.T) {
	ctx := context.Background()
	config := Config{
		FeedURL:      "https://example.com/feed",
		Subscription: constants.SubscriptionID("demo"),
	}

	t.Run("startup does not register", func(t *testing.T) {
		c := cron.New()
		entryID, ok := addSchedule(ctx, c, config, constants.ScheduleTypeStartup, "")
		if ok {
			t.Fatalf("expected not registered, got registered with entry id %d", entryID)
		}
		if got := len(c.Entries()); got != 0 {
			t.Fatalf("expected 0 entries, got %d", got)
		}
	})

	t.Run("live registers", func(t *testing.T) {
		c := cron.New()
		entryID, ok := addSchedule(ctx, c, config, constants.ScheduleTypeLive, "")
		if !ok || entryID == 0 {
			t.Fatalf("expected live schedule registered, got ok=%v entryID=%d", ok, entryID)
		}
		if got := len(c.Entries()); got != 1 {
			t.Fatalf("expected 1 entries, got %d", got)
		}
	})

	t.Run("cron with empty spec does not register", func(t *testing.T) {
		c := cron.New()
		entryID, ok := addSchedule(ctx, c, config, constants.ScheduleTypeCron, "")
		if ok {
			t.Fatalf("expected not registered, got registered with entry id %d", entryID)
		}
		if got := len(c.Entries()); got != 0 {
			t.Fatalf("expected 0 entries, got %d", got)
		}
	})

	t.Run("cron with invalid spec does not register", func(t *testing.T) {
		c := cron.New()
		entryID, ok := addSchedule(ctx, c, config, constants.ScheduleTypeCron, "invalid")
		if ok {
			t.Fatalf("expected not registered, got registered with entry id %d", entryID)
		}
		if got := len(c.Entries()); got != 0 {
			t.Fatalf("expected 0 entries, got %d", got)
		}
	})

	t.Run("cron with valid spec registers", func(t *testing.T) {
		c := cron.New()
		entryID, ok := addSchedule(ctx, c, config, constants.ScheduleTypeCron, "0 */2 * * *")
		if !ok || entryID == 0 {
			t.Fatalf("expected cron schedule registered, got ok=%v entryID=%d", ok, entryID)
		}
		if got := len(c.Entries()); got != 1 {
			t.Fatalf("expected 1 entries, got %d", got)
		}
	})
}

func TestReconcileSchedules(t *testing.T) {
	ctx := context.Background()
	c := cron.New()
	state := newScheduleState()
	feedSources := []models.FeedSource{
		{
			ID:           1,
			Subscription: constants.SubscriptionID("demo"),
			FeedURL:      "https://example.com/feed",
			ContentField: constants.FeedContentFieldDescription,
			ScheduleType: constants.ScheduleTypeLive,
		},
	}
	reconcileSchedules(ctx, c, feedSources, state)
	if got := len(c.Entries()); got != 1 {
		t.Fatalf("expected 1 entry after initial reconcile, got %d", got)
	}

	// same source should not duplicate entry
	reconcileSchedules(ctx, c, feedSources, state)
	if got := len(c.Entries()); got != 1 {
		t.Fatalf("expected still 1 entry after no-op reconcile, got %d", got)
	}

	// change schedule type/spec should replace entry
	feedSources[0].ScheduleType = constants.ScheduleTypeCron
	feedSources[0].CronSpec = "0 */2 * * *"
	reconcileSchedules(ctx, c, feedSources, state)
	if got := len(c.Entries()); got != 1 {
		t.Fatalf("expected 1 entry after update reconcile, got %d", got)
	}

	// remove source should remove entry
	reconcileSchedules(ctx, c, []models.FeedSource{}, state)
	if got := len(c.Entries()); got != 0 {
		t.Fatalf("expected 0 entries after remove reconcile, got %d", got)
	}
}
