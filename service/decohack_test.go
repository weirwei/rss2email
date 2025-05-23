package service

import (
	"testing"

	"github.com/weirwei/rss2email/test"
)

func TestDecoHackService(t *testing.T) {
	test.Init()
	ctx := test.NewCtx()
	err := DecoHackService(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
