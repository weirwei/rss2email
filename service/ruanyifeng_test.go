package service

import (
	"testing"

	"github.com/weirwei/rss2email/test"
)

func TestRuanyifengService(t *testing.T) {
	test.Init()
	ctx := test.NewCtx()
	err := RuanyifengService(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
