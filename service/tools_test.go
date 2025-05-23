package service

import (
	"testing"

	"github.com/weirwei/rss2email/test"
)

func TestSQLExec(t *testing.T) {
	test.Init()
	ctx := test.NewCtx()
	t.Run("query", func(t *testing.T) {
		err := SQLExec(ctx, "select * from user_subscriptions limit 10;")
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("exec", func(t *testing.T) {
		err := SQLExec(ctx, `update user_subscriptions set process = "https://decohack.com/producthunt-daily-2025-05-21/" where id = 1;`)
		if err != nil {
			t.Fatal(err)
		}
	})
}
