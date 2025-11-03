package service

import (
	"testing"

	"github.com/weirwei/rss2email/test"
)

func TestAIInsightDailyService(t *testing.T) {
	test.Init()
	ctx := test.NewCtx()
	err := AIInsightDailyService(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
